package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	wStructs "github.com/figment-networks/graph-demo/cosmos-worker/structs"
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

func (c *Client) GetBlockByHeight(ctx context.Context, height uint64) (structs.Block, []structs.Transaction, error) {
	var getBlockResp wStructs.GetBlockResp

	resp, err := http.Get(fmt.Sprintf("%s/getBlock/%d", c.url, height))
	if err != nil {
		return structs.Block{}, nil, err
	}
	defer resp.Body.Close()

	byteResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return structs.Block{}, nil, err
	}

	if err = json.Unmarshal(byteResp, &getBlockResp); err != nil {
		return structs.Block{}, nil, err
	}

	return getBlockResp.Block, getBlockResp.Txs, nil

}
