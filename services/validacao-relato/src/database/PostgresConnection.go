package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// NewPostgresConnection abre e testa a conexão com o PostgreSQL
func NewPostgresConnection(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("erro ao configurar a conexão com o banco: %w", err)
	}

	// Testa se a conexão está realmente ativa e respondendo
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao testar a conexão com o banco (ping): %w", err)
	}

	log.Println("Conexão com o PostgreSQL estabelecida com sucesso!")
	return db, nil
}
