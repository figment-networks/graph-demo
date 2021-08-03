package structs

import (
	"errors"
	"math/big"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("record not found")
)

type BlockEvidence string

type QueriesResp map[string]map[uint64]BlockAndTx
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

	Header     BlockHeader     `json:"header"`
	Data       BlockData       `json:"data"`
	Evidence   []BlockEvidence `json:"evidence,omitempty"`
	LastCommit *Commit         `json:"last_commit,omitempty"`
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
	LastCommitHash string `json:"last_commit_hash,omitempty"`
	DataHash       string `json:"data_hash,omitempty"`
	// hashes from the app output from the prev block
	ValidatorsHash     string `json:"validators_hash,omitempty"`
	NextValidatorsHash string `json:"next_validators_hash,omitempty"`
	ConsensusHash      string `json:"consensus_hash,omitempty"`
	AppHash            string `json:"app_hash,omitempty"`
	LastResultsHash    string `json:"last_results_hash,omitempty"`
	// consensus info
	EvidenceHash    string `json:"evidence_hash,omitempty"`
	ProposerAddress string `json:"proposer_address,omitempty"`
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
	ValidatorAddress string    `json:"validator_address,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
	Signature        string    `json:"signature,omitempty"`
}

// Data contains the set of transactions included in the block
type BlockData struct {
	// Txs that will be applied by state @ block.Height+1.
	// NOTE: not all txs here are valid.  We're just agreeing on the order first.
	// This means that block.AppHash does not include these txs.
	Txs [][]byte `json:"txs,omitempty"`
}
type BlockID struct {
	Hash          string        `json:"hash,omitempty"`
	PartSetHeader PartSetHeader `json:"part_set_header"`
}

// PartsetHeader
type PartSetHeader struct {
	Total uint32 `json:"total,omitempty"`
	Hash  string `json:"hash,omitempty"`
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

	// ChainID - chain id of transacion
	ChainID string `json:"chain_id,omitempty"`
	// Height - height of the block of transaction
	Height uint64 `json:"height,omitempty"`
	// Hash of the transaction
	Hash string `json:"hash,omitempty"`
	// BlockHash - hash of the block of transaction
	BlockHash string `json:"block_hash,omitempty"`
	// Time - time of transaction
	Time time.Time `json:"time,omitempty"`

	// Namespace for the Code
	CodeSpace string `json:"code_space,omitempty"`

	Code uint64 `json:"code,omitempty"`
	// GasWanted
	GasWanted uint64 `json:"gas_wanted,omitempty"`
	// GasUsed
	GasUsed uint64 `json:"gas_used,omitempty"`

	Info string `json:"info,omitempty"`
	// Memo - the description attached to transactions
	Memo string `json:"memo,omitempty"`

	Result string `json:"result,omitempty"`

	Signatures []string `json:"signatures,omitempty"`

	AuthInfo *AuthInfo `json:"auth_info,omitempty"`

	ExtensionOptions []ExtensionOptions `json:"extension_options,omitempty"`

	Logs []Log `json:"logs,omitempty"`

	Messages []Message `json:"messages,omitempty"`

	NonCriticalExtensionOptions []NonCriticalExtensionOptions `json:"non_critical_extension_options,omitempty"`

	// RawLog - RawLog transaction's log bytes
	RawLog []byte `json:"raw_log,omitempty"`
	// TxRaw - Raw transaction bytes
	TxRaw TxRaw `json:"tx_raw,omitempty"`
}

type Log struct {
	MsgIndex uint64  `json:"msg_index,omitempty"`
	Log      string  `json:"log,omitempty"`
	Events   []Event `json:"events,omitempty"`
}

type Message struct {
	Message []byte `json:"message,omitempty"`
	Raw     Any    `json:"raw,omitempty"`
}

type Event struct {
	Type       string            `json:"type,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type TxRaw struct {
	TxRaw []byte `json:"message,omietmpty"`
	Raw   Any    `json:"raw,omitempty"`
}

type ExtensionOptions struct {
	ExtensionOption []byte `json:"extension_option,omitempty"`
	Raw             Any    `json:"raw,omitempty"`
}

type NonCriticalExtensionOptions struct {
	NonCriticalExtensionOption []byte `json:"non_critical_extension_option,omitempty"`
	Raw                        Any    `json:"raw,omitempty"`
}

type Any struct {
	TypeURL string `json:"type_url,omitempty"`
	Value   []byte `json:"value,omitempty"`
}

type AuthInfo struct {
	Fee         *Fee         `json:"fee,omitempty"`
	SignerInfos []SignerInfo `json:"signer_infos,omitempty"`
}

type SignerInfo struct {
	PublicKey *PublicKey `json:"public_key,omitempty"`
	ModeInfo  string     `json:"mode_info,omitempty"`
	Sequence  uint64
}

type PublicKey struct {
	Key string `json:"key,omitempty"`
	Raw Any    `json:"raw,omitempty"`
}

type Fee struct {
	Amount    *big.Int `json:"amount,omitempty"`
	Currency  string   `json:"currency,omitempty"`
	GasLimit  uint64   `json:"gas_limit,omitempty"`
	Sender    string   `json:"payer,omitempty"`
	Recipient string   `json:"grater,omitempty"`
}
