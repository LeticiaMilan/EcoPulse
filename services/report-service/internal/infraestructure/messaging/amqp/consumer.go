package amqp

type Consumer struct {
	manager *QueueManager
	fila    string
}

func NewConsumer(manager *QueueManager, fila string) *Consumer {
	return &Consumer{
		manager: manager,
		fila:    fila,
	}
}

func (c *Consumer) Start() error {
	ch, err := c.manager.CreateChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	msgs, err := ch.Consume(c.fila, "topic", false, false, false, false, nil)

	if err != nil {
		return err
	}

	// PESQUISAR SOBRE PADRÃO STRATEGIES PARA ROTEAMENTO DE ACORDO COM TÓPICO
	// PESQUISAR SOBRE PADRÃO ENVELOPE PARA SCHEMA DE DADOS NO AMQP (CloudEvents)

	return nil
}
