package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/LeticiaMilan/ecopulse-login-mod/events"
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
func getUserID(r *http.Request) (string, error) {
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")

	id, err := uuid.Parse(idStr)
	if err != nil {
		return "", err
	}

	return id.String(), nil
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

	// Valida o tamanho da senha (limite do bcrypt)
	if len([]byte(request.Password)) > 72 {
		http.Error(w, "A senha deve ter no máximo 72 bytes", http.StatusBadRequest)
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

	// Inicia uma transação no banco de dados
	tx, err := userHandler.DB.BeginTx(r.Context(), nil)

	if err != nil {
		http.Error(w, "Erro ao iniciar transação", http.StatusInternalServerError)
		return
	}

	defer tx.Rollback()

	// Cadastra o usuário dentro da transação
	var response models.UserResponse

	err = tx.QueryRowContext(
		r.Context(),
		`INSERT INTO users (username, password)
		 VALUES ($1, $2)
		 RETURNING id, username`,
		request.Username,
		string(hashedPassword),
	).Scan(&response.ID, &response.Username)

	if err != nil {
		http.Error(w, "Erro ao cadastrar usuário. Verifique se o username já existe.", http.StatusConflict)
		return
	}

	// Registra o evento UserRegistered no Outbox
	err = events.SaveOutboxEvent(
		tx,
		"UserRegistered",
		"user.registered",
		map[string]interface{}{
			"user_id":  response.ID,
			"username": response.Username,
		},
	)

	if err != nil {
		http.Error(w, "Erro ao registrar evento", http.StatusInternalServerError)
		return
	}

	// Confirma a transação (cadastro + evento)
	err = tx.Commit()

	if err != nil {
		http.Error(w, "Erro ao confirmar cadastro", http.StatusInternalServerError)
		return
	}

	// Retorna somente informações públicas
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)
}

// PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := getUserID(r)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var request models.UpdateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	request.Username = strings.TrimSpace(request.Username)

	if request.Username == "" {
		http.Error(w, "Username é obrigatório", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Erro ao iniciar transação", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var user models.UserResponse
	passwordChanged := request.Password != nil

	if passwordChanged {
		if *request.Password == "" {
			http.Error(w, "Password não pode ser vazio", http.StatusBadRequest)
			return
		}

		if len([]byte(*request.Password)) > 72 {
			http.Error(w, "Password deve ter no máximo 72 bytes", http.StatusBadRequest)
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

		err = tx.QueryRowContext(
			r.Context(),
			`UPDATE users
			 SET username = $1, password = $2
			 WHERE id = $3
			 RETURNING id, username`,
			request.Username,
			string(hashedPassword),
			id,
		).Scan(&user.ID, &user.Username)

	} else {
		err = tx.QueryRowContext(
			r.Context(),
			`UPDATE users
			 SET username = $1
			 WHERE id = $2
			 RETURNING id, username`,
			request.Username,
			id,
		).Scan(&user.ID, &user.Username)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Usuário não encontrado", http.StatusNotFound)
			return
		}

		http.Error(w, "Erro ao atualizar usuário", http.StatusConflict)
		return
	}

	// Registra o evento de atualização
	err = events.SaveOutboxEvent(
		tx,
		"UserUpdated",
		"user.updated",
		map[string]interface{}{
			"user_id":          user.ID,
			"username":         user.Username,
			"password_changed": passwordChanged,
		},
	)

	if err != nil {
		http.Error(w, "Erro ao registrar evento", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Erro ao confirmar atualização", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// DELETE /users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := getUserID(r)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Erro ao iniciar transação", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var username string

	// Exclui o usuário e recupera seu username
	err = tx.QueryRowContext(
		r.Context(),
		`DELETE FROM users
		 WHERE id = $1
		 RETURNING username`,
		id,
	).Scan(&username)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Usuário não encontrado", http.StatusNotFound)
			return
		}

		http.Error(w, "Erro ao excluir usuário", http.StatusInternalServerError)
		return
	}

	// Registra o evento de exclusão
	err = events.SaveOutboxEvent(
		tx,
		"UserDeleted",
		"user.deleted",
		map[string]interface{}{
			"user_id":  id,
			"username": username,
		},
	)

	if err != nil {
		http.Error(w, "Erro ao registrar evento", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Erro ao confirmar exclusão", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
