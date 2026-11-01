# Estágio de compilação
FROM golang:1.27-alpine as BUILDER

WORKDIR /app

# Baixa as dependências do Go
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compila o binário estático
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Estágio de configuração da imagem Go mais leve
FROM alpine:3.19

WORKDIR /app

# Instala certificados ca-certificates caso sua app faça chamadas HTTPS externas
RUN apk --no-cache add ca-certificates

# Copia o binário compilado do stage anterior
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
