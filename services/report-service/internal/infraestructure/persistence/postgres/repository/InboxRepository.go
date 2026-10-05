package repository

import (
	"context"
	"log"
	"report-service/internal/application/inbox"
	"report-service/internal/infraestructure/persistence/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InboxRepositoryImpl struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewCreateInboxParams(inbox *inbox.Inbox) (db.CreateInboxParams, error) {
	var pgEvent pgtype.Text
	var pgStatus pgtype.Text
	var ReceivedAt pgtype.Timestamptz
	var ProcessedAt pgtype.Timestamptz

	if err := pgEvent.Scan(inbox.EventType); err != nil {
		return db.CreateInboxParams{}, err
	}

	if err := pgStatus.Scan(string(inbox.Status)); err != nil {
		return db.CreateInboxParams{}, err
	}

	if err := ReceivedAt.Scan(inbox.Received_at); err != nil {
		return db.CreateInboxParams{}, err
	}
	if err := ProcessedAt.Scan(inbox.Processed_at); err != nil {
		return db.CreateInboxParams{}, err
	}
	return db.CreateInboxParams{
		ID:          inbox.ID,
		EventType:   pgEvent,
		Status:      pgStatus,
		Content:     inbox.Content,
		Retries:     inbox.Retries,
		ReceivedAt:  ReceivedAt,
		ProcessedAt: ProcessedAt,
	}, nil
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
	params, err := NewCreateInboxParams(inbox)
	if err != nil {
		return err
	}
	InboxSaved, err := repo.queries.CreateInbox(ctx, params)
	if err != nil {
		return err
	}
	log.Print(InboxSaved)
	return nil
}

func (repo *InboxRepositoryImpl) Tx(ctx context.Context, fn func(repository inbox.InboxRepository) error) error {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	return nil
}
