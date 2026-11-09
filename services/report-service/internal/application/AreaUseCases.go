package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"report-service/internal/domain"
	"report-service/internal/ports"

	"github.com/google/uuid"
)

type CenterReportAreaDefinedEventDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ReportAreaDefinedEventDTO struct {
	Area_id uuid.UUID                       `json:"area_id"`
	Type    string                          `json:"type"`
	Status  string                          `json:"status"`
	Area_M2 float64                         `json:"area_m2"`
	Center  CenterReportAreaDefinedEventDTO `json:"center"`
	Polygon json.RawMessage                 `json:"polygon"`
	Reports []uuid.UUID                     `json:"report_ids"`
	Users   int                             `json:"users"`
	Votes   domain.AreaVotes                `json:"votes"`
}

type AreaUseCase struct {
	AreaRepository   ports.AreaRepository
	ReportRepository ports.ReportRepository
	OutboxRepository ports.OutboxRepository
}

func NewAreaUseCase(AreaRepository ports.AreaRepository, ReportRepository ports.ReportRepository, OutboxRepository ports.OutboxRepository) *AreaUseCase {
	return &AreaUseCase{
		AreaRepository:   AreaRepository,
		ReportRepository: ReportRepository,
		OutboxRepository: OutboxRepository,
	}
}

type AreaDefinedUseCase struct {
	AreaUseCase *AreaUseCase
}

func NewAreaDefined(AreaUseCase *AreaUseCase) *AreaDefinedUseCase {
	return &AreaDefinedUseCase{
		AreaUseCase: AreaUseCase,
	}
}

func (uc *AreaDefinedUseCase) Execute(ctx context.Context, Area *domain.Area) error {

	if err := uc.AreaUseCase.AreaRepository.Save(ctx, Area); err != nil {
		return err
	}

	for _, reportID := range Area.Reports {
		report, erro := uc.AreaUseCase.ReportRepository.DefineArea(ctx, &domain.Report{ID: reportID, Area: Area.Id})
		if erro != nil {
			if errors.Is(erro, sql.ErrNoRows) {
				continue
			}
			return fmt.Errorf("falha ao associar a área ao relato %s: %w", reportID, erro)
		} else {
			if report.Area != Area.Id {
				log.Printf("[AreaDefinedUseCase] O relato %s possui uma área distinta associada. A avançar.", reportID)
				continue
			}

			reportDTO := MapReportToReportDTO(report)
			content, err := json.Marshal(reportDTO)

			if err != nil {
				return fmt.Errorf("falha na serialização do DTO do relato %s: %w", reportID, err)
			}

			outbox := domain.NewOutbox(domain.OutboxReportAreaDefined, content)
			if err := uc.AreaUseCase.OutboxRepository.Save(ctx, outbox); err != nil {
				return fmt.Errorf("falha ao registar o evento de outbox para o relato %s: %w", reportID, err)
			}

		}
	}
	return nil
}
