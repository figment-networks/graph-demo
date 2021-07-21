package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	wStructs "github.com/figment-networks/graph-demo/cosmos-worker/api/structs"
	"github.com/figment-networks/graph-demo/manager/structs"
)

type NetworkClient interface {
	GetByHeight(ctx context.Context, height uint64) (structs.BlockAndTx, error)
}

type RunnerClient interface {
	PopulateEvent(ctx context.Context, event uint64)
}

type Client struct {
	nc NetworkClient
}

func NewClient(nc NetworkClient) *Client {
	return &Client{nc: nc}
}

func (c *Client) GetByHeight(ctx context.Context, height uint64) (structs.BlockAndTx, error) {
	return c.nc.GetByHeight(ctx, height)
	// var bTx wStructs.BlockAndTx

	// resp, err := http.Get(fmt.Sprintf("%s/getBlock/%d", c.url, height))
	// if err != nil {
	// 	return structs.BlockAndTx{}, err
	// }
	// defer resp.Body.Close()

	// byteResp, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return structs.BlockAndTx{}, err
	// }

	// if err = json.Unmarshal(byteResp, &bTx); err != nil {
	// 	return structs.BlockAndTx{}, err
	// }

	// return structs.BlockAndTx{
	// 	Block:        bTx.Block,
	// 	Transactions: bTx.Txs,
	// }, nil

}

func (c *Client) PopulateEvent(ctx context.Context, event uint64) error {
	return c.PopulateEvent(ctx, event)
}
