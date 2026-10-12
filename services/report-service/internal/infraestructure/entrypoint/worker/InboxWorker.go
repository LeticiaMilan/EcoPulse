package worker

import (
	"context"
	"log"
	"report-service/internal/infraestructure/entrypoint/delivery"
	"report-service/internal/infraestructure/persistence/postgres/db"
	"report-service/internal/infraestructure/persistence/postgres/repository"
	"report-service/internal/ports"
	"time"
)

type InboxWorker struct {
	Processor  *delivery.InboxProcessor
	Repository ports.InboxRepository
	UnitOfWork db.UnitOfWork
}

func NewInboxWorker(repository ports.InboxRepository, processor *delivery.InboxProcessor, UnitOfWork db.UnitOfWork) *InboxWorker {
	return &InboxWorker{
		Processor:  processor,
		Repository: repository,
		UnitOfWork: UnitOfWork,
	}
}

func (i *InboxWorker) Process(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {

		return i.UnitOfWork.RunInTx(ctx, func(work db.UnitOfWork) error {
			inboxRepository := repository.NewInboxRepository(work.Pool, work.Queries)

			inbox, _ := inboxRepository.GetUnprocessedMessage(&ctx)
			if inbox == nil {
			} else {
				if err := i.Processor.Process(ctx, *inbox, &work); err != nil {
					log.Printf("[Worker] Process err: %+v", err)
					return err
				}
			}
			return nil
		})
	}
	return nil
}
