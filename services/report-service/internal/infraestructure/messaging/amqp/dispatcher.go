package amqp

import (
	"context"
	"encoding/json"
	"log"
	"report-service/internal/ports"
	"time"

	"report-service/internal/infraestructure/entrypoint/worker"

	"github.com/rabbitmq/amqp091-go"
)

type Dispatcher struct {
	OutboxRepository ports.OutboxRepository
	Manager          *QueueManager
	ExchangeName     string
	ExchangeKind     string
}

func NewDispatcher(manager *QueueManager, outboxRepository ports.OutboxRepository, ExchangeName string, ExchangeKind string) *Dispatcher {
	return &Dispatcher{
		OutboxRepository: outboxRepository,
		Manager:          manager,
		ExchangeName:     ExchangeName,
		ExchangeKind:     ExchangeKind,
	}
}

func (d *Dispatcher) Start(ctx context.Context, canal <-chan worker.EventEnvelope) error {
	ch, err := d.Manager.CreateChannel()
	if err != nil {
		return err
	}
	err = ch.ExchangeDeclare(d.ExchangeName, d.ExchangeKind, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		return err
	}

	go func() {
		defer ch.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-canal:
				defer func() {
					if msg.Tx != nil {
						_ = msg.Tx.Rollback(ctx)
					}
				}()
				txCtx := context.WithValue(ctx, "tx", msg.Tx)

				if !ok {
					return
				}
				encoded, err := json.Marshal(msg)
				if err != nil {
					msg.Tx.Rollback(ctx)
					msg.Tx = nil
					log.Printf("Failed to encode canal message to JSON, %v", err)
					continue
				}
				pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second) // Aumentado para dar tempo do ACK

				log.Printf("Publishing canal message to queue, %v", string(encoded))

				errPublish := ch.PublishWithContext(pubCtx, d.ExchangeName, msg.RoutingKey, false, false,
					amqp091.Publishing{
						ContentType: "application/json",
						Body:        encoded,
					})

				if errPublish != nil {
					msg.Tx.Rollback(ctx)
					msg.Tx = nil
					cancel()
					log.Printf("Error publishing message to outbox channel: %s", err)
					continue
				}
				log.Printf("Published message to outbox channel, %v")
				cancel()

				d.OutboxRepository.MarkAsProcessed(txCtx, msg.ID)
				err = msg.Tx.Commit(ctx)

				if err != nil {
					msg.Tx.Rollback(ctx)
					msg.Tx = nil
					log.Printf("Error committing message to outbox channel: %s", err)
				}
				msg.Tx = nil
			}
		}
	}()
	return nil
}
