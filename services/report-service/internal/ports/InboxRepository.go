package ports

import (
	"context"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/persistence/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type InboxRepository interface {
	Save(ctx context.Context, inbox *domain.Inbox) error
	GetUnprocessedMessage(ctx context.Context) (*domain.Inbox, error)
	MarkAsProcessed(ctx *context.Context, id uuid.UUID) error
	GetExecutor(ctx context.Context) *db.Queries
	BeginTransaction(ctx context.Context) (pgx.Tx, error)
}
