package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// funções públicas começam com letra maiúscula,
// funções privadas começam com letra minúscula

// pega as variáveis de ambiente e inicializa a conexão com o banco de dados
func SetupDB() *sql.DB {
	// Carrega o .env quando disponível.
	// No Docker, as variáveis serão fornecidas pelo docker-compose.
	err := godotenv.Load()
	if err != nil {
		log.Println("Arquivo .env não encontrado. Utilizando variáveis de ambiente.")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Verifica se todas as variáveis necessárias foram configuradas
	if dbHost == "" || dbPort == "" || dbUser == "" ||
		dbPassword == "" || dbName == "" {
		log.Fatal("Variáveis de ambiente do PostgreSQL não configuradas corretamente")
	}

	connectionStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	dbConnection, err := sql.Open("postgres", connectionStr)
	if err != nil {
		log.Fatal(err)
	}

	err = dbConnection.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Conexão com o banco de dados estabelecida com sucesso!")

	return dbConnection
}
