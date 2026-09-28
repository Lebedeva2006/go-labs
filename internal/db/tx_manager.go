package db

import (
 "context"
 "fmt"

 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}


type TxManager interface {
 Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type pgxTxManager struct {
 pool *pgxpool.Pool
}


func NewTxManager(pool *pgxpool.Pool) TxManager {
 return &pgxTxManager{pool: pool}
}



func (m *pgxTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
 if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {

  return fn(ctx)
 }

 tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{
  IsoLevel: pgx.ReadCommitted,
 })
 if err != nil {
  return fmt.Errorf("db: failed to begin transaction: %w", err)
 }

 ctxWithTx := context.WithValue(ctx, txKey{}, tx)

 defer func() {
  if p := recover(); p != nil {
   _ = tx.Rollback(ctx)
   panic(p)
  }
 }()

 if err := fn(ctxWithTx); err != nil {
  if rbErr := tx.Rollback(ctx); rbErr != nil {
   return fmt.Errorf("db: rollback failed: %v (original error: %w)", rbErr, err)
  }
  return err
 }

 if err := tx.Commit(ctx); err != nil {
  return fmt.Errorf("db: failed to commit transaction: %w", err)
 }

 return nil
}

func ExecutorFromContext(ctx context.Context, pool *pgxpool.Pool) Executor {
 if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
  return tx
 }
 return pool
}
