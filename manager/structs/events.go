package structs

const (
	EVENT_NEW_BLOCK       = "newBlock"
	EVENT_NEW_TRANSACTION = "newTransaction"
)

type EventNewBlock struct {
	ID     string `json:"id"`
	Height uint64 `json:"height"`
}

type EventNewTransaction struct {
	BlockID string   `json:"block_id"`
	TxIDs   []string `json:"tx_ids"`
	Height  uint64   `json:"height"`
}

type Subs struct {
	Name           string
	StartingHeight uint64
}

type Register struct {
	Name    string `json:"name"`
	ChainID string `json:"chainID"`
}
