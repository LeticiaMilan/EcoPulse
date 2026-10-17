package delivery

import (
	"report-service/internal/domain"
	"report-service/internal/infraestructure/persistence/postgres/db"
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
