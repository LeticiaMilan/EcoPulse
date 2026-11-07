package repository

import (
	"context"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/persistence/postgres/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AreaRepositoryImpl struct {
	pool    *pgxpool.Pool
	Queries *db.Queries
}

func NewAreaRepository(pool *pgxpool.Pool) *AreaRepositoryImpl {
	return &AreaRepositoryImpl{
		pool:    pool,
		Queries: db.New(pool),
	}
}

func (repo *AreaRepositoryImpl) Save(ctx context.Context, area *domain.Area) error {
	params := db.CreateAreaParams{
		ID:        area.Id,
		Type:      area.Type,
		Status:    area.Status,
		AreaM2:    area.Area_M2,
		CenterLat: area.Center.Latitude,
		CenterLng: area.Center.Longitude,
		Polygon:   area.Polygon,
		UpdatedAt: pgtype.Timestamptz{},
	}

	Query := repo.GetExecutor(ctx)
	_, err := Query.CreateArea(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (repo *AreaRepositoryImpl) GetExecutor(ctx context.Context) *db.Queries {
	if tx, ok := ctx.Value("tx").(pgx.Tx); ok {
		return db.New(repo.pool).WithTx(tx)
	}
	return repo.Queries
}

func (repo *AreaRepositoryImpl) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
