package model

import "time"

type EstadoRelato string

const (
	PossivelRelato EstadoRelato = "POSSIVEL_RELATO"
	RelatoValidado EstadoRelato = "RELATO_VALIDADO"
	Expirado       EstadoRelato = "EXPIRADO"
	Rejeitado      EstadoRelato = "REJEITADO"
)

type Relato struct {
	ID                 string       `json:"id" db:"id"` // <-- Alterado de int para string
	UsuarioID          string       `json:"usuario_id" db:"usuario_id"`
	Tipo               string       `json:"tipo" db:"tipo"`
	Descricao          string       `json:"descricao" db:"descricao"`
	Latitude           float64      `json:"latitude" db:"latitude"`
	Longitude          float64      `json:"longitude" db:"longitude"`
	VotosPositivos     int          `json:"votos_positivos" db:"votos_positivos"`
	VotosNegativos     int          `json:"votos_negativos" db:"votos_negativos"`
	Estado             EstadoRelato `json:"estado" db:"estado"`
	CriadoEm           time.Time    `json:"criado_em" db:"criado_em"`
	ProximaAvaliacaoEm time.Time    `json:"proxima_avaliacao_em" db:"proxima_avaliacao_em"`
	AtualizadoEm       time.Time    `json:"atualizado_em" db:"atualizado_em"`
}
