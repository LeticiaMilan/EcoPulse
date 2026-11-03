package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/LeticiaMilan/ecopulse-login-mod/models"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userID"

// Recupera o ID do usuário autenticado no contexto da requisição
func UserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDKey).(int)
	return userID, ok
}

// Middleware responsável por validar o JWT
func AuthMiddleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Obtém o cabeçalho Authorization
			authHeader := r.Header.Get("Authorization")

			parts := strings.Fields(authHeader)

			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "Token não informado ou formato inválido", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Valida o token
			claims := &models.AuthClaims{}

			token, err := jwt.ParseWithClaims(
				tokenString,
				claims,
				func(token *jwt.Token) (interface{}, error) {
					if token.Method != jwt.SigningMethodHS256 {
						return nil, jwt.ErrSignatureInvalid
					}

					return secret, nil
				},
				jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			)

			if err != nil || !token.Valid {
				http.Error(w, "Token inválido ou expirado", http.StatusUnauthorized)
				return
			}

			if claims.UserID <= 0 {
				http.Error(w, "Token inválido", http.StatusUnauthorized)
				return
			}

			// Armazena o ID do usuário no contexto
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)

			// Encaminha a requisição com o contexto atualizado
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
