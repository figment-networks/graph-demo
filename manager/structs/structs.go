package structs

import (
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("record not found")
)

// Block contains the block details
type Block struct {
	// ID
	ID uuid.UUID `json:"id,omitempty"`
	// CreatedAt of block creation time in database
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// UpdatedAt of block update time in database
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	// Hash of the Block
	Hash string `json:"hash,omitempty"`
	// Height of the Block
	Height uint64 `json:"height,omitempty"`
	// Time of the Block
	Time time.Time `json:"time,omitempty"`
	// Epoch number
	Epoch string `json:"epoch,omitempty"`
	// Network name
	Network string `json:"network,omitempty"`
	// ChainID
	ChainID string `json:"chain_id,omitempty"`

	NumberOfTransactions uint64 `json:"num_txs,omitempty"`
}

// Transaction contains the blockchain transaction details
type Transaction struct {
	// ID of transaction assigned on database write
	ID uuid.UUID `json:"id,omitempty"`
	// Created at
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// Updated at
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	// Network name
	Network string `json:"network,omitempty"`

	// Hash of the transaction
	Hash string `json:"hash,omitempty"`
	// BlockHash - hash of the block of transaction
	BlockHash string `json:"block_hash,omitempty"`
	// Height - height of the block of transaction
	Height uint64 `json:"height,omitempty"`

	Epoch string `json:"epoch,omitempty"`
	// ChainID - chain id of transacion
	ChainID string `json:"chain_id,omitempty"`
	// Time - time of transaction
	Time time.Time `json:"time,omitempty"`

	// Fee - Fees for transaction (if applies)
	Fee []TransactionAmount `json:"transaction_fee,omitempty"`
	// GasWanted
	GasWanted uint64 `json:"gas_wanted,omitempty"`
	// GasUsed
	GasUsed uint64 `json:"gas_used,omitempty"`
	// Memo - the description attached to transactions
	Memo string `json:"memo,omitempty"`

	// Events - Transaction contents
	Events TransactionEvents `json:"events,omitempty"`

	// Raw - Raw transaction bytes
	Raw []byte `json:"raw,omitempty"`

	// RawLog - RawLog transaction's log bytes
	RawLog []byte `json:"raw_log,omitempty"`

	// HasErrors - indicates if Transaction has any errors inside
	HasErrors bool `json:"has_errors"`
}

// TransactionEvents - a set of TransactionEvent
type TransactionEvents []TransactionEvent

func (te *TransactionEvents) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &te)
}

// TransactionEvent part of transaction contents
type TransactionEvent struct {
	// ID UniqueID of event
	ID string `json:"id,omitempty"`
	// The Kind of event
	Kind string `json:"kind,omitempty"`
	// Type of transaction
	Type []string `json:"type,omitempty"`
	// Collection from where transaction came from
	Module string `json:"module,omitempty"`
	// List of sender accounts with optional amounts
	// Subcontents of event
	Sub []SubsetEvent `json:"sub,omitempty"`
}

// TransactionAmount structure holding amount information with decimal implementation (numeric * 10 ^ exp)
type TransactionAmount struct {
	// Textual representation of Amount
	Text string `json:"text,omitempty"`
	// The currency in what amount is returned (if applies)
	Currency string `json:"currency,omitempty"`

	// Numeric part of the amount
	Numeric *big.Int `json:"numeric,omitempty"`
	// Exponential part of amount obviously 0 by default
	Exp int32 `json:"exp,omitempty"`
}

// SubsetEvent - structure storing main contents of transacion
type SubsetEvent struct {
	// ID UniqueID of subsetevent
	ID string `json:"id,omitempty"`
	// Type of transaction
	Type   []string `json:"type,omitempty"`
	Action string   `json:"action,omitempty"`
	// Collection from where transaction came from
	Module string `json:"module,omitempty"`
	// List of sender accounts with optional amounts
	Sender []EventTransfer `json:"sender,omitempty"`
	// List of recipient accounts with optional amounts
	Recipient []EventTransfer `json:"recipient,omitempty"`
	// The list of all accounts that took part in the subsetevent
	Node map[string][]Account `json:"node,omitempty"`
	// Transaction nonce
	Nonce string `json:"nonce,omitempty"`
	// Completion time
	Completion *time.Time `json:"completion,omitempty"`
	// List of Amounts
	Amount map[string]TransactionAmount `json:"amount,omitempty"`
	// List of Transfers with amounts and optional recipients
	Transfers map[string][]EventTransfer `json:"transfers,omitempty"`
	// Optional error if occurred
	Error *SubsetEventError `json:"error,omitempty"`
	// Set of additional parameters attached to transaction (used as last resort)
	Additional map[string][]string `json:"additional,omitempty"`
	// SubEvents because some messages are in fact carying another messages inside
	Sub []SubsetEvent `json:"sub,omitempty"`
}

// EventTransfer - Account and Amounts pair
type EventTransfer struct {
	// Account recipient
	Account Account `json:"account,omitempty"`
	// Amounts from Transfer
	Amounts []TransactionAmount `json:"amounts,omitempty"`
}

// Account - Extended Account information
type Account struct {
	// Unique account identifier
	ID string `json:"id"`
	// External optional account details (if applies)
	Details *AccountDetails `json:"detail,omitempty"`
}

// AccountDetails External optional account details (if applies)
type AccountDetails struct {
	// Description of account
	Description string `json:"description,omitempty"`
	// Contact information
	Contact string `json:"contact,omitempty"`
	// Name of account
	Name string `json:"name,omitempty"`
	// Website address
	Website string `json:"website,omitempty"`
}

// SubsetEventError error structure for event
type SubsetEventError struct {
	// Message from error event
	Message string `json:"message,omitempty"`
}

type TransactionSearch struct {
	Height    uint64    `json:"height"`
	Type      SearchArr `json:"type"`
	BlockHash string    `json:"block_hash"`
	Hash      string    `json:"hash"`
	Account   []string  `json:"account"`
	Sender    []string  `json:"sender"`
	Receiver  []string  `json:"receiver"`
	Memo      string    `json:"memo"`

	AfterTime  time.Time `json:"before_time"`
	BeforeTime time.Time `json:"after_time"`

	AfterHeight  uint64 `json:"after_height"`
	BeforeHeight uint64 `json:"before_height"`
	Limit        uint64 `json:"limit"`
	Offset       uint64 `json:"offset"`

	Network  string   `json:"network"`
	ChainIDs []string `json:"chain_ids"`
	Epoch    string   `json:"epoch"`

	WithRaw    bool `json:"with_raw"`
	WithRawLog bool `json:"with_raw_log"`

	HasErrors bool `json:"has_errors"`
}

type SearchArr struct {
	Value []string `json:"value"`
	Any   bool     `json:"any"`
}
