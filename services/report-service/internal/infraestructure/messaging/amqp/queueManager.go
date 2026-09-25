package amqp

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueManager struct {
	conn *amqp.Connection
}

func NewQueueManager(url string) (*QueueManager, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return &QueueManager{
		conn: conn,
	}, nil
}

func (qm *QueueManager) CreateChannel() (*amqp.Channel, error) {
	return qm.conn.Channel()
}
