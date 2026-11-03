package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"validacao-relato/src/model"
	"validacao-relato/src/repository"
)

type RelatoService struct {
	db   *sql.DB
	repo repository.RelatoRepository
}

func NewRelatoService(db *sql.DB, repo repository.RelatoRepository) *RelatoService {
	return &RelatoService{db: db, repo: repo}
}

type RelatoEvent struct {
	ID        string             `json:"id"`
	UsuarioID string             `json:"usuario_id"`
	Tipo      string             `json:"tipo"`
	Descricao string             `json:"descricao,omitempty"`
	Latitude  float64            `json:"latitude,omitempty"`
	Longitude float64            `json:"longitude,omitempty"`
	Estado    model.EstadoRelato `json:"estado,omitempty"`
}

func (s *RelatoService) CriarRelato(ctx context.Context, ID, usuarioID, tipo, descricao string, lat, lon float64) (*model.Relato, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	relato := &model.Relato{
		ID:        ID,
		UsuarioID: usuarioID,
		Tipo:      tipo,
		Descricao: descricao,
		Latitude:  lat,
		Longitude: lon,
		Estado:    model.PossivelRelato,
	}

	err = s.repo.CriarComOutbox(ctx, tx, relato, "EVENTS.VALIDATIONRESPONSE.CREATED", RelatoEvent{
		ID:        relato.ID,
		UsuarioID: usuarioID,
		Tipo:      tipo,
		Descricao: descricao,
		Latitude:  lat,
		Longitude: lon,
		Estado:    relato.Estado,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return relato, nil
}

type VotoEvent struct {
	ID       string `json:"id"`
	Positivo bool   `json:"positivo"`
}

func (s *RelatoService) VotarRelato(ctx context.Context, relatoID string, positivo bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Verifica se o relato existe antes de votar
	relato, err := s.repo.BuscarPorID(ctx, relatoID)
	if err != nil || relato == nil {
		return errors.New("relato não encontrado")
	}

	eventPayload := VotoEvent{
		ID:       relatoID,
		Positivo: positivo,
	}

	err = s.repo.AdicionarVoto(ctx, tx, relatoID, positivo, "EVENTS.VALIDATIONRESPONSE.VOTED", eventPayload)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *RelatoService) BuscarPorID(ctx context.Context, id string) (*model.Relato, error) {
	relato, err := s.repo.BuscarPorIDComOutbox(ctx, id, "EVENTS.VALIDATIONRESPONSE.REPORT")
	if err != nil {
		return nil, fmt.Errorf("erro no service ao buscar relato por ID com outbox: %w", err)
	}

	if relato == nil {
		return nil, nil // Relato não encontrado
	}

	return relato, nil
}

func (s *RelatoService) ListarRelatos(ctx context.Context, limit, offset int) ([]model.Relato, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.Listar(ctx, limit, offset)
}

func (s *RelatoService) AtualizarRelato(ctx context.Context, id, tipo, descricao string, lat, lon float64, estado model.EstadoRelato) (*model.Relato, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	relato, err := s.repo.BuscarPorID(ctx, id)
	if err != nil || relato == nil {
		return nil, errors.New("relato não encontrado")
	}

	relato.Tipo = tipo
	relato.Descricao = descricao
	relato.Latitude = lat
	relato.Longitude = lon
	relato.Estado = estado

	err = s.repo.AtualizarComOutbox(ctx, tx, relato, "EVENTS.VALIDATIONRESPONSE.UPDATED", RelatoEvent{
		ID:        relato.ID,
		UsuarioID: relato.UsuarioID,
		Tipo:      relato.Tipo,
		Descricao: relato.Descricao,
		Latitude:  relato.Latitude,
		Longitude: relato.Longitude,
		Estado:    relato.Estado,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return relato, nil
}

func (s *RelatoService) DeletarRelato(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	relato, err := s.repo.BuscarPorID(ctx, id)
	if err != nil || relato == nil {
		return errors.New("relato não encontrado")
	}

	err = s.repo.DeletarComOutbox(ctx, tx, id, "EVENTS.VALIDATIONRESPONSE.DELETED", RelatoEvent{
		ID:        relato.ID,
		UsuarioID: relato.UsuarioID,
		Tipo:      relato.Tipo,
		Descricao: relato.Descricao,
		Latitude:  relato.Latitude,
		Longitude: relato.Longitude,
		Estado:    relato.Estado,
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}
