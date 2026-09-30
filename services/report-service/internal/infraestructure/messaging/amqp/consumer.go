package amqp

import (
	"context"
	"encoding/json"
	"report-service/internal/application/inbox"
	"report-service/internal/infraestructure/persistence/postgres/mappers"
)

type Consumer struct {
	InboxRepository inbox.InboxRepository
	Manager         *QueueManager
	Fila            string
}

func NewConsumer(manager *QueueManager, fila string, inboxRepository inbox.InboxRepository) *Consumer {
	return &Consumer{
		Manager:         manager,
		Fila:            fila,
		InboxRepository: inboxRepository,
	}
}

func (c *Consumer) Start() error {
	ch, err := c.Manager.CreateChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(c.Fila, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(c.Fila, true, false, false, false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	var envelope inbox.EventEnvelope

	for msg := range msgs {
		if err := json.Unmarshal(msg.Body, &envelope); err != nil {
			return err
		}

		inboxModel := mappers.MapInboxEventToApplication(envelope)
		ctx := context.Background()

		if err := c.InboxRepository.Save(ctx, inboxModel); err != nil {
			return err
		}
		if err := msg.Ack(false); err != nil {
			return err
		}
	}

	return nil
}
