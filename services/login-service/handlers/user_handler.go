package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/LeticiaMilan/ecopulse-login-mod/models"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	DB *sql.DB
}

// construtor
func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{DB: db}
}

// Extrai e valida o ID da URL
func getUserID(r *http.Request) (int, error) {
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, err
	}

	return id, nil
}

// GET /users
func (userHandler *UserHandler) ReadUsers(w http.ResponseWriter, r *http.Request) {

	rows, err := userHandler.DB.Query(
		"SELECT id, username FROM users",
	)

	if err != nil {
		http.Error(w, "Erro ao buscar usuários", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var users []models.UserResponse

	for rows.Next() {
		var user models.UserResponse

		err := rows.Scan(&user.ID, &user.Username)

		if err != nil {
			http.Error(w, "Erro ao ler usuários", http.StatusInternalServerError)
			return
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Erro ao percorrer usuários", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GET /users/{id}
func (userHandler *UserHandler) ReadUserByID(w http.ResponseWriter, r *http.Request) {

	id, err := getUserID(r)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var user models.UserResponse

	err = userHandler.DB.QueryRow(
		"SELECT id, username FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Username)

	if err == sql.ErrNoRows {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Erro ao buscar usuário", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// POST /users
func (userHandler *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {

	var request models.CreateUserRequest

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Validação básica
	request.Username = strings.TrimSpace(request.Username)

	if request.Username == "" || request.Password == "" {
		http.Error(w, "Username e senha são obrigatórios", http.StatusBadRequest)
		return
	}

	// Gera o hash da senha utilizando bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		http.Error(w, "Erro ao processar senha", http.StatusInternalServerError)
		return
	}

	var userID int

	err = userHandler.DB.QueryRow(
		`INSERT INTO users (username, password)
		 VALUES ($1, $2)
		 RETURNING id`,
		request.Username,
		string(hashedPassword),
	).Scan(&userID)

	if err != nil {
		http.Error(w, "Erro ao cadastrar usuário. Verifique se o username já existe.", http.StatusConflict)
		return
	}

	// Retorna somente informações públicas
	response := models.UserResponse{
		ID:       userID,
		Username: request.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)
}

// PUT /users/{id}
func (userHandler *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {

	id, err := getUserID(r)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var request models.UpdateUserRequest

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	request.Username = strings.TrimSpace(request.Username)

	if request.Username == "" {
		http.Error(w, "Username é obrigatório", http.StatusBadRequest)
		return
	}

	// Verifica se o usuário existe
	var existingID int

	err = userHandler.DB.QueryRow(
		"SELECT id FROM users WHERE id = $1",
		id,
	).Scan(&existingID)

	if err == sql.ErrNoRows {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Erro ao buscar usuário", http.StatusInternalServerError)
		return
	}

	// Se uma nova senha foi informada, atualiza também o hash
	if request.Password != nil {

		if *request.Password == "" {
			http.Error(w, "A senha não pode ser vazia", http.StatusBadRequest)
			return
		}

		if len([]byte(*request.Password)) > 72 {
			http.Error(w, "A senha deve ter no máximo 72 bytes", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(*request.Password),
			bcrypt.DefaultCost,
		)

		if err != nil {
			http.Error(w, "Erro ao processar senha", http.StatusInternalServerError)
			return
		}

		_, err = userHandler.DB.Exec(
			`UPDATE users
			 SET username = $1, password = $2
			 WHERE id = $3`,
			request.Username,
			string(hashedPassword),
			id,
		)

		if err != nil {
			http.Error(w, "Erro ao atualizar usuário", http.StatusConflict)
			return
		}

	} else {

		// Se nenhuma senha foi informada, mantém o hash atual
		_, err = userHandler.DB.Exec(
			`UPDATE users
			 SET username = $1
			 WHERE id = $2`,
			request.Username,
			id,
		)

		if err != nil {
			http.Error(w, "Erro ao atualizar usuário", http.StatusConflict)
			return
		}
	}

	response := models.UserResponse{
		ID:       id,
		Username: request.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DELETE /users/{id}
func (userHandler *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {

	id, err := getUserID(r)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	result, err := userHandler.DB.Exec(
		"DELETE FROM users WHERE id = $1",
		id,
	)

	if err != nil {
		http.Error(w, "Erro ao excluir usuário", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Erro ao verificar exclusão", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
