package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"uuid"
	"validacao-relato/src/model"
	"validacao-relato/src/repository"
	"validacao-relato/src/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

type InboxConsumer struct {
	channel *amqp.Channel
	repo    repository.RelatoRepository
	svc     *service.RelatoService
}

func NewInboxConsumer(channel *amqp.Channel, repo repository.RelatoRepository, svc *service.RelatoService) *InboxConsumer {
	return &InboxConsumer{channel: channel, repo: repo, svc: svc}
}

func (c *InboxConsumer) StartConsuming(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		"validacao_relatos_queue", // nome da fila
		"inbox_worker",            // consumer tag
		false,                     // auto-ack (false para controlarmos manualmente com Ack/Nack)
		false,                     // exclusive
		false,                     // no-local
		false,                     // no-wait
		nil,                       // args
	)
	if err != nil {
		return err
	}

	log.Println("Consumidor do RabbitMQ (Inbox) iniciado com sucesso. Aguardando mensagens...")

	go func() {
		for d := range msgs {
			// Extraímos ou geramos um ID único para a mensagem recebida.
			// O ideal é que o remetente envie o ID no MessageId do AMQP ou dentro do JSON.
			// No internal/queue/inbox_consumer.go, ajuste a obtenção do msgID:
			msgID := d.MessageId
			if msgID == "" {
				msgID = d.CorrelationId
				if msgID == "" {
					msgID = uuid.New().String()
				}
			}

			eventType := d.RoutingKey
			payload := d.Body

			// Definimos a regra de negócio que será executada de forma idempotente
			handler := func() error {
				switch eventType {
				case "validacaorelato.criar":
					var req struct {
						ID        string  `json:"id"`
						UsuarioID string  `json:"usuario_id"`
						Tipo      string  `json:"tipo"`
						Descricao string  `json:"descricao"`
						Latitude  float64 `json:"latitude"`
						Longitude float64 `json:"longitude"`
					}

					if err := json.Unmarshal(payload, &req); err != nil {
						return fmt.Errorf("erro ao decodificar JSON de criação: %w", err)
					}

					_, err := c.svc.CriarRelato(ctx, req.ID, req.UsuarioID, req.Tipo, req.Descricao, req.Latitude, req.Longitude)
					if err != nil {
						return err
					}

				case "validacaorelato.votar":
					var req struct {
						ID   string `json:"id"`
						Tipo string `json:"tipo"`
					}
					if err := json.Unmarshal(payload, &req); err != nil {
						return fmt.Errorf("erro ao decodificar JSON de voto: %w", err)
					}

					positivo := false
					if req.Tipo == "positivo" {
						positivo = true
					}

					err := c.svc.VotarRelato(ctx, req.ID, positivo)
					if err != nil {
						return err
					}

				case "validacaorelato.atualizar":
					var req struct {
						ID        string             `json:"id"`
						Tipo      string             `json:"tipo"`
						Descricao string             `json:"descricao"`
						Latitude  float64            `json:"latitude"`
						Longitude float64            `json:"longitude"`
						Estado    model.EstadoRelato `json:"estado"`
					}
					if err := json.Unmarshal(payload, &req); err != nil {
						return fmt.Errorf("erro ao decodificar JSON de atualização: %w", err)
					}

					_, err := c.svc.AtualizarRelato(ctx, req.ID, req.Tipo, req.Descricao, req.Latitude, req.Longitude, req.Estado)
					if err != nil {
						return err
					}

				case "validacaorelato.deletar":
					var req struct {
						ID string `json:"id"`
					}
					if err := json.Unmarshal(payload, &req); err != nil {
						return fmt.Errorf("erro ao decodificar JSON de deleção: %w", err)
					}

					err := c.svc.DeletarRelato(ctx, req.ID)
					if err != nil {
						return err
					}

				case "validacaorelato.buscar":
					var req struct {
						ID string `json:"id"`
					}
					if err := json.Unmarshal(payload, &req); err != nil {
						return fmt.Errorf("erro ao decodificar JSON de busca: %w", err)
					}

					_, err := c.svc.BuscarPorID(ctx, req.ID)
					if err != nil {
						return err
					}
				}

				return nil
			}

			// Processa utilizando o Inbox Pattern no banco de dados
			err := c.repo.ProcessarInbox(ctx, msgID, eventType, payload, handler)
			if err != nil {
				log.Printf("[INBOX] Erro ao processar mensagem %s: %v", msgID, err)
				// Rejeita a mensagem e coloca de volta na fila (requeue = true) ou manda para DLQ
				_ = d.Nack(false, false)
				continue
			}

			// Confirma o sucesso da mensagem para o RabbitMQ (Ack)
			_ = d.Ack(false)
		}
	}()

	return nil
}
