package main

import (
	"log"
	"net/http"
	"os"

	"github.com/LeticiaMilan/ecopulse-login-mod/config"
	"github.com/LeticiaMilan/ecopulse-login-mod/handlers"
	"github.com/LeticiaMilan/ecopulse-login-mod/middleware"
	"github.com/LeticiaMilan/ecopulse-login-mod/models"
	"github.com/gorilla/mux"
)

func main() {
	dbConnection := config.SetupDB()

	_, err := dbConnection.Exec(models.CreateTableSQL)
	if err != nil {
		log.Fatalf("Erro ao criar a tabela de usuários: %v", err)
	}

	defer dbConnection.Close()

	// Obtém o segredo JWT das variáveis de ambiente
	jwtSecret := os.Getenv("JWT_SECRET")

	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET deve possuir pelo menos 32 caracteres")
	}

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

	// Rota protegida por JWT
	router.Handle(
		"/auth/me",
		middleware.AuthMiddleware([]byte(jwtSecret))(
			http.HandlerFunc(authHandler.Me),
		),
	).Methods("GET")

	log.Println("Login service iniciado na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
