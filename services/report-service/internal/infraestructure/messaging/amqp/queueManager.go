package amqp

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueManager struct {
	conn *amqp.Connection
}

func (q *QueueManager) NewEventManager(url string) error {
	var err error
	q.conn, err = amqp.Dial(url)

	if err != nil {
		return err
	}

	return nil
}

func (qm *QueueManager) CreateChannel() (*amqp.Channel, error) {
	return qm.conn.Channel()
}
