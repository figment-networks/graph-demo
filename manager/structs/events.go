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
	Height uint64 `json:"height"`
	Hash   string `json:"hash"`
}

type Subs struct {
	Name           string
	StartingHeight uint64
}

type Register struct {
	Name    string `json:"name"`
	ChainID string `json:"chainID"`
}
