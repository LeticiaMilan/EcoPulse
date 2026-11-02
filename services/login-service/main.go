package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/LeticiaMilan/ecopulse-login-mod/config"
	"github.com/LeticiaMilan/ecopulse-login-mod/events"
	"github.com/LeticiaMilan/ecopulse-login-mod/handlers"
	"github.com/LeticiaMilan/ecopulse-login-mod/middleware"
	"github.com/LeticiaMilan/ecopulse-login-mod/models"
	"github.com/gorilla/mux"
)

func main() {
	dbConnection := config.SetupDB()
	defer dbConnection.Close()

	// Cria a tabela de usuários
	_, err := dbConnection.Exec(models.CreateTableSQL)
	if err != nil {
		log.Fatalf("Erro ao criar tabela de usuários: %v", err)
	}

	// Cria a tabela de eventos pendentes
	_, err = dbConnection.Exec(events.CreateOutboxTableSQL)
	if err != nil {
		log.Fatalf("Erro ao criar tabela Outbox: %v", err)
	}

	// Configura o JWT
	jwtSecret := os.Getenv("JWT_SECRET")

	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET deve possuir pelo menos 32 caracteres")
	}

	// Configura o RabbitMQ
	rabbitURL := os.Getenv("RABBITMQ_URL")

	if rabbitURL == "" {
		log.Fatal("RABBITMQ_URL não foi configurada")
	}

	// Inicia o publicador de eventos em segundo plano
	publisher := events.NewPublisher(dbConnection, rabbitURL)
	go publisher.Run(context.Background())

	router := mux.NewRouter()

	userHandler := handlers.NewUserHandler(dbConnection)
	authHandler := handlers.NewAuthHandler(dbConnection, []byte(jwtSecret))

	// Rotas de usuários
	router.HandleFunc("/users", userHandler.ReadUsers).Methods("GET")
	router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", userHandler.ReadUserByID).Methods("GET")
	router.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// Rotas de autenticação
	router.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	router.Handle(
		"/auth/me",
		middleware.AuthMiddleware([]byte(jwtSecret))(
			http.HandlerFunc(authHandler.Me),
		),
	).Methods("GET")

	log.Println("Login service iniciado na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
