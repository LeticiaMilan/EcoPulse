package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ReportGeneratedEventDTO struct {
	ID        uuid.UUID `json:"id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Type      string    `json:"type"`
	UserID    uuid.UUID `json:"user_id"`
	Datetime  time.Time `json:"datetime"`
}

type ReportGeneratedHandler struct {
	reportUseCase *ProcessReportUseCase
}

func NewReportGeneratedHandler(reportUseCase *ProcessReportUseCase) *ReportGeneratedHandler {
	return &ReportGeneratedHandler{reportUseCase: reportUseCase}
}

func (h *ReportGeneratedHandler) Handle(ctx context.Context, payload []byte) error {
	var ReportDTO ReportGeneratedEventDTO

	if err := json.Unmarshal(payload, &ReportDTO); err != nil {
		return fmt.Errorf("cannot unmarshal payload: %w", err)
	}

	return nil
}
