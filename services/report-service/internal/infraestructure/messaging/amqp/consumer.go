package amqp

import (
	"context"
	"encoding/json"
	"log"
	"report-service/internal/infraestructure/entrypoint/worker"
	"report-service/internal/ports"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	InboxRepository ports.InboxRepository
	Manager         *QueueManager
	QueueName       string
	ExchangeName    string
	ExchangeKind    string
	RoutingKeys     []string
}

func NewConsumer(manager *QueueManager, inboxRepository ports.InboxRepository, ExchangeName string, ExchangeKind string, QueueName string, RoutingKeys []string) *Consumer {
	return &Consumer{
		Manager:         manager,
		ExchangeName:    ExchangeName,
		InboxRepository: inboxRepository,
		ExchangeKind:    ExchangeKind,
		QueueName:       QueueName,
		RoutingKeys:     RoutingKeys,
	}
}

func (c *Consumer) Start() (*amqp.Channel, <-chan amqp.Delivery, error) {

	ch, err := c.Manager.CreateChannel()
	if err != nil {
		return nil, nil, err
	}

	err = ch.ExchangeDeclare(c.ExchangeName, c.ExchangeKind, true, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}

	if err := ch.Qos(10, 0, false); err != nil {
		return nil, nil, err
	}

	q, err := ch.QueueDeclare(c.QueueName, true, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}

	for _, RoutingKey := range c.RoutingKeys {
		if err := ch.QueueBind(c.QueueName, RoutingKey, c.ExchangeName, false, nil); err != nil {
			return nil, nil, err
		}
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}

	return ch, msgs, nil
}

func (c *Consumer) Consumir(ctx context.Context, msgs <-chan amqp.Delivery) error {
	go func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case msg, ok := <-msgs:
				if !ok {
					log.Printf("Consumer channel closed")
				}
				log.Printf("Received message: %s", msg.Body)
				var envelope worker.EventEnvelope
				if err := json.Unmarshal(msg.Body, &envelope); err != nil {
					log.Print("Consumer failed to unmarshal event: " + err.Error())
					_ = msg.Nack(false, false)
					continue
				}

				envelope.RoutingKey = msg.RoutingKey
				inboxModel := worker.MapInboxEventToApplication(envelope)
				log.Printf("Mapped inbox model: %+v", inboxModel)
				if err := c.InboxRepository.Save(ctx, inboxModel); err != nil {
					log.Print("Consumer failed to save event: " + err.Error())
					_ = msg.Nack(false, false)
					continue
				}
				if err := msg.Ack(false); err != nil {
					log.Print("Consumer failed to ack event: " + err.Error())
					continue
				}
			}
		}
	}()

	return nil
}
