package store

import (
	"context"
	"errors"

	"github.com/figment-networks/graph-demo/manager/structs"
)

var (
	ErrDriverDoesNotExists    = errors.New("driver does not exist")
	ErrEmptyTransactionPassed = errors.New("empty transaction passed")
)

type Storager interface {
	Close() error

	StoreBlock(ctx context.Context, bl structs.Block) error
	StoreTransactions(ctx context.Context, txs []structs.Transaction) error
	GetBlockByHeight(ctx context.Context, height uint64, chainID string) (structs.Block, error)
	GetTransactionsByParam(ctx context.Context, chainID string, param string, value interface{}) ([]structs.Transaction, error)

	SetLatestHeight(ctx context.Context, chainID string, height uint64) (err error)
	GetLatestHeight(ctx context.Context, chainID string) (height uint64, err error)
}

type Store struct {
	driver Storager
}

func NewStore(d Storager) *Store {
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
	return s.driver.GetBlockByHeight(ctx, height, chainID)
}

func (s *Store) GetTransactionsByParam(ctx context.Context, chainID string, param string, value interface{}) ([]structs.Transaction, error) {
	return s.driver.GetTransactionsByParam(ctx, chainID, param, value)
}

func (s *Store) GetLatestHeight(ctx context.Context, chainID string) (height uint64, err error) {
	return s.driver.GetLatestHeight(ctx, chainID)
}

func (s *Store) SetLatestHeight(ctx context.Context, chainID string, height uint64) (err error) {
	return s.driver.SetLatestHeight(ctx, chainID, height)
}
