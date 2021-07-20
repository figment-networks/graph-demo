package client

import (
	"context"
	"time"

	"github.com/figment-networks/graph-demo/manager/structs"

	// cryptoAmino "github.com/cosmos/amino-js/go/lib/tendermint/tendermint/crypto/encoding/amino"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"

	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"

	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	// "github.com/cosmos/cosmos-sdk/x/authz"

	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	// capabilityTypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	crisisTypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidenceTypes "github.com/cosmos/cosmos-sdk/x/evidence/types"

	// feegrantTypes "github.com/cosmos/cosmos-sdk/x/feegrant/types"

	// genutilTypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	// mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	// paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	// simulationTypes "github.com/cosmos/cosmos-sdk/x/simulation/types"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	// sTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/types/tx"
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

	log       *zap.Logger
	txDecoder types.TxDecoder

	// GRPC
	txServiceClient tx.ServiceClient
	tmServiceClient tmservice.ServiceClient
	rateLimiterGRPC *rate.Limiter

	// msgService
	interfaceRegistry codecTypes.InterfaceRegistry
	routes            map[string]MsgServiceHandler

	cfg *ClientConfig
}

type MsgServiceHandler = func(ctx types.Context, req types.Msg) (*types.Result, error)

// New returns a new client for a given endpoint
func New(logger *zap.Logger, cli *grpc.ClientConn, cfg *ClientConfig, chainID string) *Client {
	rateLimiterGRPC := rate.NewLimiter(rate.Limit(cfg.ReqPerSecond), cfg.ReqPerSecond)

	interfaceRegistry := codecTypes.NewInterfaceRegistry()
	authTypes.RegisterInterfaces(interfaceRegistry)
	bankTypes.RegisterInterfaces(interfaceRegistry)
	crisisTypes.RegisterInterfaces(interfaceRegistry)
	distributionTypes.RegisterInterfaces(interfaceRegistry)
	evidenceTypes.RegisterInterfaces(interfaceRegistry)
	govTypes.RegisterInterfaces(interfaceRegistry)
	slashingTypes.RegisterInterfaces(interfaceRegistry)
	stakingTypes.RegisterInterfaces(interfaceRegistry)
	upgradeTypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)

	codec := codec.NewProtoCodec(interfaceRegistry)
	txDecoder := defaultTxDecoder(codec)

	return &Client{
		chainID:         chainID,
		log:             logger,
		txDecoder:       txDecoder,
		tmServiceClient: tmservice.NewServiceClient(cli),
		txServiceClient: tx.NewServiceClient(cli),
		rateLimiterGRPC: rateLimiterGRPC,
		cfg:             cfg,
	}
}
