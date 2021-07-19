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

type Client struct {
	url        string
	httpClient *http.Client
}

func New(hc *http.Client, url string) *Client {
	return &Client{
		url:        url,
		httpClient: hc,
	}
}

func (c *Client) GetByHeight(ctx context.Context, height uint64) (structs.BlockAndTx, error) {
	var bTx wStructs.BlockAndTx

	resp, err := http.Get(fmt.Sprintf("%s/getBlock/%d", c.url, height))
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
		Transactions: bTx.Txs,
	}, nil

}
