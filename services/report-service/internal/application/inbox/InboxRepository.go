package inbox

import (
	"context"

	"github.com/google/uuid"
)

type InboxRepository interface {
	GetUnprocessedMessage(ctx context.Context) (Inbox, error)
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error
}
