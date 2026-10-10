package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"report-service/internal/application"
	"report-service/internal/infraestructure/entrypoint/delivery"
	"report-service/internal/infraestructure/entrypoint/worker"
	"report-service/internal/infraestructure/external/BrasilNtp"
	"report-service/internal/infraestructure/external/LatLngWorkGeocoding"
	"report-service/internal/infraestructure/messaging/amqp"
	"report-service/internal/infraestructure/persistence/postgres/repository"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	pool := database()
	defer pool.Close()
	log.Println("Database connection established")

	con := os.Getenv("AMQP_URL")
	if con == "" {
		log.Fatal("AMQP_URL environment variable is required")
	}
	QueueManager, err := amqp.NewEventManager(con)

	if err != nil {
		log.Fatal(err)
	}

	log.Print("RabbitMQ connection established")
	InboxRepository := repository.NewInboxRepository(pool)

	consumer1 := amqp.Consumer{
		InboxRepository: InboxRepository,
		Manager:         QueueManager,
		Fila:            "teste",
	}

	go consumer1.Start()

	var httpClient *http.Client
	httpClient = &http.Client{
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

	ReportRepository := repository.NewReportRepository(pool)
	GeoCodingAPI := LatLngWorkGeocoding.NewLatLngWorkGeocodingApi(httpClient)
	BrasilNtpClock, err := BrasilNtp.NewBrasilNtpClock()
	if err != nil {
		log.Fatal(err)
	}
	reportUseCase1 := application.NewProcessReportUseCase(ReportRepository, GeoCodingAPI, BrasilNtpClock)
	ReportGeneratedHandler := delivery.NewReportGeneratedHandler(reportUseCase1)
	processor1 := delivery.NewInboxProcessor()
	processor1.RegisterHandler("EVENTS.GENERATED.REPORT", ReportGeneratedHandler)

	// Escutar interrupções do S.O
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	inboxWorker1 := worker.NewInboxWorker(InboxRepository, processor1)
	cont := context.Background()
	go inboxWorker1.Process(&cont)

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
