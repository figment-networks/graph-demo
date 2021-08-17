package client

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/figment-networks/graph-demo/manager/structs"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Presistor interface {
	StoreTransactions(ctx context.Context, txs []structs.Transaction) ([]string, error)
	StoreBlock(ctx context.Context, block structs.Block) (string, error)
}
type ClientConfig struct {
	TimeoutBlockCall    time.Duration
	TimeoutSearchTxCall time.Duration
}

// Client
type Client struct {
	log             *zap.Logger
	cfg             *ClientConfig
	txServiceClient tx.ServiceClient
	tmServiceClient tmservice.ServiceClient
	persistor       Presistor
}

type MsgServiceHandler = func(ctx types.Context, req types.Msg) (*types.Result, error)

// New returns a new client for a given endpoint
func NewClient(logger *zap.Logger, cli *grpc.ClientConn, cfg *ClientConfig) *Client {
	return &Client{
		log:             logger,
		cfg:             cfg,
		txServiceClient: tx.NewServiceClient(cli),
		tmServiceClient: tmservice.NewServiceClient(cli),
	}
}

func (c *Client) LinkPersistor(p Presistor) {
	c.persistor = p
}
