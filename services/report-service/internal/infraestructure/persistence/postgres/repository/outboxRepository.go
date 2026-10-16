package repository

import (
	"context"
	"log"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/persistence/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRepositoryImpl struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewCreateOutboxParams(outbox domain.Outbox) (*db.CreateOutboxParams, error) {
	var pgEvent pgtype.Text
	var pgStatus pgtype.Text
	var CreatedAt pgtype.Timestamptz
	var ProcessedAt pgtype.Timestamptz

	if err := pgEvent.Scan(string(outbox.EventType)); err != nil {
		return nil, err
	}

	if err := pgStatus.Scan(string(outbox.Status)); err != nil {
		return nil, err
	}

	if err := CreatedAt.Scan(outbox.Created_at); err != nil {
		return nil, err
	}
	if err := ProcessedAt.Scan(outbox.Processed_at); err != nil {
		return nil, err
	}
	return &db.CreateOutboxParams{
		ID:          outbox.ID,
		EventType:   pgEvent,
		Status:      pgStatus,
		Content:     outbox.Content,
		Retries:     outbox.Retries,
		CreatedAt:   CreatedAt,
		ProcessedAt: ProcessedAt,
	}, nil
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepositoryImpl {
	return &OutboxRepositoryImpl{
		pool:    pool,
		queries: db.New(pool),
	}
}
func (repo *OutboxRepositoryImpl) GetUnprocessedMessage(ctx *context.Context) (domain.Outbox, error) {

	return domain.Outbox{}, nil
}

func (repo *OutboxRepositoryImpl) MarkAsProcessed(ctx *context.Context, id uuid.UUID) error {

	return nil
}

func (repo *OutboxRepositoryImpl) Save(ctx context.Context, outbox domain.Outbox) error {
	log.Printf("Saving outbox %+v", outbox)
	params, err := NewCreateOutboxParams(outbox)
	if err != nil {
		return err
	}
	Query := repo.GetExecutor(ctx)
	OutboxSaved, err := Query.CreateOutbox(ctx, *params)
	if err != nil {
		return err
	}
	log.Printf("Saved outbox %+v", OutboxSaved)
	return nil
}

func (repo *OutboxRepositoryImpl) GetExecutor(ctx context.Context) *db.Queries {
	if tx, ok := ctx.Value("tx").(pgx.Tx); ok {
		return db.New(repo.pool).WithTx(tx)
	}
	return repo.queries
}

func (repo *OutboxRepositoryImpl) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
