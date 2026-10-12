package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"report-service/internal/infraestructure/entrypoint/worker"
	"report-service/internal/ports"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	InboxRepository ports.InboxRepository
	Manager         *QueueManager
	Fila            string
	ExchangeName    string
}

func NewConsumer(manager *QueueManager, fila string, inboxRepository ports.InboxRepository, ExchangeName string) *Consumer {
	return &Consumer{
		Manager:         manager,
		Fila:            fila,
		InboxRepository: inboxRepository,
		ExchangeName:    ExchangeName,
	}
}

func (c *Consumer) Start(ctx context.Context) (*amqp.Channel, <-chan amqp.Delivery, error) {

	ch, err := c.Manager.CreateChannel()
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Channel created")
	err = ch.ExchangeDeclare(c.Fila, c.ExchangeName, true, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Exchange Declared")
	if err := ch.Qos(10, 0, false); err != nil {
		return nil, nil, err
	}
	log.Printf("QoS Defined")
	q, err := ch.QueueDeclare(c.Fila, true, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Queue Declared")
	if err := ch.QueueBind(c.Fila, "#", c.Fila, false, nil); err != nil {
		fmt.Printf(err.Error())
		return nil, nil, err
	}
	log.Printf("Queue Binded")
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}

	log.Print("Consumer started")
	return ch, msgs, nil
}

func (c *Consumer) Consumir(ctx context.Context, msgs <-chan amqp.Delivery) error {
	for msg := range msgs {
		var envelope worker.EventEnvelope
		if err := json.Unmarshal(msg.Body, &envelope); err != nil {
			log.Print("Consumer failed to unmarshal event: " + err.Error())
			if err := msg.Nack(false, false); err != nil {
				log.Print("Consumer failed to nack event: " + err.Error())
			}
			continue
		} else {
			inboxModel := worker.MapInboxEventToApplication(&envelope)
			if err := c.InboxRepository.Save(&ctx, inboxModel); err != nil {
				log.Print("Consumer failed to save event: " + err.Error())
				if err := msg.Nack(false, false); err != nil {
					log.Print("Consumer failed to nack event: " + err.Error())
				}
				continue
			} else {
				if err := msg.Ack(false); err != nil {
					log.Print("Consumer failed to ack event: " + err.Error())

					if err := msg.Nack(false, false); err != nil {
						log.Print("Consumer failed to nack event: " + err.Error())
					}
					continue
				}
			}
		}
	}
	return nil
}
