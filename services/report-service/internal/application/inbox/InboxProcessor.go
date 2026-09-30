package inbox

import (
	"context"
	"fmt"
	"report-service/internal/application"
)

type InboxProcessor struct {
	handlers map[string]application.EventHandler
}

func NewInboxProcessor() *InboxProcessor {
	return &InboxProcessor{
		handlers: make(map[string]application.EventHandler),
	}
}

func (p *InboxProcessor) RegisterHandler(name string, handler application.EventHandler) {
	p.handlers[name] = handler
}

func (p *InboxProcessor) Process(ctx context.Context, payload Inbox) error {
	handler, exists := p.handlers[payload.EventType]
	if !exists {
		return fmt.Errorf("event type %s does not exist", payload.EventType)
	}

	return handler.Handle(ctx, payload.Content)
}
