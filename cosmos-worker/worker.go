package cosmosworker

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/figment-networks/graph-demo/manager/structs"
	cStructs "github.com/figment-networks/graph-demo/manager/worker/connectivity/structs"

	"go.uber.org/zap"
)

const page = 100
const blockchainEndpointLimit = 20

var (
	// ErrBadRequest is returned when cannot unmarshal message
	ErrBadRequest = errors.New("bad request")
)

type GRPC interface {
	GetBlock(ctx context.Context, height uint64) (block structs.Block, er error)
	SearchTx(ctx context.Context, block structs.Block, height, perPage uint64) (txs []structs.Transaction, err error)
}

type OutputSender interface {
	Send(cStructs.TaskResponse) error
}

// IndexerClient is implementation of a client (main worker code)

type Worker struct {
	grpc  GRPC
	log   *zap.Logger
	sLock sync.Mutex
}

// NewWorker is Worker constructor
func NewWorker(ctx context.Context, logger *zap.Logger, grpc GRPC, maximumHeightsToGet uint64) *Worker {
	ic := &Worker{
		grpc: grpc,
		log:  logger,
	}

	return ic
}

// GetTransactions gets new transactions and blocks from cosmos for given range
func (w *Worker) GetTransactions(ctx context.Context, height uint64) (block structs.Block, txs []structs.Transaction, err error) {
	w.log.Debug("[COSMOS-WORKER] Getting block", zap.Uint64("height", height))

	block, err = w.grpc.GetBlock(ctx, height)
	if err != nil {
		w.log.Debug("[COSMOS-CLIENT] Err Getting block", zap.Uint64("block", height), zap.Error(err), zap.Uint64("txs", block.NumberOfTransactions))
		return structs.Block{}, nil, fmt.Errorf("error fetching block: %d %w ", uint64(height), err)
	}

	if block.NumberOfTransactions > 0 {
		w.log.Debug("[COSMOS-CLIENT] Getting txs", zap.Uint64("block", height), zap.Uint64("txs", block.NumberOfTransactions))
		txs, err = w.grpc.SearchTx(ctx, block, height, page)
		if err != nil {
			return structs.Block{}, nil, err
		}

		w.log.Debug("[COSMOS-CLIENT] txErr Getting txs", zap.Uint64("block", height), zap.Error(err), zap.Uint64("txs", block.NumberOfTransactions))
	}

	w.log.Debug("[COSMOS-CLIENT] Got block", zap.Uint64("block", height), zap.Uint64("txs", block.NumberOfTransactions))
	return block, txs, err
}
