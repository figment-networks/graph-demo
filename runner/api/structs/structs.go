package structs

import "github.com/figment-networks/graph-demo/manager/structs"

type QueriesResp map[string]map[uint64]BlockAndTx

type BlockAndTx struct {
	Block structs.Block
	Txs   []structs.Transaction
}
