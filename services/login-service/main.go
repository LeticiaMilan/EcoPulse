package main

import (
	"log"
	"net/http"

	"github.com/LeticiaMilan/ecopulse-login-mod/config"
)

// função chamada quando a aplicação é iniciada
func main() {
	dbConnection := config.SetupDB()

	// fecha a conexão com o banco de dados quando a função main terminar
	defer dbConnection.Close()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
