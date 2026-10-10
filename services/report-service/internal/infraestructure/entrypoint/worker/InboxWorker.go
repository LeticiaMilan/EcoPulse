package worker

import (
	"context"
	"report-service/internal/infraestructure/entrypoint/delivery"
	"report-service/internal/ports"
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

func (i *InboxWorker) Process(ctx *context.Context) error {
	inbox, _ := i.Repository.GetUnprocessedMessage(ctx)

	// Iniciar transaçao

	if err := i.Processor.Process(*ctx, inbox); err != nil {
		return err
	}

	if err := i.Repository.MarkAsProcessed(ctx, inbox.ID); err != nil {
		return err
	}

	// fechar transação
	return nil
}
