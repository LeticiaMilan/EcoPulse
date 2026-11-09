package main

import (
	"context"
	_ "embed"
	"log"
	"net/http"
	http2 "report-service/internal/infraestructure/http"

	"os"
	"os/signal"
	"report-service/internal/application"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/entrypoint/delivery"
	"report-service/internal/infraestructure/entrypoint/worker"
	"report-service/internal/infraestructure/external/BrasilNtp"
	"report-service/internal/infraestructure/messaging/amqp"
	"report-service/internal/infraestructure/persistence/postgres/repository"
	"report-service/internal/ports"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	pool := database()
	defer pool.Close()
	log.Println("Database connection established")

	// Escutar interrupções do S.O
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	schema, err := os.ReadFile("internal/infraestructure/persistence/postgres/schema.sql")
	_, err = pool.Exec(ctx, string(schema))

	log.Printf("%s", string(schema))

	if err != nil {
		// IMPORTANTE: Isso vai te mostrar exatamente qual linha do DDL falhou!
		log.Fatalf("Erro CRÍTICO ao aplicar o schema SQL: %v", err)
	}
	InboxRepository := repository.NewInboxRepository(pool)
	OutboxRepository := repository.NewOutboxRepository(pool)
	ReportRepository := repository.NewReportRepository(pool)
	AreaRepository := repository.NewAreaRepository(pool)

	dispatcherChan := make(chan worker.EventEnvelope, 100)
	ConsumerChannel := StartRabbit(ctx, InboxRepository, OutboxRepository, dispatcherChan)
	defer ConsumerChannel.Close()

	BrasilNtpClock, err := BrasilNtp.NewBrasilNtpClock()
	if err != nil {
		log.Fatal(err)
	}

	reportUseCase := application.NewReportUseCase(ReportRepository, BrasilNtpClock, OutboxRepository)
	AreaUseCase := application.NewAreaUseCase(AreaRepository, ReportRepository, OutboxRepository)
	ReportGeneratedHandler := delivery.NewReportHandler(application.NewGeneratedReportUseCase(reportUseCase))
	ReportValidatedHandler := delivery.NewReportHandler(application.NewValidatedReport(reportUseCase))
	ReportDefinedAreaHandler := delivery.NewAreaHandler(application.NewAreaDefined(AreaUseCase))
	Inboxprocessor := delivery.NewInboxProcessor()
	Inboxprocessor.RegisterHandler(string(domain.InboxGeneratedReportEvent), ReportGeneratedHandler)
	Inboxprocessor.RegisterHandler(string(domain.InboxValidatedReportEvent), ReportValidatedHandler)
	for _, areaEvent := range domain.InboxAreaEvents {
		Inboxprocessor.RegisterHandler(string(areaEvent), ReportDefinedAreaHandler)
	}

	OutboxWorker := worker.NewOutboxWorker(OutboxRepository, dispatcherChan)
	inboxWorker := worker.NewInboxWorker(InboxRepository, Inboxprocessor, OutboxRepository)
	go inboxWorker.Process(ctx)
	go OutboxWorker.Process(ctx)

	HttpHandler := http2.NewHttpReportHandler(reportUseCase)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /report/api/sync", HttpHandler.HandlerReportSync)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}

	<-ctx.Done()
}

func database() *pgxpool.Pool {
	ctx := context.Background()
	dsn := os.Getenv("POSTGRES_URL")

	if dsn == "" {
		log.Fatal("POSTGRES_URL environment variable is required")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal("Failed to parse PostgreSQL config: ", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2 // Numero minimo de conexões que o golang mantém ativa - diminuir latência nas primeiras leituras
	cfg.MaxConnLifetime = time.Hour
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	return pool
}

func StartRabbit(ctx context.Context, inboxRepository ports.InboxRepository, outboxRepository ports.OutboxRepository, dispatcherChan chan worker.EventEnvelope) *amqp091.Channel {
	con := os.Getenv("AMQP_URL")
	if con == "" {
		log.Fatal("AMQP_URL environment variable is required")
	}
	QueueManager, err := amqp.NewEventManager(con)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("RabbitMQ connection established")

	var RoutingKeys []string
	RoutingKeys = append(RoutingKeys, string(domain.InboxGeneratedReportEvent))
	RoutingKeys = append(RoutingKeys, string(domain.InboxValidatedReportEvent))
	for _, areaEvent := range domain.InboxAreaEvents {
		RoutingKeys = append(RoutingKeys, string(areaEvent))
	}

	consumer := amqp.NewConsumer(QueueManager, inboxRepository, "exchange.ecopulse", "topic", "queue.service.report", RoutingKeys)
	ConsumerChannel, ReceivedMessages, err := consumer.Start()

	dispatcher := amqp.NewDispatcher(QueueManager, outboxRepository, "exchange.ecopulse", "topic")
	err = dispatcher.Start(ctx, dispatcherChan)
	if err != nil {
		log.Fatal(err)
	}
	if ConsumerChannel == nil || ReceivedMessages == nil || err != nil {
		log.Fatal("Failed to start consumer channel")
	}

	go consumer.Consumir(ctx, ReceivedMessages)
	go consumer.Consumir(ctx, ReceivedMessages)

	return ConsumerChannel
}
