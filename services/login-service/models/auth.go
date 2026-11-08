package models

import "github.com/golang-jwt/jwt/v5"

// Dados recebidos no login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Dados enviados após um login bem-sucedido
type LoginResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	ExpiresIn int64  `json:"expires_in"`
}

// Informações armazenadas dentro do JWT
type AuthClaims struct {
	UserID   string    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
