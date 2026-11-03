package main

import (
	"context"
	"log"
	"validacao-relato/src/database"
	"validacao-relato/src/repository"
	"validacao-relato/src/service"
	queue "validacao-relato/src/worker"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()

	// Conexão com o PostgreSQL usando a nova função centralizada
	// dsn := os.Getenv("DB_URL")
	dsn := "postgres://postgres:dbadmin@localhost:5432/aps_validacao?sslmode=disable"
	db, err := database.NewPostgresConnection(dsn)
	if err != nil {
		log.Fatalf("Falha crítica ao iniciar banco de dados: %v", err)
	}
	defer db.Close()

	// Conexão e Configuração do RabbitMQ
	// rabbitConn, err := queue.NewRabbitMQConn(os.Getenv("RABBITMQ_URL"))
	rabbitConn, err := queue.NewRabbitMQConn("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Erro ao conectar no RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()

	// Inicializar Repositórios e Serviços
	repo := repository.NewPostgresRelatoRepository(db)
	svc := service.NewRelatoService(db, repo)

	// Inicializar Workers e Consumidores
	outboxWorker := queue.NewOutboxWorker(db, rabbitConn.Channel)
	go outboxWorker.Start(ctx)

	inboxConsumer := queue.NewInboxConsumer(rabbitConn.Channel, repo, svc)
	if err := inboxConsumer.StartConsuming(ctx); err != nil {
		log.Fatalf("Erro ao iniciar consumidor Inbox: %v", err)
	}

	avaliacaoWorker := queue.NewAvaliacaoWorker(repo)
	go avaliacaoWorker.Start(ctx)

	log.Println("Microsserviço iniciado com sucesso! Todos os workers estão em execução.")

	// Mantém a aplicação rodando em background
	select {}
}
