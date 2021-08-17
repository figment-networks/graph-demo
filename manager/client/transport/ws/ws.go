package ws

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/figment-networks/graph-demo/connectivity"
	wsapi "github.com/figment-networks/graph-demo/connectivity/ws"
)

type CosmosWSTransport struct {
	sess wsapi.SyncSender
}

func NewCosmosWSTransport(sess wsapi.SyncSender) *CosmosWSTransport {
	return &CosmosWSTransport{sess: sess}
}

func (ng *CosmosWSTransport) GetAll(ctx context.Context, height uint64) (string, []string, error) {
	resp, err := ng.sess.SendSync("get_all", []json.RawMessage{[]byte(strconv.FormatUint(height, 10))})
	if err != nil {
		return "", nil, err
	}
	if resp.Error != nil && resp.Error.Message != "" {
		return "", nil, errors.New(resp.Error.Message)
	}

	var response connectivity.BlockAndTransactionIDs
	if err = json.Unmarshal(resp.Result, &response); err != nil {
		return "", nil, err
	}

	return response.BlockID, response.TxsIDs, err
}

func (ng *CosmosWSTransport) GetLatest(ctx context.Context) (h uint64, err error) {
	resp, err := ng.sess.SendSync("get_latest", nil)
	if err != nil {
		return h, err
	}

	if err = json.Unmarshal(resp.Result, &h); err != nil {
		return h, err
	}

	return h, nil

}
