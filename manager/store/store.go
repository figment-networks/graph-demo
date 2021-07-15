package store

import (
	"context"
	"errors"

	"github.com/figment-networks/graph-demo/manager/store/params"
	"github.com/figment-networks/graph-demo/manager/store/postgres"
	"github.com/figment-networks/graph-demo/manager/structs"
)

var (
	ErrDriverDoesNotExists    = errors.New("driver does not exist")
	ErrEmptyTransactionPassed = errors.New("empty transaction passed")
)

type StoreIface interface {
	Close() error

	StoreTransactions(ctx context.Context, txs []structs.Transaction) error
	StoreBlock(ctx context.Context, bl structs.Block) error
	GetTransactions(ctx context.Context, tsearch params.TransactionSearch) ([]structs.Transaction, error)
	GetBlockByHeight(ctx context.Context, height uint64, chainID, network string) (b structs.Block, err error)
}

type Store struct {
	driver postgres.Driver
}

func New(d postgres.Driver) *Store {
	return &Store{
		driver: d,
	}
}

func (s *Store) Close() error {
	return s.driver.Close()
}

func (s *Store) StoreTransactions(ctx context.Context, txs []structs.Transaction) error {
	if len(txs) == 0 {
		return ErrEmptyTransactionPassed
	}

	return s.driver.StoreTransactions(ctx, txs)
}

func (s *Store) StoreBlock(ctx context.Context, bl structs.Block) error {
	return s.driver.StoreBlock(ctx, bl)
}

func (s *Store) GetTransactions(ctx context.Context, tsearch params.TransactionSearch) ([]structs.Transaction, error) {
	return s.driver.GetTransactions(ctx, tsearch)
}

func (s *Store) GetBlockByHeight(ctx context.Context, height uint64, chainID, network string) (b structs.Block, err error) {
	return s.driver.GetBlockBytHeight(ctx, height, chainID, network)
}
