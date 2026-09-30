package main

import (
	"context"
	"log"
	"os"
	"report-service/internal/infraestructure/messaging/amqp"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ctx := context.Background()
	dsn := os.Getenv("POSTGRES_URL")

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal(err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2 // Numero minimo de conexões que o golang mantém ativa - diminuir latência nas primeiras leituras
	cfg.MaxConnLifetime = time.Hour
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	log.Print(os.Getenv("AMQP_URL"))
	QueueManager := amqp.QueueManager{}
	err = QueueManager.NewEventManager((os.Getenv("AMQP_URL")))

	if err != nil {
		log.Fatal(err)
	}

	consumer1 := amqp.Consumer{}
	consumer1.NewConsumer(&QueueManager, "teste")

	consumer1.Start()
}
