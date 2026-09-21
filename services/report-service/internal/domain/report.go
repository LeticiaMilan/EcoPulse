package domain

import (
	"errors"
	"report-service/pkg"
	"time"

	"github.com/google/uuid"
)

type ReportTypeChecker interface {
	Exists(ReportType ReportType) error
}

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
func (r *Report) Validate(clock BrasilClock, checker ReportTypeChecker) error {

	if err := r.IsCoordinatesValid(); err != nil {
		return err
	}

	if err := r.isReportTypeValid(checker); err != nil {
		return err
	}

	if err := r.isDatetimeValid(clock); err != nil {
		return err
	}

	// 4a validação: Área afetada >= 0

	return nil
}

func (r *Report) Transite() {

}

func (r *Report) IsCoordinatesValid() error {
	if r.Latitude <= -90 || r.Latitude >= 90 || r.Longitude <= -180 || r.Longitude >= 180 {
		return ErrInvalidCoordinates
	}

	// if não é brasil
	//  return ErrInvalidLocation
	//
	return nil
}

func (r *Report) isReportTypeValid(checker ReportTypeChecker) error {
	if err := checker.Exists(r.Type); err != nil {
		return ErrInvalidReportType
	}
	return nil
}

func (r *Report) isDatetimeValid(clock BrasilClock) error {
	today := pkg.StartOfDay(clock.Now())

	if today.After(pkg.StartOfDay(r.Datetime)) && r.Status == InProgress {
		return ErrInvalidDatetime
	}

	return nil
}
