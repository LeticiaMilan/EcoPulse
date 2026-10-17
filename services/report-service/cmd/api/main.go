package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"report-service/internal/application"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/entrypoint/delivery"
	"report-service/internal/infraestructure/entrypoint/worker"
	"report-service/internal/infraestructure/external/BrasilNtp"
	"report-service/internal/infraestructure/external/LatLngWorkGeocoding"
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

	InboxRepository := repository.NewInboxRepository(pool)
	OutboxRepository := repository.NewOutboxRepository(pool)
	ReportRepository := repository.NewReportRepository(pool)

	// Escutar interrupções do S.O
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	dispatcherChan := make(chan worker.EventEnvelope, 100)
	ConsumerChannel := StartRabbit(ctx, InboxRepository, OutboxRepository, dispatcherChan)
	defer ConsumerChannel.Close()

	var httpClient *http.Client
	httpClient = ConfigHttpClient()
	GeoCodingAPI := LatLngWorkGeocoding.NewLatLngWorkGeocodingApi(httpClient)
	BrasilNtpClock, err := BrasilNtp.NewBrasilNtpClock()
	if err != nil {
		log.Fatal(err)
	}

	reportUseCase := application.NewProcessReportUseCase(ReportRepository, GeoCodingAPI, BrasilNtpClock, OutboxRepository)
	ReportGeneratedHandler := delivery.NewReportGeneratedHandler(reportUseCase)

	Inboxprocessor := delivery.NewInboxProcessor()
	Inboxprocessor.RegisterHandler(string(domain.InboxGeneratedReportEvent), ReportGeneratedHandler)

	OutboxWorker := worker.NewOutboxWorker(OutboxRepository, dispatcherChan)
	inboxWorker := worker.NewInboxWorker(InboxRepository, Inboxprocessor, OutboxRepository)
	go inboxWorker.Process(ctx)
	go OutboxWorker.Process(ctx)

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

func ConfigHttpClient() *http.Client {
	return &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,              // Total de conexões inativas mantidas no pool
			MaxIdleConnsPerHost: 3,                // Conexões inativas por host (padrão do Go é apenas 2)
			IdleConnTimeout:     90 * time.Second, // Tempo que uma conexão pode ficar inativa antes de fechar
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,  // Timeout para abrir a conexão TCP
				KeepAlive: 30 * time.Second, // Frequência do Keep-Alive no nível do SO
			}).DialContext,
		},
	}
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
