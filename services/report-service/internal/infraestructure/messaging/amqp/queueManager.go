package amqp

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueManager struct {
	conn *amqp.Connection
}

func NewEventManager(url string) (*QueueManager, error) {
	connection, err := amqp.Dial(url)

	if err != nil {
		return nil, err
	}

	return &QueueManager{
		conn: connection,
	}, nil
}

func (qm *QueueManager) CreateChannel() (*amqp.Channel, error) {
	return qm.conn.Channel()
}
