package delivery

import (
	"report-service/internal/application"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/persistence/postgres/db"
	"time"

	"github.com/google/uuid"
)

func MapInboxApplicationToModel(inboxApplication *domain.Inbox) *db.Inbox {
	return &db.Inbox{}
}

func MapInboxModelToApplication(inboxModel *db.Inbox) *domain.Inbox {
	return &domain.Inbox{
		ID:           inboxModel.ID,
		EventType:    domain.InboxEventType(inboxModel.EventType.String),
		Status:       domain.InboxStatus(inboxModel.Status.String),
		Content:      inboxModel.Content,
		Retries:      inboxModel.Retries,
		Received_at:  inboxModel.ReceivedAt.Time,
		Processed_at: inboxModel.ProcessedAt.Time,
	}
}

func MapReportModelToApplication(reportModel *db.Report) *domain.Report {
	return &domain.Report{
		ID:        reportModel.ID,
		Latitude:  reportModel.Latitude.Float64,
		Longitude: reportModel.Longitude.Float64,
		Type:      domain.ReportType(reportModel.Type.String),
		Status:    domain.ReportStatus(reportModel.Status.String),
		UserID:    reportModel.UserID,
		Area:      reportModel.Area,
		Datetime:  reportModel.Datetime.Time,
	}
}
func MapOutboxModelToApplication(outboxModel *db.Outbox) *domain.Outbox {
	return &domain.Outbox{
		ID:           outboxModel.ID,
		EventType:    domain.OutboxEventType(outboxModel.EventType.String),
		Status:       domain.OutboxStatus(outboxModel.Status.String),
		Content:      outboxModel.Content,
		Retries:      outboxModel.Retries,
		Created_at:   outboxModel.CreatedAt.Time,
		Processed_at: outboxModel.ProcessedAt.Time,
	}
}

func MapReportGeneratedDTOToApplication(dto *application.ReportGeneratedEventDTO) *domain.Report {
	return &domain.Report{
		ID:          dto.ID,
		Latitude:    dto.Latitude,
		Longitude:   dto.Longitude,
		Type:        domain.ReportType(dto.Type),
		Status:      domain.InProgress,
		UserID:      dto.UserID,
		Datetime:    dto.Datetime,
		Description: dto.Description,
	}
}

func MapReportValidatedDTOToApplication(dto *application.ReportValidatedEventDTO) *domain.Report {
	return &domain.Report{
		ID:        dto.ID,
		Latitude:  dto.Latitude,
		Longitude: dto.Longitude,
		Type:      domain.ReportType(dto.Tipo),
		Status:    domain.ReportStatus(dto.Estado),
		UserID:    dto.UserID,
		Area:      uuid.UUID{},
		Datetime:  time.Time{},
	}
}

func MapReportAffectedAreaDTOToApplication(dto *application.ReportAreaDefinedEventDTO) *domain.Area {
	return &domain.Area{
		Id:      dto.Area_id,
		Type:    dto.Type,
		Status:  dto.Status,
		Area_M2: dto.Area_M2,
		Center: domain.AreaCenter{
			Latitude:  dto.Center.Latitude,
			Longitude: dto.Center.Longitude,
		},
		Polygon: dto.Polygon,
		Reports: dto.Reports,
		Users:   dto.Users,
		Votes:   dto.Votes,
	}
}
