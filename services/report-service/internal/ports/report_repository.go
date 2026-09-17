package ports

import (
	"context"
	"report-service/internal/domain/entity"
)

type ReportRepository interface {
	Create(ctx context.Context, report *entity.Report) error
	Update(ctx context.Context, report *entity.Report) error
}
