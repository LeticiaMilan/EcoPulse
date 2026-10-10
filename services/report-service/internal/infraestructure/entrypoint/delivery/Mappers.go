package delivery

import (
	"report-service/internal/domain"
	"report-service/internal/infraestructure/persistence/postgres/db"
)

func MapInboxApplicationToModel(inboxApplication *domain.Inbox) *db.Inbox {
	return &db.Inbox{}
}

func MapInboxModelToApplication(inboxModel *db.Inbox) *domain.Inbox {
	return &domain.Inbox{}
}

func MapReportGeneratedDTOToApplication(dto *ReportGeneratedEventDTO) *domain.Report {
	return &domain.Report{
		ID:        dto.ID,
		Latitude:  dto.Latitude,
		Longitude: dto.Longitude,
		Type:      domain.ReportType(dto.Type),
		Status:    domain.InProgress,
		UserID:    dto.UserID,
		Datetime:  dto.Datetime,
	}
}
