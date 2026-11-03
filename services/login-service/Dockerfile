# Etapa 1: Compilação da aplicação
FROM golang:1.27.1-alpine AS builder

WORKDIR /app

# Copia os arquivos de dependências
COPY go.mod go.sum ./

# Baixa as dependências
RUN go mod download

# Copia o restante do código
COPY . .

# Compila a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o login-service .

# Etapa 2: Imagem final
FROM alpine:3.22

WORKDIR /app

# Certificados necessários para conexões TLS, se utilizadas
RUN apk --no-cache add ca-certificates

# Copia somente o executável compilado
COPY --from=builder /app/login-service .

# Porta utilizada pela API
EXPOSE 8080

# Inicializa o microsserviço
CMD ["./login-service"]