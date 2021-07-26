package ws

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	wsapi "github.com/figment-networks/graph-demo/connectivity/ws"
	"github.com/figment-networks/graph-demo/manager/structs"
)

type CosmosWSTransport struct {
	sess wsapi.SyncSender
}

func NewCosmosWSTransport(sess wsapi.SyncSender) *CosmosWSTransport {
	return &CosmosWSTransport{sess: sess}
}

func (ng *CosmosWSTransport) GetAll(ctx context.Context, height uint64) (err error) {
	resp, err := ng.sess.SendSync("get_all", []json.RawMessage{[]byte(strconv.FormatUint(height, 10))})
	if err != nil {
		return err
	}
	if resp.Error != nil && resp.Error.Message != "" {
		return errors.New(resp.Error.Message)
	}

	return nil

}

func (ng *CosmosWSTransport) GetLatest(ctx context.Context) (b structs.Block, err error) {
	resp, err := ng.sess.SendSync("get_latest", nil)
	if err != nil {
		return b, err
	}

	if err = json.Unmarshal(resp.Result, &b); err != nil {
		return b, err
	}

	return b, nil

}
