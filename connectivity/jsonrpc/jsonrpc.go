package jsonrpc

import "encoding/json"

type JsonRPCRequest struct {
	ID      uint64          `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type JsonRPCError struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	Data    []interface{} `json:"data"`
}

type JsonRPCResponse struct {
	ID      uint64          `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Error   *JsonRPCError   `json:"error,omitempty"`
	Result  json.RawMessage `json:"result"`
}
