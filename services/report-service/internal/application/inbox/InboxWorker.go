package inbox

import "context"

type InboxWorker struct {
	Processor  *InboxProcessor
	Repository InboxRepository
}

func NewInboxWorker(repository InboxRepository, processor *InboxProcessor) *InboxWorker {
	return &InboxWorker{
		Processor:  processor,
		Repository: repository,
	}
}

func (i *InboxWorker) Process(ctx context.Context) error {
	inbox, _ := i.Repository.GetUnprocessedMessage(ctx)

	// Iniciar transaçao

	if err := i.Processor.Process(ctx, inbox); err != nil {
		return err
	}

	if err := i.Repository.MarkAsProcessed(ctx, inbox.ID); err != nil {
		return err
	}

	// fechar transação
	return nil
}
