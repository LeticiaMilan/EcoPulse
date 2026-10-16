package ports

import (
	"context"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/persistence/postgres/db"

	"github.com/jackc/pgx/v5"
)

type ReportRepository interface {
	Create(ctx context.Context, report *domain.Report) error
	Update(ctx context.Context, report *domain.Report) error
	GetExecutor(ctx context.Context) *db.Queries
	BeginTransaction(ctx context.Context) (pgx.Tx, error)
}
