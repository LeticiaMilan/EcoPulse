package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"validacao-relato/src/model"

	"github.com/google/uuid"
)

type RelatoRepository interface {
	CriarComOutbox(ctx context.Context, tx *sql.Tx, relato *model.Relato, eventType string, eventPayload interface{}) error
	BuscarPorID(ctx context.Context, id string) (*model.Relato, error)
	Listar(ctx context.Context, limit, offset int) ([]model.Relato, error)
	AtualizarComOutbox(ctx context.Context, tx *sql.Tx, relato *model.Relato, eventType string, eventPayload interface{}) error
	DeletarComOutbox(ctx context.Context, tx *sql.Tx, id string, eventType string, eventPayload interface{}) error
	ProcessarInbox(ctx context.Context, msgID, eventType string, payload []byte, handler func() error) error
	AdicionarVoto(ctx context.Context, tx *sql.Tx, relatoID string, positivo bool, eventType string, eventPayload interface{}) error
	ProcessarAvaliacoesPeriodicas(ctx context.Context) error
	BuscarPorIDComOutbox(ctx context.Context, id string, eventType string) (*model.Relato, error)
}

type PostgresRelatoRepository struct {
	db *sql.DB
}

func NewPostgresRelatoRepository(db *sql.DB) *PostgresRelatoRepository {
	return &PostgresRelatoRepository{db: db}
}

func (r *PostgresRelatoRepository) CriarComOutbox(ctx context.Context, tx *sql.Tx, relato *model.Relato, eventType string, eventPayload interface{}) error {
	queryRelato := `
		INSERT INTO relatos (id, usuario_id, tipo, descricao, latitude, longitude, estado)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING criado_em, atualizado_em, votos_positivos, votos_negativos`

	err := tx.QueryRowContext(ctx, queryRelato,
		relato.ID,
		relato.UsuarioID,
		relato.Tipo,
		relato.Descricao,
		relato.Latitude,
		relato.Longitude,
		relato.Estado,
	).Scan(
		&relato.CriadoEm,
		&relato.AtualizadoEm,
		&relato.VotosPositivos,
		&relato.VotosNegativos,
	)

	if err != nil {
		return fmt.Errorf("erro ao inserir relato: %w", err)
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload do evento: %w", err)
	}

	outboxID := uuid.New().String()
	queryOutbox := `
		INSERT INTO outbox (id, event_type, status, content, created_at, next_attempt)
		VALUES ($1, $2, 'PENDING', $3, NOW(), NOW())`

	_, err = tx.ExecContext(ctx, queryOutbox, outboxID, eventType, payloadBytes)
	if err != nil {
		return fmt.Errorf("erro ao inserir na outbox: %w", err)
	}

	return nil
}

func (r *PostgresRelatoRepository) BuscarPorID(ctx context.Context, id string) (*model.Relato, error) {
	query := `
		SELECT id, usuario_id, tipo, descricao, latitude, longitude, votos_positivos, votos_negativos, estado, criado_em, proxima_avaliacao_em, atualizado_em 
		FROM relatos 
		WHERE id = $1`

	var relato model.Relato
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&relato.ID,
		&relato.UsuarioID,
		&relato.Tipo,
		&relato.Descricao,
		&relato.Latitude,
		&relato.Longitude,
		&relato.VotosPositivos,
		&relato.VotosNegativos,
		&relato.Estado,
		&relato.CriadoEm,
		&relato.ProximaAvaliacaoEm,
		&relato.AtualizadoEm,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Não encontrado
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar relato por ID: %w", err)
	}

	return &relato, nil
}

func (r *PostgresRelatoRepository) BuscarPorIDComOutbox(ctx context.Context, id string, eventType string) (*model.Relato, error) {
	relato, err := r.BuscarPorID(ctx, id)
	if err != nil {
		return nil, err
	}
	if relato == nil {
		return nil, nil // Não encontrado
	}

	payload, err := json.Marshal(relato)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar relato para o outbox: %w", err)
	}

	outboxID := uuid.New().String()
	queryOutbox := `
		INSERT INTO outbox (id, event_type, status, content, created_at, next_attempt)
		VALUES ($1, $2, 'PENDING', $3, NOW(), NOW())`

	_, err = r.db.ExecContext(ctx, queryOutbox, outboxID, eventType, payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao inserir evento no outbox: %w", err)
	}

	return relato, nil
}

func (r *PostgresRelatoRepository) AdicionarVoto(ctx context.Context, tx *sql.Tx, relatoID string, positivo bool, eventType string, eventPayload interface{}) error {
	var queryRelato string
	if positivo {
		queryRelato = `UPDATE relatos SET votos_positivos = votos_positivos + 1, atualizado_em = NOW() WHERE id = $1`
	} else {
		queryRelato = `UPDATE relatos SET votos_negativos = votos_negativos + 1, atualizado_em = NOW() WHERE id = $1`
	}

	res, err := tx.ExecContext(ctx, queryRelato, relatoID)
	if err != nil {
		return fmt.Errorf("erro ao registrar voto no banco: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("relato não encontrado para votação")
	}

	// Registra o evento na outbox na mesma transação
	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload do evento de voto: %w", err)
	}

	outboxID := uuid.New().String()
	queryOutbox := `
		INSERT INTO outbox (id, event_type, status, content, created_at, next_attempt)
		VALUES ($1, $2, 'PENDING', $3, NOW(), NOW())`

	_, err = tx.ExecContext(ctx, queryOutbox, outboxID, eventType, payloadBytes)
	if err != nil {
		return fmt.Errorf("erro ao inserir evento de voto na outbox: %w", err)
	}

	return nil
}

func (r *PostgresRelatoRepository) Listar(ctx context.Context, limit, offset int) ([]model.Relato, error) {
	query := `
		SELECT id, usuario_id, tipo, descricao, latitude, longitude, votos_positivos, votos_negativos, estado, criado_em, proxima_avaliacao_em, atualizado_em 
		FROM relatos 
		ORDER BY criado_em DESC 
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar relatos: %w", err)
	}
	defer rows.Close()

	var relatos []model.Relato
	for rows.Next() {
		var relato model.Relato
		err := rows.Scan(
			&relato.ID,
			&relato.UsuarioID,
			&relato.Tipo,
			&relato.Descricao,
			&relato.Latitude,
			&relato.Longitude,
			&relato.VotosPositivos,
			&relato.VotosNegativos,
			&relato.Estado,
			&relato.CriadoEm,
			&relato.ProximaAvaliacaoEm,
			&relato.AtualizadoEm,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler linha de relato: %w", err)
		}
		relatos = append(relatos, relato)
	}

	return relatos, nil
}

func (r *PostgresRelatoRepository) AtualizarComOutbox(ctx context.Context, tx *sql.Tx, relato *model.Relato, eventType string, eventPayload interface{}) error {
	queryRelato := `
		UPDATE relatos 
		SET tipo = $1, descricao = $2, latitude = $3, longitude = $4, estado = $5, votos_positivos = $6, votos_negativos = $7, atualizado_em = NOW()
		WHERE id = $8
		RETURNING atualizado_em`

	err := tx.QueryRowContext(ctx, queryRelato,
		relato.Tipo, relato.Descricao, relato.Latitude, relato.Longitude,
		relato.Estado, relato.VotosPositivos, relato.VotosNegativos, relato.ID,
	).Scan(&relato.AtualizadoEm)

	if err != nil {
		return fmt.Errorf("erro ao atualizar relato: %w", err)
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload do evento: %w", err)
	}

	outboxID := uuid.New().String()
	queryOutbox := `
		INSERT INTO outbox (id, event_type, status, content, created_at, next_attempt)
		VALUES ($1, $2, 'PENDING', $3, NOW(), NOW())`

	_, err = tx.ExecContext(ctx, queryOutbox, outboxID, eventType, payloadBytes)
	if err != nil {
		return fmt.Errorf("erro ao inserir na outbox: %w", err)
	}

	return nil
}

func (r *PostgresRelatoRepository) DeletarComOutbox(ctx context.Context, tx *sql.Tx, id string, eventType string, eventPayload interface{}) error {
	queryRelato := `DELETE FROM relatos WHERE id = $1`
	res, err := tx.ExecContext(ctx, queryRelato, id)
	if err != nil {
		return fmt.Errorf("erro ao deletar relato: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("relato não encontrado para exclusão")
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload do evento: %w", err)
	}

	outboxID := uuid.New().String()
	queryOutbox := `
		INSERT INTO outbox (id, event_type, status, content, created_at, next_attempt)
		VALUES ($1, $2, 'PENDING', $3, NOW(), NOW())`

	_, err = tx.ExecContext(ctx, queryOutbox, outboxID, eventType, payloadBytes)
	if err != nil {
		return fmt.Errorf("erro ao inserir na outbox: %w", err)
	}

	return nil
}

func (r *PostgresRelatoRepository) ProcessarInbox(ctx context.Context, msgID, eventType string, payload []byte, handler func() error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Tenta verificar se a mensagem já foi processada (Idempotência)
	queryCheck := `SELECT status FROM inbox WHERE id = $1 FOR UPDATE`
	var status string
	err = tx.QueryRowContext(ctx, queryCheck, msgID).Scan(&status)

	if err == sql.ErrNoRows {
		// Mensagem nova, insere como PROCESSING
		insertQuery := `INSERT INTO inbox (id, event_type, status, content, received_at, next_attempt) VALUES ($1, $2, 'PROCESSING', $3, NOW(), NOW())`
		_, err = tx.ExecContext(ctx, insertQuery, msgID, eventType, payload)
		if err != nil {
			return err
		}
	} else if err == nil {
		if status == "PROCESSED" {
			// Já processada anteriormente, ignora (idempotente)
			return nil
		}
	} else {
		return err
	}

	// Executa a regra de negócio associada ao evento
	if err := handler(); err != nil {
		_, _ = tx.ExecContext(ctx, `UPDATE inbox SET status = 'FAILED', retries = retries + 1 WHERE id = $1`, msgID)
		return err
	}

	// Marca como processada com sucesso
	_, err = tx.ExecContext(ctx, `UPDATE inbox SET status = 'PROCESSED', processed_at = NOW() WHERE id = $1`, msgID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRelatoRepository) ProcessarAvaliacoesPeriodicas(ctx context.Context) error {
	// 1. Regra para POSSIVEL_RELATO:
	// Se votos positivos > negativos -> vira RELATO_VALIDADO e ganha +1 hora
	// Se não (positivos <= negativos) -> vira REJEITADO
	queryPossivel := `
		UPDATE relatos 
		SET estado = CASE
				WHEN votos_positivos > votos_negativos THEN 'RELATO_VALIDADO'::estado_relato 
				ELSE 'REJEITADO'::estado_relato 
			END,
			votos_positivos = 0, 
			votos_negativos = 0,
			proxima_avaliacao_em = CASE 
				WHEN votos_positivos > votos_negativos THEN NOW() + INTERVAL '1 hour' 
				ELSE proxima_avaliacao_em 
			END,
			atualizado_em = NOW()
		WHERE estado = 'POSSIVEL_RELATO' AND proxima_avaliacao_em <= NOW()`

	_, err := r.db.ExecContext(ctx, queryPossivel)
	if err != nil {
		return fmt.Errorf("erro ao processar possíveis relatos: %w", err)
	}

	// 2. Regra para RELATO_VALIDADO:
	// Se o tempo chegou, volta para POSSIVEL_RELATO e define novo tempo para daqui a 15 minutos
	queryValidado := `
		UPDATE relatos 
		SET estado = 'POSSIVEL_RELATO'::estado_relato,
			proxima_avaliacao_em = NOW() + INTERVAL '15 minutes',
			atualizado_em = NOW()
		WHERE estado = 'RELATO_VALIDADO' AND proxima_avaliacao_em <= NOW()`

	_, err = r.db.ExecContext(ctx, queryValidado)
	if err != nil {
		return fmt.Errorf("erro ao reavaliar relatos validados: %w", err)
	}

	return nil
}
