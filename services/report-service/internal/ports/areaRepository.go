package ports

import (
	"context"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/persistence/postgres/db"

	"github.com/jackc/pgx/v5"
)

type AreaRepository interface {
	Save(ctx context.Context, inbox *domain.Area) error
	GetExecutor(ctx context.Context) *db.Queries
	BeginTransaction(ctx context.Context) (pgx.Tx, error)
}
