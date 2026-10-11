package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"report-service/internal/application"
	"time"

	"github.com/google/uuid"
)

type EventHandler interface {
	Handle(ctx context.Context, payload []byte) error
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

func (h *ReportGeneratedHandler) Handle(ctx context.Context, payload []byte) error {
	var ReportDTO ReportGeneratedEventDTO

	if err := json.Unmarshal(payload, &ReportDTO); err != nil {
		return fmt.Errorf("cannot unmarshal payload: %w", err)
	}

	if err := h.ReportUseCase.Execute(MapReportGeneratedDTOToApplication(&ReportDTO)); err != nil {
		return fmt.Errorf("cannot execute report: %w", err)
	}

	return nil
}

type ReportRejectedHandler struct{}
type ReportValidatedHandler struct{}
type ReportAreaDefinedHandler struct{}
