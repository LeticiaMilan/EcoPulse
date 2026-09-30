package inbox

import (
	"context"

	"github.com/google/uuid"
)

type InboxRepository interface {
	Save(ctx context.Context, inbox *Inbox) error
	GetUnprocessedMessage(ctx context.Context) (Inbox, error)
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error
}
