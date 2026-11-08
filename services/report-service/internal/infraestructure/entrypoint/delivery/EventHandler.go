package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"report-service/internal/application"
	"report-service/internal/domain"
)

type EventHandler interface {
	Handle(ctx context.Context, payload []byte, EventType string) error
}

/*
	type ReportGeneratedEventDTO struct {
		ID        uuid.UUID `json:"id"`
		Latitude  float64   `json:"latitude"`
		Longitude float64   `json:"longitude"`
		Type      string    `json:"type"`
		UserID    uuid.UUID `json:"user_id"`
		Datetime  time.Time `json:"datetime"`
	}

	type ReportValidatedEventDTO struct {
		ID              uuid.UUID `json:"id"`
		UserID          uuid.UUID `json:"usuario_id"`
		Tipo            string    `json:"tipo"`
		Descricao       string    `json:"descricao"`
		Latitude        float64   `json:"latitude"`
		Longitude       float64   `json:"longitude"`
		Votos_positivos int       `json:"votos_positivos"`
		Votos_negativos int       `json:"votos_negativos"`
		Estado          string    `json:"estado"`
	}

	type CenterReportAreaDefinedEventDTO struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	type PolygonReportAreaDefinedEventDTO struct {
		Type        string `json:"type"`
		Coordinates [][]float64
	}

	type ReportAreaDefinedEventDTO struct {
		Area_id uuid.UUID                        `json:"area_id"`
		Type    string                           `json:"type"`
		Status  string                           `json:"status"`
		Area_M2 float64                          `json:"area_m2"`
		Center  CenterReportAreaDefinedEventDTO  `json:"center"`
		Polygon PolygonReportAreaDefinedEventDTO `json:"polygon"`
		Reports []uuid.UUID                      `json:"report_ids"`
		Users   int                              `json:"users"`
	}
*/
type GenericHandler[T any] struct {
	UseCase application.IUseCase[T]
}

func NewHandler[T any](uc application.IUseCase[T]) *GenericHandler[T] {
	return &GenericHandler[T]{
		UseCase: uc,
	}
}

func (h *GenericHandler[T]) Handle(ctx context.Context, payload []byte, eventType string) error {
	return fmt.Errorf("handler for event %s is not implemented yet", eventType)
}

type ReportHandler struct {
	useCase application.IUseCase[*domain.Report]
}

type AreaHandler struct {
	useCase application.IUseCase[*domain.Area]
}

func (h *ReportHandler) Handle(ctx context.Context, payload []byte, EventType string) error {
	switch EventType {
	case string(domain.InboxGeneratedReportEvent):
		var GeneratedReportDTO application.ReportGeneratedEventDTO
		if err := json.Unmarshal(payload, &GeneratedReportDTO); err != nil {
			return fmt.Errorf("cannot unmarshal payload: %w", err)
		}
		domainInbox := MapReportGeneratedDTOToApplication(&GeneratedReportDTO)
		if err := h.useCase.Execute(ctx, domainInbox); err != nil {
			log.Printf("[GeneratedEventHandler] Error while processing event: %w", err.Error())
			return err
		}
	case string(domain.InboxValidatedReportEvent):
		var ValidatedReportDTO application.ReportValidatedEventDTO
		if err := json.Unmarshal(payload, &ValidatedReportDTO); err != nil {
			return fmt.Errorf("cannot unmarshal payload: %w", err)
		}

		if err := h.useCase.Execute(ctx, MapReportValidatedDTOToApplication(&ValidatedReportDTO)); err != nil {
			log.Printf("[ValidatedEventHandler] Error while processing event: %w", err.Error())
			return err
		}

	default:
		return fmt.Errorf("invalid event type: %s", EventType)
	}

	return nil
}

func (h *AreaHandler) Handle(ctx context.Context, payload []byte, EventType string) error {
	switch EventType {
	case string(domain.InboxAreaDefinedReportEvent):
		var AreaDefinedReportDTO application.ReportAreaDefinedEventDTO
		if err := json.Unmarshal(payload, &AreaDefinedReportDTO); err != nil {
			return fmt.Errorf("cannot unmarshal payload: %w", err)
		}
		if err := h.useCase.Execute(ctx, MapReportAffectedAreaDTOToApplication(&AreaDefinedReportDTO)); err != nil {
			log.Printf("[ValidatedEventHandler] Error while processing event: %w", err.Error())
			return err
		}

	default:
		return fmt.Errorf("invalid event type: %s", EventType)
	}

	return nil
}
