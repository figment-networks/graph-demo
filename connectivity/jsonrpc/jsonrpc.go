package jsonrpc

import "encoding/json"

type Request struct {
	ID      uint64            `json:"id"`
	JSONRPC string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}

type Error struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	Data    []interface{} `json:"data"`
}

type Response struct {
	ID      uint64          `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Error   *Error          `json:"error,omitempty"`
	Result  json.RawMessage `json:"result"`
}

type Hybrid struct {
	ID      uint64            `json:"id"`
	JSONRPC string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
	Error   *Error            `json:"error,omitempty"`
	Result  json.RawMessage   `json:"result"`
}
