package repository

import (
	"context"
	"log"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/entrypoint/delivery"
	"report-service/internal/infraestructure/persistence/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InboxRepositoryImpl struct {
	pool    *pgxpool.Pool
	Queries *db.Queries
}

func NewCreateInboxParams(inbox *domain.Inbox) (*db.CreateInboxParams, error) {
	var pgEvent pgtype.Text
	var pgStatus pgtype.Text
	var ReceivedAt pgtype.Timestamptz
	var ProcessedAt pgtype.Timestamptz

	if err := pgEvent.Scan(string(inbox.EventType)); err != nil {
		return nil, err
	}

	if err := pgStatus.Scan(string(inbox.Status)); err != nil {
		return nil, err
	}

	if err := ReceivedAt.Scan(inbox.Received_at); err != nil {
		return nil, err
	}
	if err := ProcessedAt.Scan(inbox.Processed_at); err != nil {
		return nil, err
	}
	return &db.CreateInboxParams{
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
		Queries: db.New(pool),
	}
}

func (repo *InboxRepositoryImpl) GetUnprocessedMessage(ctx context.Context) (*domain.Inbox, error) {
	Query := repo.GetExecutor(ctx)
	retorno, err := Query.PoolUnprocessedMessage(ctx)
	if err != nil {
		return nil, err
	}
	return delivery.MapInboxModelToApplication(&retorno), nil
}

func (repo *InboxRepositoryImpl) MarkAsProcessed(ctx *context.Context, id uuid.UUID) error {
	return nil
}

func (repo *InboxRepositoryImpl) Save(ctx context.Context, inbox *domain.Inbox) error {
	params, err := NewCreateInboxParams(inbox)
	if err != nil {
		return err
	}
	Query := repo.GetExecutor(ctx)
	InboxSaved, err := Query.CreateInbox(ctx, *params)
	if err != nil {
		return err
	}
	log.Print(InboxSaved)
	return nil
}

func (repo *InboxRepositoryImpl) GetExecutor(ctx context.Context) *db.Queries {
	if tx, ok := ctx.Value("tx").(pgx.Tx); ok {
		return db.New(repo.pool).WithTx(tx)
	}
	return repo.Queries
}

func (repo *InboxRepositoryImpl) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
