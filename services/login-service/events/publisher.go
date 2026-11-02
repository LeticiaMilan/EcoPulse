package events

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const ExchangeName = "ecopulse.events"

type Publisher struct {
	DB  *sql.DB
	URL string
}

type PendingEvent struct {
	ID         string
	EventType  string
	RoutingKey string
	Payload    string
	CreatedAt  time.Time
}

func NewPublisher(db *sql.DB, url string) *Publisher {
	return &Publisher{
		DB:  db,
		URL: url,
	}
}

// Executa continuamente, tentando publicar os eventos pendentes
func (p *Publisher) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		conn, err := amqp.Dial(p.URL)
		if err != nil {
			log.Printf("Erro ao conectar ao RabbitMQ: %v", err)
			if !wait(ctx, 5*time.Second) {
				return
			}
			continue
		}

		channel, err := conn.Channel()
		if err != nil {
			log.Printf("Erro ao criar canal RabbitMQ: %v", err)
			conn.Close()

			if !wait(ctx, 5*time.Second) {
				return
			}
			continue
		}

		err = channel.ExchangeDeclare(
			ExchangeName,
			"topic",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Printf("Erro ao declarar exchange: %v", err)
			channel.Close()
			conn.Close()

			if !wait(ctx, 5*time.Second) {
				return
			}
			continue
		}

		err = channel.Confirm(false)
		if err != nil {
			log.Printf("Erro ao ativar confirmações: %v", err)
			channel.Close()
			conn.Close()

			if !wait(ctx, 5*time.Second) {
				return
			}
			continue
		}

		confirms := channel.NotifyPublish(
			make(chan amqp.Confirmation, 1),
		)

		returns := channel.NotifyReturn(
			make(chan amqp.Return, 1),
		)

		err = p.publishLoop(ctx, channel, confirms, returns)

		channel.Close()
		conn.Close()

		if ctx.Err() != nil {
			return
		}

		if err != nil {
			log.Printf("Publicador desconectado: %v", err)
		}

		if !wait(ctx, 5*time.Second) {
			return
		}
	}
}

// Busca e publica os eventos periodicamente
func (p *Publisher) publishLoop(
	ctx context.Context,
	channel *amqp.Channel,
	confirms <-chan amqp.Confirmation,
	returns <-chan amqp.Return,
) error {

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		if err := p.publishPending(ctx, channel, confirms, returns); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// Busca os eventos ainda não publicados
func (p *Publisher) publishPending(
	ctx context.Context,
	channel *amqp.Channel,
	confirms <-chan amqp.Confirmation,
	returns <-chan amqp.Return,
) error {

	rows, err := p.DB.QueryContext(ctx, `
		SELECT id, event_type, routing_key, payload, created_at
		FROM outbox_events
		WHERE published_at IS NULL
		ORDER BY created_at
		LIMIT 50
	`)
	if err != nil {
		return fmt.Errorf("erro ao buscar eventos: %w", err)
	}

	var pending []PendingEvent

	for rows.Next() {
		var event PendingEvent

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.RoutingKey,
			&event.Payload,
			&event.CreatedAt,
		)
		if err != nil {
			rows.Close()
			return err
		}

		pending = append(pending, event)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}

	rows.Close()

	for _, event := range pending {
		message := amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    event.ID,
			Type:         event.EventType,
			Timestamp:    event.CreatedAt,
			Body:         []byte(event.Payload),
		}

		err := channel.PublishWithContext(
			ctx,
			ExchangeName,
			event.RoutingKey,
			true,
			false,
			message,
		)
		if err != nil {
			return fmt.Errorf("erro ao publicar evento %s: %w", event.ID, err)
		}

		// Aguarda confirmação do RabbitMQ
		for {
			select {
			case returned, ok := <-returns:
				if !ok {
					return fmt.Errorf("canal de mensagens não roteadas foi fechado")
				}

				if returned.MessageId == event.ID {
					return fmt.Errorf(
						"evento %s não encontrou nenhuma fila vinculada à routing key %s",
						event.ID,
						event.RoutingKey,
					)
				}

			case confirmation, ok := <-confirms:
				if !ok {
					return fmt.Errorf("canal de confirmações foi fechado")
				}

				if !confirmation.Ack {
					return fmt.Errorf("RabbitMQ rejeitou o evento %s", event.ID)
				}

				// Verifica se a mensagem foi devolvida por falta de fila
				select {
				case returned := <-returns:
					if returned.MessageId == event.ID {
						return fmt.Errorf(
							"evento %s não foi roteado para nenhuma fila",
							event.ID,
						)
					}
				default:
				}

				goto confirmed

			case <-ctx.Done():
				return ctx.Err()
			}
		}

	confirmed:

		// Marca como publicado somente após confirmação do RabbitMQ
		_, err = p.DB.ExecContext(ctx, `
			UPDATE outbox_events
			SET published_at = NOW()
			WHERE id = $1 AND published_at IS NULL
		`, event.ID)

		if err != nil {
			return fmt.Errorf("erro ao atualizar evento %s: %w", event.ID, err)
		}
	}

	return nil
}

func wait(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
