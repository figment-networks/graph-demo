package store

import (
	"context"
	"errors"
	"time"

	"github.com/figment-networks/graph-demo/manager/store/loader"
	"github.com/figment-networks/graph-demo/manager/store/params"

	"github.com/figment-networks/indexing-engine/structs"
)

var (
	ErrDriverDoesNotExists    = errors.New("driver does not exist")
	ErrEmptyTransactionPassed = errors.New("empty transaction passed")
)

type DBDriver interface {
	TransactionStore
	BlockStore

	Close() error
}

type DataStore interface {
	TransactionStore
	BlockStore
}

type TransactionStore interface {
	StoreTransactions(ctx context.Context, txs []structs.Transaction) error

	GetTransactions(ctx context.Context, tsearch params.TransactionSearch) ([]structs.Transaction, error)
}

type BlockStore interface {
	StoreBlock(ctx context.Context, bwm structs.Block) error

	GetBlockForTime(ctx context.Context, blx structs.Block, time time.Time) (structs.Block, bool, error)
}

type Store struct {
	l *loader.Loader
}

func New(l *loader.Loader) *Store {
	return &Store{l: l}
}

func (s *Store) StoreTransactions(ctx context.Context, txs []structs.Transaction) error {
	if len(txs) == 0 {
		return ErrEmptyTransactionPassed
	}

	d, ok := s.l.Get(loader.NC{Network: txs[0].Network, ChainID: txs[0].ChainID})
	if !ok {
		return ErrDriverDoesNotExists
	}

	return d.StoreTransactions(ctx, txs)
}

func (s *Store) StoreBlock(ctx context.Context, bl structs.Block) error {
	d, ok := s.l.Get(loader.NC{Network: bl.Network, ChainID: bl.ChainID})
	if !ok {
		return ErrDriverDoesNotExists
	}
	return d.StoreBlock(ctx, bl)
}

func (s *Store) GetTransactions(ctx context.Context, tsearch params.TransactionSearch) ([]structs.Transaction, error) {

	chain := ""
	if len(tsearch.ChainIDs) == 1 {
		chain = tsearch.ChainIDs[0]
	} else {
		chain = "" // in case of bigger numbers of chains we assume the same db
	}

	d, ok := s.l.Get(loader.NC{Network: tsearch.Network, ChainID: chain})
	if !ok {
		return nil, ErrDriverDoesNotExists
	}
	return d.GetTransactions(ctx, tsearch)
}

func (s *Store) GetBlockForTime(ctx context.Context, blx structs.Block, time time.Time) (b structs.Block, ok bool, err error) {
	d, ok := s.l.Get(loader.NC{Network: blx.Network, ChainID: blx.ChainID})
	if !ok {
		return b, false, ErrDriverDoesNotExists
	}

	return d.GetBlockForTime(ctx, blx, time)
}
