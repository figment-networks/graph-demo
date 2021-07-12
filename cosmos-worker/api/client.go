package api

import (
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types/tx"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

type ClientConfig struct {
	ReqPerSecond        int
	TimeoutBlockCall    time.Duration
	TimeoutSearchTxCall time.Duration
}

// Client
type Client struct {
	logger *zap.Logger
	cli    *grpc.ClientConn
	Sbc    *SimpleBlockCache

	// GRPC
	txServiceClient    tx.ServiceClient
	tmServiceClient    tmservice.ServiceClient
	rateLimiterGRPC    *rate.Limiter
	bankClient         bankTypes.QueryClient
	distributionClient distributionTypes.QueryClient
	stakingClient      stakingTypes.QueryClient

	cfg *ClientConfig
}

// NewClient returns a new client for a given endpoint
func NewClient(logger *zap.Logger, cli *grpc.ClientConn, cfg *ClientConfig) *Client {
	rateLimiterGRPC := rate.NewLimiter(rate.Limit(cfg.ReqPerSecond), cfg.ReqPerSecond)

	return &Client{
		logger:             logger,
		Sbc:                NewSimpleBlockCache(400),
		tmServiceClient:    tmservice.NewServiceClient(cli),
		txServiceClient:    tx.NewServiceClient(cli),
		bankClient:         bankTypes.NewQueryClient(cli),
		distributionClient: distributionTypes.NewQueryClient(cli),
		stakingClient:      stakingTypes.NewQueryClient(cli),
		rateLimiterGRPC:    rateLimiterGRPC,
		cfg:                cfg,
	}
}
