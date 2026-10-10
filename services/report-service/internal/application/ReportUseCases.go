package application

import (
	"context"
	"encoding/json"
	"report-service/internal/domain"
	"report-service/internal/ports"
	"time"

	"github.com/google/uuid"
)

type ReportDTO struct {
	ID        uuid.UUID `json:"id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	UserId    uuid.UUID `json:"user_id"`
	Area      string    `json:"area"`
	Datetime  time.Time `json:"datetime"`
}

type GeneratedReportUseCase struct {
	ReportRepository ports.ReportRepository
	OutboxRepository ports.OutboxRepository
	GeocodingAPI     ports.GeocodingAPI
	clock            domain.BrasilClock
}

func NewProcessReportUseCase(reportRepository ports.ReportRepository, geocodingAPI ports.GeocodingAPI, clock domain.BrasilClock) *GeneratedReportUseCase {
	return &GeneratedReportUseCase{
		ReportRepository: reportRepository,
		GeocodingAPI:     geocodingAPI,
		clock:            clock,
	}
}

func (uc *GeneratedReportUseCase) Execute(report *domain.Report) error {
	ctx := context.Background()

	if err := report.Validate(uc.clock); err != nil {
		return err
	}

	if err := uc.IsCoordinatesValid(&ctx, report); err != nil {
		return err
	}

	if err := uc.ReportRepository.Create(ctx, report); err != nil {
		return err
	}

	reportDTO := ReportDTO{
		ID:        report.ID,
		Latitude:  report.Latitude,
		Longitude: report.Longitude,
		Type:      string(report.Type),
		Status:    string(report.Status),
		UserId:    report.UserID,
		Area:      "",
		Datetime:  report.Datetime,
	}
	content, err := json.Marshal(reportDTO)

	if err != nil {
		return err
	}

	outbox := domain.Outbox{
		ID:           report.ID,
		EventType:    "EVENTS.CREATED.REPORT",
		Status:       domain.OutboxStatusPending,
		Content:      content,
		Retries:      0,
		Created_at:   time.Now(),
		Processed_at: time.Time{},
	}

	if err := uc.OutboxRepository.Save(&ctx, &outbox); err != nil {
		return err
	}

	return nil
}

func (uc *GeneratedReportUseCase) IsCoordinatesValid(ctx *context.Context, r *domain.Report) error {
	if r.Latitude <= -90 || r.Latitude >= 90 || r.Longitude <= -180 || r.Longitude >= 180 {
		return domain.ErrInvalidCoordinates
	}
	inputParams := ports.GeocodingInputParams{
		Latitude:  r.Latitude,
		Longitude: r.Longitude,
	}
	if isBrasil, err := uc.GeocodingAPI.IsBrazil(*ctx, &inputParams); !isBrasil || err != nil {
		if err != nil {
			return err
		}
		return domain.ErrInvalidLocation
	}
	return nil
}

type CancelReportUseCase struct{}

func (uc *CancelReportUseCase) Execute() error {
	return nil
}
