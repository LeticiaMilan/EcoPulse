package ports

import (
	"context"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/persistence/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OutboxRepository interface {
	Save(ctx context.Context, outbox domain.Outbox) error
	GetUnprocessedMessage(ctx *context.Context) (domain.Outbox, error)
	MarkAsProcessed(ctx *context.Context, id uuid.UUID) error
	GetExecutor(ctx context.Context) *db.Queries
	BeginTransaction(ctx context.Context) (pgx.Tx, error)
}
