package queue

import (
	"context"
	"log"
	"time"
	"validacao-relato/src/repository"
)

type AvaliacaoWorker struct {
	repo repository.RelatoRepository
}

func NewAvaliacaoWorker(repo repository.RelatoRepository) *AvaliacaoWorker {
	return &AvaliacaoWorker{repo: repo}
}

func (w *AvaliacaoWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Worker de avaliação periódica de relatos iniciado.")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker de avaliação encerrado.")
			return
		case <-ticker.C:
			err := w.repo.ProcessarAvaliacoesPeriodicas(ctx)
			if err != nil {
				log.Printf("Erro ao executar avaliações periódicas: %v", err)
			}
		}
	}
}
