package structs

import "github.com/figment-networks/graph-demo/manager/structs"

type GetBlockResp struct {
	Block structs.Block
	Txs   []structs.Transaction
}
