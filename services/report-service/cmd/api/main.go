package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"report-service/internal/application"
	"report-service/internal/application/inbox"
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

	ReportGeneratedHandler := application.ReportGeneratedHandler{}
	processor1 := inbox.NewInboxProcessor()
	processor1.RegisterHandler("events.report.generated", &ReportGeneratedHandler)

	// Escutar interrupções do S.O
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	inboxWorker1 := inbox.NewInboxWorker(InboxRepository, processor1)
	cont := context.Background()
	go inboxWorker1.Process(cont)

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
