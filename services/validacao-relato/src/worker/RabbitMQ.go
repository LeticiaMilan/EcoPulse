package queue

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConn struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQConn(url string) (*RabbitMQConn, error) {
	// 1. Conecta ao servidor RabbitMQ
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no RabbitMQ: %w", err)
	}

	// 2. Abre um canal de comunicação
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("falha ao abrir canal no RabbitMQ: %w", err)
	}

	// 3. Declara a Exchange do tipo 'topic'
	// O tipo 'topic' permite roteamento flexível usando padrões (wildcards)
	err = ch.ExchangeDeclare(
		"validacao_relatos_exchange", // Nome da exchange
		"topic",                      // Tipo topic
		true,                         // Durable (sobrevive a reinicializações do RabbitMQ)
		false,                        // Auto-deleted
		false,                        // Internal
		false,                        // No-wait
		nil,                          // Arguments
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao declarar exchange: %w", err)
	}

	filaValidacao := "validacao_relatos_queue"
	_, err = ch.QueueDeclare(filaValidacao, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	err = ch.QueueBind(filaValidacao, "validacaorelato.*", "validacao_relatos_exchange", false, nil)
	if err != nil {
		return nil, err
	}

	filaResposta := "validacao_resposta_queue"
	_, err = ch.QueueDeclare(filaResposta, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao declarar nova fila: %w", err)
	}
	err = ch.QueueBind(
		filaResposta,
		"validacaoresposta.*", // Padrão que vai cair nesta nova fila
		"validacao_relatos_exchange",
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("falha no bind da nova fila: %w", err)
	}

	return &RabbitMQConn{Conn: conn, Channel: ch}, nil
}

func (r *RabbitMQConn) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Conn != nil {
		r.Conn.Close()
	}
}
