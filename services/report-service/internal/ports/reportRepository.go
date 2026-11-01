package ports

import (
	"context"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/persistence/postgres/db"

	"github.com/jackc/pgx/v5"
)

type ReportRepository interface {
	Create(ctx context.Context, report *domain.Report) error
	Validate(ctx context.Context, report *domain.Report) (*domain.Report, error)
	Reject(ctx context.Context, report *domain.Report) (*domain.Report, error)
	DefineArea(ctx context.Context, report *domain.Report) (*domain.Report, error)
	GetExecutor(ctx context.Context) *db.Queries
	BeginTransaction(ctx context.Context) (pgx.Tx, error)
}
