FROM golang:1.27.1-alpine AS builder

WORKDIR /app

# Copy modules from the subfolder
COPY go.mod go.sum ./services/report-service/
RUN cd services/report-service && go mod download

# Copy the entire workspace
COPY . .

# Build by specifying the path from the workspace root
RUN CGO_ENABLED=0 GOOS=linux go build -o report-service ./cmd/api

# Etapa 2: Imagem final
FROM alpine:3.22

WORKDIR /app

# Certificados necessários para conexões TLS, se utilizadas
RUN apk --no-cache add ca-certificates

# Copia somente o executável compilado
COPY --from=builder /app/report-service .

COPY --from=builder /app/internal/infraestructure/persistence/postgres/schema.sql ./internal/infraestructure/persistence/postgres/schema.sql

# Porta utilizada pela API
EXPOSE 85

CMD ["./report-service"]