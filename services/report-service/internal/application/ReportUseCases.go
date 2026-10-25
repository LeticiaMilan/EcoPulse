package application

import (
	"context"
	"encoding/json"
	"log"
	"report-service/internal/domain"
	"report-service/internal/ports"
	"report-service/pkg"
	"time"

	"github.com/google/uuid"
)

type ReportUseCase struct {
	ReportRepository ports.ReportRepository
	OutboxRepository ports.OutboxRepository
	clock            ports.BrasilClock
}

func NewReportUseCase(reportRepository ports.ReportRepository, clock ports.BrasilClock, OutboxRepository ports.OutboxRepository) *ReportUseCase {
	return &ReportUseCase{
		ReportRepository: reportRepository,
		clock:            clock,
		OutboxRepository: OutboxRepository,
	}
}

type GeneratedReportUseCase struct {
	ReportUseCase *ReportUseCase
}

func NewGeneratedReportUseCase(ReportUseCase *ReportUseCase) *GeneratedReportUseCase {
	return &GeneratedReportUseCase{
		ReportUseCase: ReportUseCase,
	}
}

func (uc *GeneratedReportUseCase) Execute(ctx context.Context, report *domain.Report) error {
	if err := report.Validate(); err != nil {
		log.Printf("[ReportUseCase] Execute err: %+v", err)
		return err
	}
	if err := uc.IsCoordinatesValid(report); err != nil {
		return err
	}

	if err := uc.isDatetimeValid(report, uc.ReportUseCase.clock); err != nil {
		return err
	}

	if err := uc.ReportUseCase.ReportRepository.Create(ctx, report); err != nil {
		return err
	}

	reportDTO := MapReportToReportDTO(report)
	content, err := json.Marshal(reportDTO)

	if err != nil {
		return err
	}

	outbox := domain.NewOutbox(domain.OutboxReportGenerated, content)

	if err := uc.ReportUseCase.OutboxRepository.Save(ctx, outbox); err != nil {
		return err
	}

	return nil
}

func (uc *GeneratedReportUseCase) IsCoordinatesValid(r *domain.Report) error {
	if r.Latitude <= -90 || r.Latitude >= 90 || r.Longitude <= -180 || r.Longitude >= 180 {
		return domain.ErrInvalidCoordinates
	}

	return nil
}

func (uc *GeneratedReportUseCase) isDatetimeValid(report *domain.Report, clock ports.BrasilClock) error {
	today := pkg.StartOfDay(clock.Now())

	if today.After(pkg.StartOfDay(report.Datetime)) && report.Status == domain.InProgress {
		return domain.ErrInvalidDatetime
	}

	return nil
}

type ValidatedReportUseCase struct {
	ReportUseCase *ReportUseCase
}

func NewValidatedReport(ReportUseCase *ReportUseCase) *ValidatedReportUseCase {
	return &ValidatedReportUseCase{
		ReportUseCase: ReportUseCase,
	}
}

func (uc *ValidatedReportUseCase) Execute(ctx context.Context, report *domain.Report) error {

	report, err := uc.ReportUseCase.ReportRepository.Validate(ctx, report)

	if err != nil {
		return err
	}

	reportDTO := MapReportToReportDTO(report)
	content, err := json.Marshal(reportDTO)

	outbox := domain.NewOutbox(domain.OutboxReportValidated, content)

	if err := uc.ReportUseCase.OutboxRepository.Save(ctx, outbox); err != nil {
		return err
	}
	return nil
}

type RejectedReportUseCase struct {
	ReportUseCase *ReportUseCase
}

func NewRejectedReport(ReportUseCase *ReportUseCase) *RejectedReportUseCase {
	return &RejectedReportUseCase{
		ReportUseCase: ReportUseCase,
	}
}
func (uc *RejectedReportUseCase) Execute(ctx context.Context, report *domain.Report) error {
	report, err := uc.ReportUseCase.ReportRepository.Reject(ctx, report)

	if err != nil {
		return err
	}

	reportDTO := MapReportToReportDTO(report)
	content, err := json.Marshal(reportDTO)

	outbox := domain.NewOutbox(domain.OutboxReportRejected, content)

	if err := uc.ReportUseCase.OutboxRepository.Save(ctx, outbox); err != nil {
		return err
	}
	return nil
}

type AreaDefinedReportUseCase struct {
	ReportUseCase *ReportUseCase
}

func NewAreaDefinedReport(ReportUseCase *ReportUseCase) *AreaDefinedReportUseCase {
	return &AreaDefinedReportUseCase{
		ReportUseCase: ReportUseCase,
	}
}
func (uc *AreaDefinedReportUseCase) Execute(ctx context.Context, report *domain.Report) error {
	report, err := uc.ReportUseCase.ReportRepository.DefineArea(ctx, report)

	if err != nil {
		return err
	}

	reportDTO := MapReportToReportDTO(report)
	content, err := json.Marshal(reportDTO)

	outbox := domain.NewOutbox(domain.OutboxReportCreated, content)

	if err := uc.ReportUseCase.OutboxRepository.Save(ctx, outbox); err != nil {
		return err
	}
	return nil
}

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

func MapReportToReportDTO(report *domain.Report) ReportDTO {
	return ReportDTO{
		ID:        report.ID,
		Latitude:  report.Latitude,
		Longitude: report.Longitude,
		Type:      string(report.Type),
		Status:    string(report.Status),
		UserId:    report.UserID,
		Area:      "",
		Datetime:  report.Datetime,
	}
}
