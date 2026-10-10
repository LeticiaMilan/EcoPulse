package ports

import (
	"context"
	"report-service/internal/domain"

	"github.com/google/uuid"
)

type InboxRepository interface {
	Save(ctx *context.Context, inbox *domain.Inbox) error
	GetUnprocessedMessage(ctx *context.Context) (domain.Inbox, error)
	MarkAsProcessed(ctx *context.Context, id uuid.UUID) error
	Tx(ctx *context.Context, fn func(repository InboxRepository) error) error
}
