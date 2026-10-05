package amqp

import (
	"context"
	"encoding/json"
	"log"
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

	log.Print("Consumer started")
	var envelope inbox.EventEnvelope

	for msg := range msgs {
		if err := json.Unmarshal(msg.Body, &envelope); err != nil {
			log.Print("Consumer failed to unmarshal event: " + err.Error())
		} else {

			inboxModel := mappers.MapInboxEventToApplication(envelope)
			ctx := context.Background()

			if err := c.InboxRepository.Save(ctx, inboxModel); err != nil {
				log.Print("Consumer failed to save event: " + err.Error())
			} else {

				if err := msg.Ack(false); err != nil {
					log.Print("Consumer failed to ack event: " + err.Error())
				}
			}
		}
	}

	return nil
}
