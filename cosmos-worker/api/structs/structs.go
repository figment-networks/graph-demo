package structs

import "github.com/figment-networks/graph-demo/manager/structs"

type BlockAndTx struct {
	Block structs.Block
	Txs   []structs.Transaction
}

type Latest struct {
	Height uint64
}
