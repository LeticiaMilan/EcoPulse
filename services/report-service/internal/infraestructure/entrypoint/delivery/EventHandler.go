package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"report-service/internal/application"
	"report-service/internal/infraestructure/persistence/postgres/db"
	"time"

	"github.com/google/uuid"
)

type EventHandler interface {
	Handle(ctx context.Context, payload []byte, work *db.UnitOfWork) error
}

type ReportGeneratedEventDTO struct {
	ID        uuid.UUID `json:"id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Type      string    `json:"type"`
	UserID    uuid.UUID `json:"user_id"`
	Datetime  time.Time `json:"datetime"`
}

type ReportGeneratedHandler struct {
	ReportUseCase *application.GeneratedReportUseCase
}

func NewReportGeneratedHandler(reportUseCase *application.GeneratedReportUseCase) *ReportGeneratedHandler {
	return &ReportGeneratedHandler{ReportUseCase: reportUseCase}
}

func (h *ReportGeneratedHandler) Handle(ctx context.Context, payload []byte, work *db.UnitOfWork) error {
	var ReportDTO ReportGeneratedEventDTO

	if err := json.Unmarshal(payload, &ReportDTO); err != nil {
		return fmt.Errorf("cannot unmarshal payload: %w", err)
	}
	log.Printf("Received event type %s on GeneratedHandler", ReportDTO)

	domainInbox := MapReportGeneratedDTOToApplication(&ReportDTO)
	log.Printf("Parsed domainInbox: %s ", domainInbox)

	if err := h.ReportUseCase.Execute(ctx, domainInbox, work); err != nil {
		log.Printf("[EventHandler] Error while processing event: %w", err.Error())
		return err
	}

	return nil
}

type ReportRejectedHandler struct{}
type ReportValidatedHandler struct{}
type ReportAreaDefinedHandler struct{}
