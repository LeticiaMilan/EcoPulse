package queue

import (
	"context"
	"database/sql"
	"log"
	"time"
	"validacao-relato/src/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

type OutboxWorker struct {
	db      *sql.DB
	channel *amqp.Channel
}

func NewOutboxWorker(db *sql.DB, ch *amqp.Channel) *OutboxWorker {
	return &OutboxWorker{db: db, channel: ch}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.dispatchPendingMessages(ctx)
		}
	}
}

func (w *OutboxWorker) dispatchPendingMessages(ctx context.Context) {
	rows, err := w.db.QueryContext(ctx, `
		SELECT id, event_type, content, retries 
		FROM outbox 
		WHERE status = 'PENDING' AND next_attempt <= NOW() 
		LIMIT 50 FOR UPDATE SKIP LOCKED`)
	if err != nil {
		log.Printf("Erro ao buscar mensagens na outbox: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var msg model.OutboxMessage
		if err := rows.Scan(&msg.ID, &msg.EventType, &msg.Content, &msg.Retries); err != nil {
			continue
		}

		// Publica no RabbitMQ
		err = w.channel.PublishWithContext(
			ctx,
			"exchange.ecopulse", // exchange
			msg.EventType,       // routing key
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        msg.Content,
				MessageId:   msg.ID,
			},
		)

		if err != nil {
			log.Printf("Erro ao publicar outbox ID %s no RabbitMQ: %v", msg.ID, err)
			_, _ = w.db.ExecContext(ctx, `
				UPDATE outbox 
				SET retries = retries + 1, next_attempt = NOW() + INTERVAL '30 seconds' 
				WHERE id = $1`, msg.ID)
			continue
		}

		// Marca como processado/enviado
		_, _ = w.db.ExecContext(ctx, `
			UPDATE outbox 
			SET status = 'SENT', processed_at = NOW() 
			WHERE id = $1`, msg.ID)
	}
}
