package http

import "github.com/figment-networks/graph-demo/manager/structs"

type All struct {
	Block structs.Block
	Txs   []structs.Transaction
}

type Latest struct {
	Height uint64
}
