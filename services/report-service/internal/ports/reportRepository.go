package ports

import (
	"context"
	"report-service/internal/domain"
)

type ReportRepository interface {
	Create(ctx context.Context, report *domain.Report) error
	Update(ctx context.Context, report *domain.Report) error
}
