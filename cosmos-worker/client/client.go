package client

import (
	"context"
	"time"

	"github.com/figment-networks/graph-demo/manager/structs"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
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
	network string

	log *zap.Logger

	// GRPC
	txServiceClient tx.ServiceClient
	tmServiceClient tmservice.ServiceClient
	rateLimiterGRPC *rate.Limiter

	cfg *ClientConfig
}

// New returns a new client for a given endpoint
func New(logger *zap.Logger, cli *grpc.ClientConn, cfg *ClientConfig) *Client {
	rateLimiterGRPC := rate.NewLimiter(rate.Limit(cfg.ReqPerSecond), cfg.ReqPerSecond)

	return &Client{
		log:             logger,
		tmServiceClient: tmservice.NewServiceClient(cli),
		txServiceClient: tx.NewServiceClient(cli),
		rateLimiterGRPC: rateLimiterGRPC,
		cfg:             cfg,
	}
}
