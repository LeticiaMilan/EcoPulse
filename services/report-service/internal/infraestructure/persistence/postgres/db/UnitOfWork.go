package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTransactionAlreadyExists = errors.New("transaction already exists")
	ErrPoolNotInitialized       = errors.New("pool not initialized")
)

type UnitOfWork struct {
	pool    *pgxpool.Pool
	tx      pgx.Tx
	queries *Queries
}

func NewUnitOfWork(pool *pgxpool.Pool) UnitOfWork {
	return UnitOfWork{
		pool: pool,
	}
}

func (uow *UnitOfWork) RunInTx(ctx context.Context, fn func(work UnitOfWork) error) error {
	if uow.tx != nil {
		return ErrTransactionAlreadyExists
	}

	if uow.pool == nil {
		return ErrPoolNotInitialized
	}

	tx, err := uow.pool.Begin(ctx) // Inicia a transação
	if err != nil {
		return err
	}
	txUow := &UnitOfWork{ // Finaliza instanciação da UOW (Lazy Start)
		pool:    uow.pool,
		tx:      tx,
		queries: New(uow.pool),
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(*txUow); err != nil {
		if err = tx.Rollback(ctx); err != nil {
			return err
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
