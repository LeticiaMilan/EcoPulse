package worker

import (
	"context"
	"report-service/internal/infraestructure/entrypoint/delivery"
	"report-service/internal/infraestructure/persistence/postgres/db"
	"report-service/internal/ports"
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

func (i *InboxWorker) Process(ctx *context.Context) error {
	return i.UnitOfWork.RunInTx(*ctx, func(work db.UnitOfWork) error {
		inbox, _ := i.Repository.GetUnprocessedMessage(ctx)

		if err := i.Processor.Process(*ctx, *inbox); err != nil {
			return err
		}
		return nil
	})
	return nil
}
