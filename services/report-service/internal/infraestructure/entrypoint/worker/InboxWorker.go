package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/entrypoint/delivery"
	"report-service/internal/ports"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type inboxDTO struct {
	ID        uuid.UUID       `json:"id"`
	EventType string          `json:"eventType"`
	Content   json.RawMessage `json:"content"`
	Erro      string          `json:"error"`
}
type InboxWorker struct {
	Processor        *delivery.InboxProcessor
	Repository       ports.InboxRepository
	OutboxRepository ports.OutboxRepository
}

func NewInboxWorker(repository ports.InboxRepository, processor *delivery.InboxProcessor, OutboxRepository ports.OutboxRepository) *InboxWorker {
	return &InboxWorker{
		Processor:        processor,
		Repository:       repository,
		OutboxRepository: OutboxRepository,
	}
}

func (i *InboxWorker) Process(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			log.Printf("Processing inbox")
			if err := i.Run(ctx); err != nil {
				log.Printf("Error processing inbox: %v", err)
			}
		}
	}
}

func (i *InboxWorker) Run(ctx context.Context) error {

	tx, err := i.Repository.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	var inbox *domain.Inbox
	var Erro error = nil
	defer func() {
		if domain.IsDomainValidationError(Erro) {
			inboxDTO := inboxDTO{
				ID:        inbox.ID,
				EventType: string(inbox.EventType),
				Content:   inbox.Content,
				Erro:      Erro.Error(),
			}
			content, err := json.Marshal(inboxDTO)
			if err == nil {
				switch string(inbox.EventType) {
				case string(domain.InboxGeneratedReportEvent):
					err = i.OutboxRepository.Save(ctx, domain.NewOutbox(domain.OutboxReportGeneratedError, content))
					if err != nil {
						log.Printf("Error while saving outbox: %v", err)
					}
				case string(domain.InboxValidatedReportEvent):
					err = i.OutboxRepository.Save(ctx, domain.NewOutbox(domain.OutboxReportValidatedError, content))
					if err != nil {
						log.Printf("Error while saving outbox: %v", err)
					}
				case string(domain.InboxRejectedReportEvent):
					err = i.OutboxRepository.Save(ctx, domain.NewOutbox(domain.OutboxReportRejectedError, content))
					if err != nil {
						log.Printf("Error while saving outbox: %v", err)
					}
				case string(domain.InboxAreaDefinedReportEvent):
					err = i.OutboxRepository.Save(ctx, domain.NewOutbox(domain.OutboxReportDefinedAreaError, content))
					if err != nil {
						log.Printf("Error while saving outbox: %v", err)
					}
				}
			}
			Erro = tx.Commit(ctx)
			if Erro != nil {
				log.Printf("[Worker] Commit err: %+v", err)
			}
			log.Println("Transaction commit")
		} else {
			if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				log.Print("Error while rolling back transaction")
			} else {
				if Erro != nil {
					log.Println("Transaction rollback")
					retry := inbox.Retries + 1
					seconds := math.Pow(2, float64(retry)) * 2
					inbox.Retries = retry
					inbox.Next_attempt = time.Now().Add(time.Duration(seconds) * time.Second)
					erro := i.Repository.UpdateInboxRetry(ctx, *inbox)
					if erro != nil {
						log.Printf("Error while updating inbox retry: %v", erro)
					}
				}

			}
		}
	}()

	txCtx := context.WithValue(ctx, "tx", tx)
	inbox, _ = i.Repository.GetUnprocessedMessage(txCtx)
	if inbox == nil {
	} else {

		Erro = i.Processor.Process(txCtx, *inbox)
		if Erro != nil {
			log.Printf("[Worker] Process err: %+v", Erro)
			return Erro
		}

		Erro = tx.Commit(ctx)
		if Erro != nil {
			log.Printf("[Worker] Commit err: %+v", err)
		}
		log.Println("Transaction commit")
	}
	return nil
}
