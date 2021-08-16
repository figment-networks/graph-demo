package connectivity

import (
	"context"
	"encoding/json"
)

type BlockID struct {
	ID string `json:"id"`
}

type TransactionIDs struct {
	IDs []string `json:"ids"`
}

type BlockAndTransactionIDs struct {
	BlockID string   `json:"block_id"`
	TxsIDs  []string `json:"txs_ids"`
}
type Response interface {
	Send(result json.RawMessage, er error) error
}

type Request interface {
	ConnID() string
	Arguments() []json.RawMessage
}

type FunctionCallHandler interface {
	Get(name string) (h Handler, ok bool)
	Add(name string, h Handler)
}

type Handler func(ctx context.Context, req Request, resp Response)

type WsConn interface {
	Attach(string, Handler)
}
