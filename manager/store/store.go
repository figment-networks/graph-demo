package store

import (
	"context"
	"errors"

	"github.com/figment-networks/graph-demo/manager/store/postgres"
	"github.com/figment-networks/graph-demo/manager/structs"
)

var (
	ErrDriverDoesNotExists    = errors.New("driver does not exist")
	ErrEmptyTransactionPassed = errors.New("empty transaction passed")
)

type StoreIface interface {
	Close() error

	StoreBlock(ctx context.Context, bl structs.Block) error
	StoreTransactions(ctx context.Context, txs []structs.Transaction) error
	GetBlockByHeight(ctx context.Context, height uint64, chainID string) (structs.Block, error)
	GetTransactionsByHeight(ctx context.Context, height uint64, chainID string) ([]structs.Transaction, error)
}

type Store struct {
	driver *postgres.Driver
}

func New(d *postgres.Driver) *Store {
	return &Store{
		driver: d,
	}
}

func (s *Store) Close() error {
	return s.driver.Close()
}

func (s *Store) StoreBlock(ctx context.Context, b structs.Block) error {
	return s.driver.StoreBlock(ctx, b)
}

func (s *Store) StoreTransactions(ctx context.Context, txs []structs.Transaction) error {
	return s.driver.StoreTransactions(ctx, txs)
}

func (s *Store) GetBlockByHeight(ctx context.Context, height uint64, chainID string) (structs.Block, error) {
	return s.driver.GetBlockBytHeight(ctx, height, chainID)
}

func (s *Store) GetTransactionsByHeight(ctx context.Context, height uint64, chainID string) ([]structs.Transaction, error) {
	return s.driver.GetTransactions(ctx, height, chainID)
}
