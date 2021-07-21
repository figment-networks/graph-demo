package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/figment-networks/graph-demo/manager/structs"
)

type CosmosHTTPTransport struct {
	c       *http.Client
	address string
}

func NewCosmosHTTPTransport(address string, c *http.Client) *CosmosHTTPTransport {
	return &CosmosHTTPTransport{
		address: address,
		c:       c,
	}
}

func (ng *CosmosHTTPTransport) GetByHeight(ctx context.Context, height uint64) (bTx structs.BlockAndTx, er error) {
	resp, err := http.Get(fmt.Sprintf("%s/getBlock/%d", ng.address, height))
	if err != nil {
		return structs.BlockAndTx{}, err
	}
	defer resp.Body.Close()

	byteResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return structs.BlockAndTx{}, err
	}

	if err = json.Unmarshal(byteResp, &bTx); err != nil {
		return structs.BlockAndTx{}, err
	}

	return structs.BlockAndTx{
		Block:        bTx.Block,
		Transactions: bTx.Transactions,
	}, nil
}
