package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCoordinates = errors.New("Latitude ou longitude inválidas")
	ErrInvalidLocation    = errors.New("Localização inválida")
	ErrInvalidReportType  = errors.New("Tipo de relato inválido")
	ErrInvalidDatetime    = errors.New("Data inválida")
)

type ReportStatus string

const (
	Created    ReportStatus = "created"    // Estado final (Já pode ser sincronizado)
	InProgress ReportStatus = "inProgress" // Usuário criou, mas ainda não foi liberado
	Validating ReportStatus = "validating" // Etapa de validação por outros usuários
)

type ReportType string

const (
	Landslide    ReportType = "Landslide"     // Deslizamento
	Flooding     ReportType = "Flooding"      // Alagamento
	Thunderstorm ReportType = "Thunderstorm"  // Tempestade
	FireDisaster ReportType = "Fire Disaster" // Incendios
)

type AffectedArea struct{}

type Report struct {
	ID        uuid.UUID
	Latitude  float64
	Longitude float64
	Type      ReportType
	Status    ReportStatus
	UserID    uuid.UUID
	Area      AffectedArea
	Datetime  time.Time
}

// Validações futuras
func (r *Report) Validate() error {

	if err := r.isReportTypeValid(); err != nil {
		return err
	}

	// 4a validação: Área afetada >= 0

	return nil
}

func (r *Report) Transite() {

}

func (r *Report) isReportTypeValid() error {
	if r.Type != "Landslide" && r.Type != "Flooding" && r.Type != "Thunderstorm" && r.Type != "Fire Disaster" {
		return ErrInvalidReportType
	}
	return nil
}
