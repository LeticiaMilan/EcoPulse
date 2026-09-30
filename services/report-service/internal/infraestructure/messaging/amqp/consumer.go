package amqp

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type EventEnvelope struct {
	ID          uuid.UUID `json:"id"`
	Source      string    `json:"source"`
	Type        string    `json:"type"`
	Subject     string    `json:"subject"`
	Time        time.Time `json:"time"`
	ContentType string    `json:"contentType"`
	Data        any       `json:"data"`
}
type Consumer struct {
	Manager *QueueManager
	Fila    string
}

func (c *Consumer) NewConsumer(manager *QueueManager, fila string) {
	c.Manager = manager
	c.Fila = fila
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

	var envelope EventEnvelope

	// PESQUISAR SOBRE PADRÃO ENVELOPE PARA SCHEMA DE DADOS NO AMQP (CloudEvents)
	for msg := range msgs {
		err = json.Unmarshal(msg.Body, &envelope)
		fmt.Println(envelope)
		log.Printf("Received a message: %s", msg.Body)
	}

	return nil
}
