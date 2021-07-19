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

type BlockAndTx struct {
	BlockID      BlockID       `json:"block_id,omitempty"`
	Block        Block         `json:"block"`
	Transactions []Transaction `json:"transactions"`
}

// Block contains the block details
type Block struct {
	// Hash of the Block
	Hash string `json:"hash,omitempty"`
	// Height of the Block
	Height uint64 `json:"height,omitempty"`
	// Time of the Block
	Time time.Time `json:"time,omitempty"`
	// ChainID
	ChainID string `json:"chain_id,omitempty"`
	// Number of transactions
	NumberOfTransactions uint64 `json:"tx_num,omitempty"`

	Header     BlockHeader       `json:"header"`
	Data       BlockData         `json:"data"`
	Evidence   BlockEvidenceList `json:"evidence"`
	LastCommit *Commit           `json:"last_commit,omitempty"`
}

type BlockHeader struct {
	// basic block info
	Version Consensus `json:"version"`
	ChainID string    `json:"chain_id,omitempty"`
	Height  int64     `json:"height,omitempty"`
	Time    time.Time `json:"time"`
	// prev block info
	LastBlockId BlockID `json:"last_block_id"`
	// hashes of block data
	LastCommitHash []byte `json:"last_commit_hash,omitempty"`
	DataHash       []byte `json:"data_hash,omitempty"`
	// hashes from the app output from the prev block
	ValidatorsHash     []byte `json:"validators_hash,omitempty"`
	NextValidatorsHash []byte `json:"next_validators_hash,omitempty"`
	ConsensusHash      []byte `json:"consensus_hash,omitempty"`
	AppHash            []byte `json:"app_hash,omitempty"`
	LastResultsHash    []byte `json:"last_results_hash,omitempty"`
	// consensus info
	EvidenceHash    []byte `json:"evidence_hash,omitempty"`
	ProposerAddress []byte `json:"proposer_address,omitempty"`
}

type BlockEvidenceList struct {
	Evidence []BlockEvidence `json:"evidence"`
}

type BlockEvidence struct {
	// Types that are valid to be assigned to Sum:
	//	*Evidence_DuplicateVoteEvidence
	//	*Evidence_LightClientAttackEvidence
	//Sum isEvidence_Sum `protobuf_oneof:"sum"`
}

// Commit contains the evidence that a block was committed by a set of validators.
type Commit struct {
	Height     int64       `json:"height,omitempty"`
	Round      int32       `json:"round,omitempty"`
	BlockID    BlockID     `json:"block_id"`
	Signatures []CommitSig `json:"signatures"`
}

// CommitSig is a part of the Vote included in a Commit.
type CommitSig struct {
	BlockIdFlag      int32     `json:"block_id_flag,omitempty"`
	ValidatorAddress []byte    `json:"validator_address,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
	Signature        []byte    `json:"signature,omitempty"`
}

// Data contains the set of transactions included in the block
type BlockData struct {
	// Txs that will be applied by state @ block.Height+1.
	// NOTE: not all txs here are valid.  We're just agreeing on the order first.
	// This means that block.AppHash does not include these txs.
	Txs [][]byte `json:"txs,omitempty"`
}
type BlockID struct {
	Hash          []byte        `json:"hash,omitempty"`
	PartSetHeader PartSetHeader `json:"part_set_header"`
}

// PartsetHeader
type PartSetHeader struct {
	Total uint32 `json:"total,omitempty"`
	Hash  []byte `json:"hash,omitempty"`
}

// Consensus captures the consensus rules for processing a block in the blockchain,
// including all blockchain data structures and the rules of the application's
// state transition machine.
type Consensus struct {
	Block uint64 `json:"block,omitempty"`
	App   uint64 `json:"app,omitempty"`
}

// Transaction contains the blockchain transaction details
type Transaction struct {
	// ID of transaction assigned on database write
	ID uuid.UUID `json:"id,omitempty"`
	// Created at
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// Updated at
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

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
