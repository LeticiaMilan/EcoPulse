package ports

import (
	"context"
	"report-service/internal/domain"

	"github.com/google/uuid"
)

type OutboxRepository interface {
	Save(ctx *context.Context, outbox *domain.Outbox) error
	GetUnprocessedMessage(ctx *context.Context) (domain.Outbox, error)
	MarkAsProcessed(ctx *context.Context, id uuid.UUID) error
}
