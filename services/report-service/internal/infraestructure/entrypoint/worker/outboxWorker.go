package worker

import (
	"context"
	"log"
	"report-service/internal/ports"
	"time"
)

type OutboxWorker struct {
	Repository ports.OutboxRepository
	canal      chan<- EventEnvelope
}

func NewOutboxWorker(repository ports.OutboxRepository, canal chan<- EventEnvelope) *OutboxWorker {
	return &OutboxWorker{Repository: repository, canal: canal}
}

func (o *OutboxWorker) Process(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {

		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			log.Printf("Processing outbox")
			if err := o.run(ctx); err != nil {
				return err
			}
		}
	}
}

func (o *OutboxWorker) run(ctx context.Context) error {
	tx, err := o.Repository.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	txCtx := context.WithValue(ctx, "tx", tx)

	outbox, err := o.Repository.GetUnprocessedMessage(txCtx)
	if outbox == nil {
	} else {
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
		eventEnvelope := mapOutboxEventToEnvelope(outbox)
		eventEnvelope.Tx = tx
		o.canal <- eventEnvelope
		close(o.canal)
	}
	return nil
}
