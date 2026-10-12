package delivery

import (
	"context"
	"fmt"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/persistence/postgres/db"
)

type InboxProcessor struct {
	handlers map[string]EventHandler
}

func NewInboxProcessor() *InboxProcessor {
	return &InboxProcessor{
		handlers: make(map[string]EventHandler),
	}
}

func (p *InboxProcessor) RegisterHandler(name string, handler EventHandler) {
	p.handlers[name] = handler
}

func (p *InboxProcessor) Process(ctx context.Context, payload domain.Inbox, work *db.UnitOfWork) error {
	handler, exists := p.handlers[string(payload.EventType)]

	if !exists {
		return fmt.Errorf("event type %s does not exist", payload.EventType)
	}

	return handler.Handle(ctx, payload.Content, work)
}
