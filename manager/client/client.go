package client

import (
	"context"
	"errors"
	"sort"

	rStructs "github.com/figment-networks/graph-demo/manager/structs"

	cStructs "github.com/figment-networks/graph-demo/manager/connectivity/structs"

	"go.uber.org/zap"
)

//go:generate mockgen -destination=./mocks/mock_client.go  -package=mocks github.com/figment-networks/graph-demo/manager/client TaskSender

// SelfCheck Flag describing should manager check anyway the latest version for network it has
var SelfCheck = true

var ErrIntegrityCheckFailed = errors.New("integrity check failed")
var ErrNotAvailable = errors.New("call unavailable")

var rewardTxTypesMap = map[string][]string{
	"cosmos": {"begin_redelegate", "begin_unbonding", "delegate", "undelegate", "withdraw_delegator_reward"},
	"kava":   {"begin_redelegate", "begin_unbonding", "claim_hard_reward", "claim_usdx_minting_reward", "delegate", "withdraw_delegator_reward"},
	"terra":  {"begin_redelegate", "begin_unbonding", "delegate", "undelegate", "withdraw_delegator_reward"},
}

type NetworkVersion struct {
	Network string
	ChainID string
	Version string
}

// ClientContractor a format agnostic
type ControllContractor interface {
	LatestData(ctx context.Context, ldr rStructs.LatestDataRequest) (ldResp rStructs.LatestDataResponse, er error)
	SyncData(ctx context.Context, ldr rStructs.SyncDataRequest) (ldResp rStructs.SyncDataResponse, er error)

	/*
		CheckMissingTransactions(ctx context.Context, nv NetworkVersion, heightRange structs.HeightRange, mode MissingDiffType, window uint64) (missingBlocks, missingTransactions [][2]uint64, err error)
		GetMissingTransactions(ctx context.Context, nv NetworkVersion, heightRange structs.HeightRange, params GetMissingTxParams) (run *Run, err error)
		GetRunningTransactions(ctx context.Context) (run []Run, err error)
		StopRunningTransactions(ctx context.Context, nv NetworkVersion, clean bool) (err error)
	*/
}

type TaskSender interface {
	Send([]cStructs.TaskRequest) (*cStructs.Await, error)
}

type Client struct {
	sender TaskSender
	logger *zap.Logger

	r *Running
}

func NewClient(logger *zap.Logger) *Client {
	c := &Client{
		logger: logger,
		r:      NewRunning(),
	}
	return c
}

func (hc *Client) LinkSender(sender TaskSender) {
	hc.sender = sender
}

func getRanges(in []uint64) (ranges [][2]uint64) {

	sort.SliceStable(in, func(i, j int) bool { return in[i] < in[j] })

	ranges = [][2]uint64{}
	var temp = [2]uint64{}
	for i, height := range in {

		if i == 0 {
			temp[0] = height
			temp[1] = height
			continue
		}

		if temp[1]+1 == height {
			temp[1] = height
			continue
		}

		ranges = append(ranges, temp)
		temp[0] = height
		temp[1] = height

	}
	if temp[1] != 0 {
		ranges = append(ranges, temp)
	}

	return ranges
}

// groupRanges Groups ranges to fit the window of X records
func groupRanges(ranges [][2]uint64, window uint64) (out [][2]uint64) {
	pregroup := [][2]uint64{}

	// (lukanus): first slice all the bigger ranges to get max(window)
	for _, r := range ranges {
		diff := r[1] - r[0]
		if diff > window {
			current := r[0]
		SliceDiff:
			for {
				next := current + window
				if next <= r[1] {
					pregroup = append(pregroup, [2]uint64{current, next})
					current = next + 1
					continue
				}

				pregroup = append(pregroup, [2]uint64{current, r[1]})
				break SliceDiff
			}
		} else {
			pregroup = append(pregroup, r)
		}
	}

	out = [][2]uint64{}

	var temp = [2]uint64{}
	for i, n := range pregroup {
		if i == 0 {
			temp[0] = n[0]
			temp[1] = n[1]
			continue
		}

		diff := n[1] - temp[0]
		if diff <= window {
			temp[1] = n[1]
			continue
		}
		out = append(out, temp)
		temp = [2]uint64{n[0], n[1]}
	}

	if temp[1] != 0 {
		out = append(out, temp)
	}
	return out
}

func (hc *Client) recoverPanic() {
	if p := recover(); p != nil {
		hc.logger.Error("[Client] Panic ", zap.Any("contents", p))
		hc.logger.Sync()
	}
}

func contains(slc []string, want string) bool {
	for _, s := range slc {
		if want == s {
			return true
		}
	}
	return false
}
