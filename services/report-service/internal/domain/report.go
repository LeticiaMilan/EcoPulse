package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCoordinates = errors.New("Latitude ou longitude inválidas")
	ErrInvalidLocation    = errors.New("Localização inválida")
	ErrInvalidReportType  = errors.New("Tipo de relato inválido")
	ErrInvalidDatetime    = errors.New("Data inválida")
	ErrInvalidUser        = errors.New("Usuário inválido/nulo")
)

// AffectedArea representa a tabela affected_areas
type ReportAffectedAreaDTO struct {
	ID        uuid.UUID       `json:"id"`
	Type      string          `json:"type"`
	Status    string          `json:"status"`
	AreaM2    float64         `json:"area_m2"`
	CenterLat float64         `json:"center_lat"`
	CenterLng float64         `json:"center_lng"`
	Polygon   json.RawMessage `json:"polygon"` // Lê o JSONB nativamente
	UpdatedAt time.Time       `json:"updated_at"`
	Votes     AreaVotes       `json:"votes"`
}

// ReportWithArea representa o join e o formato final da resposta HTTP
type ReportWithAreaDTO struct {
	ID        uuid.UUID             `json:"id"`
	Latitude  float64               `json:"latitude"`
	Longitude float64               `json:"longitude"`
	Type      string                `json:"type"`
	Status    string                `json:"status"`
	UserID    uuid.UUID             `json:"user_id"`
	Datetime  time.Time             `json:"datetime"`
	Area      ReportAffectedAreaDTO `json:"area"`
}

func IsDomainValidationError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidCoordinates) ||
		errors.Is(err, ErrInvalidLocation) ||
		errors.Is(err, ErrInvalidReportType) ||
		errors.Is(err, ErrInvalidDatetime) ||
		errors.Is(err, ErrInvalidUser)
}

type ReportStatus string

const (
	Created    ReportStatus = "created"         // Estado final (Já pode ser sincronizado)
	InProgress ReportStatus = "inProgress"      // Usuário criou, mas ainda não foi liberado
	Validated  ReportStatus = "RELATO_VALIDADO" // Etapa de validação por outros usuários
	Rejected   ReportStatus = "REJEITADO"
)

type ReportType string

const (
	Landslide    ReportType = "Landslide"     // Deslizamento
	Flooding     ReportType = "Flooding"      // Alagamento
	Thunderstorm ReportType = "Thunderstorm"  // Tempestade
	FireDisaster ReportType = "Fire Disaster" // Incendios
)

type Report struct {
	ID        uuid.UUID
	Latitude  float64
	Longitude float64
	Type      ReportType
	Status    ReportStatus
	UserID    uuid.UUID
	Area      uuid.UUID
	Datetime  time.Time
	// Description não é gravada no banco do report; só é repassada no EVENTS.REPORT.CREATED
	// para o validacao-relato guardar na cópia de votação.
	Description string
}

// Validações futuras
func (r *Report) Validate() error {

	if err := r.isReportTypeValid(); err != nil {
		return err
	}

	if err := r.isUserValid(); err != nil {
		return err
	}

	return nil
}

func (r *Report) isUserValid() error {
	if r.UserID == uuid.Nil {
		return ErrInvalidUser
	}
	return nil
}

func (r *Report) isReportTypeValid() error {
	if r.Type != "Landslide" && r.Type != "Flooding" && r.Type != "Thunderstorm" && r.Type != "Fire Disaster" {
		return ErrInvalidReportType
	}
	return nil
}
