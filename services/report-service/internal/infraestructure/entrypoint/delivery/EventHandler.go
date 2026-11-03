package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"report-service/internal/application"
	"report-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

type EventHandler interface {
	Handle(ctx context.Context, payload []byte, EventType string) error
}

type ReportGeneratedEventDTO struct {
	ID        uuid.UUID `json:"id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Type      string    `json:"type"`
	UserID    uuid.UUID `json:"user_id"`
	Datetime  time.Time `json:"datetime"`
}

type ReportValidatedEventDTO struct{}

type ReportReceivedEventDTO struct{}

type ReportAreaDefinedEventDTO struct{}

type ReportHandler struct {
	ReportUseCase application.IReportUseCase
}

func NewHandler(reportUseCase application.IReportUseCase) *ReportHandler {
	return &ReportHandler{ReportUseCase: reportUseCase}
}

func (h *ReportHandler) Handle(ctx context.Context, payload []byte, EventType string) error {
	switch EventType {
	case string(domain.InboxGeneratedReportEvent):
		var GeneratedReportDTO ReportGeneratedEventDTO
		if err := json.Unmarshal(payload, &GeneratedReportDTO); err != nil {
			return fmt.Errorf("cannot unmarshal payload: %w", err)
		}
		domainInbox := MapReportGeneratedDTOToApplication(&GeneratedReportDTO)
		if err := h.ReportUseCase.Execute(ctx, domainInbox); err != nil {
			log.Printf("[GeneratedEventHandler] Error while processing event: %w", err.Error())
			return err
		}
	case string(domain.InboxValidatedReportEvent):
		var ValidatedReportDTO ReportValidatedEventDTO
		if err := json.Unmarshal(payload, &ValidatedReportDTO); err != nil {
			return fmt.Errorf("cannot unmarshal payload: %w", err)
		}
		/*

			if err := h.ReportUseCase.Execute(ctx, ...); err != nil{
				log.Printf("[ValidatedEventHandler] Error while processing event: %w", err.Error())
				return err
			}
		*/
	case string(domain.InboxRejectedReportEvent):
		var RejectedReportDTO ReportReceivedEventDTO
		if err := json.Unmarshal(payload, &RejectedReportDTO); err != nil {
			return fmt.Errorf("cannot unmarshal payload: %w", err)
		}
		/*

			if err := h.ReportUseCase.Execute(ctx, ...); err != nil{
				log.Printf("[ValidatedEventHandler] Error while processing event: %w", err.Error())
				return err
			}
		*/

	case string(domain.InboxAreaDefinedReportEvent):
		var AreaDefinedReportDTO ReportAreaDefinedEventDTO
		if err := json.Unmarshal(payload, &AreaDefinedReportDTO); err != nil {
			return fmt.Errorf("cannot unmarshal payload: %w", err)
		}

		/*

			if err := h.ReportUseCase.Execute(ctx, ...); err != nil{
				log.Printf("[ValidatedEventHandler] Error while processing event: %w", err.Error())
				return err
			}
		*/
	default:
		return fmt.Errorf("invalid event type: %s", EventType)
	}

	return nil
}
