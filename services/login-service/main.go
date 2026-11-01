package main

import (
	"log"
	"net/http"

	"github.com/LeticiaMilan/ecopulse-login-mod/config"
	"github.com/LeticiaMilan/ecopulse-login-mod/handlers"
	"github.com/LeticiaMilan/ecopulse-login-mod/models"
	"github.com/gorilla/mux"
)

// função chamada quando a aplicação é iniciada
func main() {
	dbConnection := config.SetupDB()

	_, err := dbConnection.Exec(models.CreateTableSQL)
	if err != nil {
		log.Fatalf("Erro ao criar a tabela de usuários: %v", err)
	}

	// fecha a conexão com o banco de dados quando a função main terminar
	defer dbConnection.Close()

	router := mux.NewRouter()

	userHandler := handlers.NewUserHandler(dbConnection)

	router.HandleFunc("/users", userHandler.ReadUsers).Methods("GET")
	router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}
