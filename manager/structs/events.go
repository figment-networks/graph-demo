package structs

const (
	EVENT_NEW_BLOCK       = "newBlock"
	EVENT_NEW_TRANSACTION = "newTransaction"
)

type EventNewBlock struct {
	Height uint64 `json:"height"`
}

type EventNewTransaction struct {
	Height uint64 `json:"height"`
}
