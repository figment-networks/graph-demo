package client

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/figment-networks/graph-demo/manager/structs"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

type GRPC interface {
	GetBlock(ctx context.Context, height uint64) (block structs.Block, er error)
	SearchTx(ctx context.Context, block structs.Block, height, perPage uint64) (txs []structs.Transaction, err error)
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

func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	bank.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	distr.RegisterCodec(cdc)
	slashing.RegisterCodec(cdc)
	gov.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	crisis.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}
