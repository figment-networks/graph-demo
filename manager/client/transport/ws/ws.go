package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	wsapi "github.com/figment-networks/graph-demo/connectivity/ws"
)

type CosmosWSTransport struct {
	sess wsapi.SyncSender
}

func NewCosmosWSTransport(sess wsapi.SyncSender) *CosmosWSTransport {
	return &CosmosWSTransport{sess: sess}
}

func (ng *CosmosWSTransport) GetAll(ctx context.Context, height uint64) error {
	resp, err := ng.sess.SendSync("get_all", []json.RawMessage{[]byte(strconv.FormatUint(height, 10))})
	if err != nil {
		return err
	}
	if resp.Error != nil && resp.Error.Message != "" {
		return errors.New(resp.Error.Message)
	}

	return err
}

func (ng *CosmosWSTransport) GetLatest(ctx context.Context) (h uint64, err error) {
	resp, err := ng.sess.SendSync("get_latest", nil)
	if err != nil {
		return h, err
	}

	if err = json.Unmarshal(resp.Result, &h); err != nil {
		var response errResponse
		if err2 := json.Unmarshal(resp.Result, &response); err2 == nil {
			err = errors.New("unexpected response")
			for _, errStr := range response.Errors {
				err = fmt.Errorf("%s: %s", err.Error(), errStr)
			}
		}
		return h, err
	}

	return h, nil

}

type errResponse struct {
	Errors []errMsg `json:"errors"`
}

type errMsg struct {
	Message string `json:"message"`
}
