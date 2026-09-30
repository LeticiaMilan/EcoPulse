package repository

import (
	"context"
	"report-service/internal/application/inbox"
	"report-service/internal/infraestructure/persistence/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InboxRepositoryImpl struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewInboxRepository(pool *pgxpool.Pool) *InboxRepositoryImpl {
	return &InboxRepositoryImpl{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (repo *InboxRepositoryImpl) GetUnprocessedMessage(ctx context.Context) (inbox.Inbox, error) {

	return inbox.Inbox{}, nil
}

func (repo *InboxRepositoryImpl) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {

	return nil
}

func (repo *InboxRepositoryImpl) Save(ctx context.Context, inbox *inbox.Inbox) error {

	return nil
}
