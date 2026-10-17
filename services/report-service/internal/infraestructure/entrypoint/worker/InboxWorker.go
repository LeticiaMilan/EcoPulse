package worker

import (
	"context"
	"errors"
	"log"
	"report-service/internal/infraestructure/entrypoint/delivery"
	"report-service/internal/ports"
	"time"

	"github.com/jackc/pgx/v5"
)

type InboxWorker struct {
	Processor  *delivery.InboxProcessor
	Repository ports.InboxRepository
}

func NewInboxWorker(repository ports.InboxRepository, processor *delivery.InboxProcessor) *InboxWorker {
	return &InboxWorker{
		Processor:  processor,
		Repository: repository,
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
				return err
			}
		}
	}
}

func (i *InboxWorker) Run(ctx context.Context) error {

	tx, err := i.Repository.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	var executando bool
	defer func() {
		if executando {
			if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				log.Print("Error while rolling back transaction")
			} else {
				log.Println("Transaction rollback")
			}
		}
	}()

	txCtx := context.WithValue(ctx, "tx", tx)
	inbox, _ := i.Repository.GetUnprocessedMessage(txCtx)
	if inbox == nil {
	} else {
		executando = true
		if err := i.Processor.Process(txCtx, *inbox); err != nil {
			log.Printf("[Worker] Process err: %+v", err)
			return err
		}
		log.Println("Done processing inbox")

		if err := tx.Commit(ctx); err != nil {
			log.Printf("[Worker] Commit err: %+v", err)
		}
		log.Println("Transaction commit")
	}
	return nil
}
