package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/LeticiaMilan/ecopulse-login-mod/middleware"
	"github.com/LeticiaMilan/ecopulse-login-mod/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB        *sql.DB
	JWTSecret []byte
}

func NewAuthHandler(db *sql.DB, jwtSecret []byte) *AuthHandler {
	return &AuthHandler{
		DB:        db,
		JWTSecret: jwtSecret,
	}
}

// Realiza o login e gera um JWT
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request models.LoginRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	request.Username = strings.TrimSpace(request.Username)

	if request.Username == "" || request.Password == "" {
		http.Error(w, "Username e password são obrigatórios", http.StatusBadRequest)
		return
	}

	// Busca o usuário no banco de dados
	var user models.User

	err = h.DB.QueryRow(
		"SELECT id, username, password FROM users WHERE username = $1",
		request.Username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
			return
		}

		http.Error(w, "Erro ao consultar usuário", http.StatusInternalServerError)
		return
	}

	// Compara a senha recebida com o hash armazenado
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(request.Password),
	)

	if err != nil {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// Define a validade do token
	now := time.Now()
	expiration := now.Add(time.Hour)

	// Cria as informações que serão armazenadas no JWT
	claims := models.AuthClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}

	// Gera o token utilizando HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(h.JWTSecret)
	if err != nil {
		http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}

	// Retorna o token para o cliente
	response := models.LoginResponse{
		Token:     tokenString,
		TokenType: "Bearer",
		ExpiresIn: int64(time.Hour.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Retorna os dados do usuário autenticado
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Usuário não identificado", http.StatusUnauthorized)
		return
	}

	var user models.UserResponse

	err := h.DB.QueryRow(
		"SELECT id, username FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Username)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Usuário não encontrado", http.StatusNotFound)
			return
		}

		http.Error(w, "Erro ao consultar usuário", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
