package client

import (
	"context"
	"time"

	"github.com/figment-networks/graph-demo/manager/structs"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

type GRPC interface {
	GetBlock(ctx context.Context, height uint64) (block structs.BlockAndTx, er error)
	GetLatest(ctx context.Context) (block structs.BlockAndTx, er error)
}
type ClientConfig struct {
	ReqPerSecond        int
	TimeoutBlockCall    time.Duration
	TimeoutSearchTxCall time.Duration
}

// Client
type Client struct {
	chainID string
	log     *zap.Logger

	// GRPC
	cfg             *ClientConfig
	txServiceClient tx.ServiceClient
	tmServiceClient tmservice.ServiceClient
	rateLimiterGRPC *rate.Limiter
}

type MsgServiceHandler = func(ctx types.Context, req types.Msg) (*types.Result, error)

// New returns a new client for a given endpoint
func New(logger *zap.Logger, cli *grpc.ClientConn, cfg *ClientConfig, chainID string) *Client {
	rateLimiterGRPC := rate.NewLimiter(rate.Limit(cfg.ReqPerSecond), cfg.ReqPerSecond)

	return &Client{
		chainID:         chainID,
		log:             logger,
		cfg:             cfg,
		txServiceClient: tx.NewServiceClient(cli),
		tmServiceClient: tmservice.NewServiceClient(cli),
		rateLimiterGRPC: rateLimiterGRPC,
	}
}
