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
	Pool    *pgxpool.Pool
	tx      pgx.Tx
	Queries *Queries
}

func NewUnitOfWork(pool *pgxpool.Pool) UnitOfWork {
	return UnitOfWork{
		Pool: pool,
	}
}

func (uow *UnitOfWork) RunInTx(ctx context.Context, fn func(work UnitOfWork) error) error {
	if uow.tx != nil {
		return ErrTransactionAlreadyExists
	}

	if uow.Pool == nil {
		return ErrPoolNotInitialized
	}

	tx, err := uow.Pool.Begin(ctx) // Inicia a transação
	if err != nil {
		return err
	}
	txUow := &UnitOfWork{ // Finaliza instanciação da UOW (Lazy Start)
		Pool:    uow.Pool,
		tx:      tx,
		Queries: New(tx),
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(*txUow); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
