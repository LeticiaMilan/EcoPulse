package application

import (
	"context"
	"report-service/internal/domain"
)

type IReportUseCase interface {
	Execute(ctx context.Context, report *domain.Report) error
}
