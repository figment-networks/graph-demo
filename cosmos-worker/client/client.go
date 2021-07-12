package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	mStructs "github.com/figment-networks/indexer-manager/structs"
	cStructs "github.com/figment-networks/indexer-manager/worker/connectivity/structs"
	"github.com/figment-networks/indexing-engine/metrics"
	"github.com/figment-networks/indexing-engine/structs"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const page = 100
const blockchainEndpointLimit = 20

var (
	getTransactionDuration        *metrics.GroupObserver
	getLatestDuration             *metrics.GroupObserver
	getBlockDuration              *metrics.GroupObserver
	getRewardDuration             *metrics.GroupObserver
	getAccountBalanceDuration     *metrics.GroupObserver
	getAccountDelegationsDuration *metrics.GroupObserver
)

type GRPC interface {
	GetBlock(ctx context.Context, params structs.HeightHash) (block structs.Block, er error)
	SearchTx(ctx context.Context, r structs.HeightHash, block structs.Block, perPage uint64) (txs []structs.Transaction, err error)
	GetReward(ctx context.Context, params structs.HeightAccount) (resp structs.GetRewardResponse, err error)
	GetAccountBalance(ctx context.Context, params structs.HeightAccount) (resp structs.GetAccountBalanceResponse, err error)
	GetAccountDelegations(ctx context.Context, params structs.HeightAccount) (resp structs.GetAccountDelegationsResponse, err error)
}

type OutputSender interface {
	Send(cStructs.TaskResponse) error
}

// IndexerClient is implementation of a client (main worker code)
type IndexerClient struct {
	grpc GRPC

	logger  *zap.Logger
	streams map[uuid.UUID]*cStructs.StreamAccess
	sLock   sync.Mutex

	maximumHeightsToGet uint64
}

// NewIndexerClient is IndexerClient constructor
func NewIndexerClient(ctx context.Context, logger *zap.Logger, grpc GRPC, maximumHeightsToGet uint64) *IndexerClient {
	return &IndexerClient{
		logger:              logger,
		grpc:                grpc,
		maximumHeightsToGet: maximumHeightsToGet,
		streams:             make(map[uuid.UUID]*cStructs.StreamAccess),
	}
}

// CloseStream removes stream from worker/client
func (ic *IndexerClient) CloseStream(ctx context.Context, streamID uuid.UUID) error {
	ic.sLock.Lock()
	defer ic.sLock.Unlock()

	ic.logger.Debug("[COSMOS-CLIENT] Close Stream", zap.Stringer("streamID", streamID))
	delete(ic.streams, streamID)

	return nil
}

// RegisterStream adds new listeners to the streams - currently fixed number per stream
func (ic *IndexerClient) RegisterStream(ctx context.Context, stream *cStructs.StreamAccess) error {
	ic.logger.Debug("[COSMOS-CLIENT] Register Stream", zap.Stringer("streamID", stream.StreamID))

	ic.sLock.Lock()
	defer ic.sLock.Unlock()
	ic.streams[stream.StreamID] = stream

	// Limit workers not to create new goroutines over and over again
	for i := 0; i < 20; i++ {
		go ic.Run(ctx, stream)
	}

	return nil
}

// Run listens on the stream events (new tasks)
func (ic *IndexerClient) Run(ctx context.Context, stream *cStructs.StreamAccess) {
	for {
		select {
		case <-ctx.Done():
			ic.sLock.Lock()
			delete(ic.streams, stream.StreamID)
			ic.sLock.Unlock()
			return
		case <-stream.Finish:
			return
		case taskRequest := <-stream.RequestListener:
			tctx, cancel := context.WithTimeout(ctx, time.Minute*10)
			switch taskRequest.Type {
			case mStructs.ReqIDGetTransactions:
				ic.GetTransactions(tctx, taskRequest, stream, ic.grpc)
			case mStructs.ReqIDLatestData:
				ic.GetLatest(tctx, taskRequest, stream, ic.grpc)
			case mStructs.ReqIDGetReward:
				ic.GetReward(tctx, taskRequest, stream, ic.grpc)
			case mStructs.ReqIDAccountBalance:
				ic.GetAccountBalance(tctx, taskRequest, stream, ic.grpc)
			case mStructs.ReqIDAccountDelegations:
				ic.GetAccountDelegations(tctx, taskRequest, stream, ic.grpc)
			default:
				stream.Send(cStructs.TaskResponse{
					Id:    taskRequest.Id,
					Error: cStructs.TaskError{Msg: "There is no such handler " + taskRequest.Type},
					Final: true,
				})
			}
			cancel()
		}
	}
}

// GetTransactions gets new transactions and blocks from cosmos for given range
func (ic *IndexerClient) GetTransactions(ctx context.Context, tr cStructs.TaskRequest, stream OutputSender, client GRPC) {
	timer := metrics.NewTimer(getTransactionDuration)
	defer timer.ObserveDuration()

	hr := &structs.HeightRange{}
	err := json.Unmarshal(tr.Payload, hr)
	if err != nil {
		ic.logger.Debug("[COSMOS-CLIENT] Cannot unmarshal payload", zap.String("contents", string(tr.Payload)))
		stream.Send(cStructs.TaskResponse{
			Id:    tr.Id,
			Error: cStructs.TaskError{Msg: "cannot unmarshal payload: " + err.Error()},
			Final: true,
		})
		return
	}

	if hr.EndHeight == 0 {
		stream.Send(cStructs.TaskResponse{
			Id:    tr.Id,
			Error: cStructs.TaskError{Msg: "end height is zero" + err.Error()},
			Final: true,
		})
		return
	}

	sCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := make(chan cStructs.OutResp, page*2+1)
	fin := make(chan bool, 2)

	// (lukanus): in separate goroutine take transaction format wrap it in transport message and send
	go sendResp(sCtx, tr.Id, out, ic.logger, stream, fin)

	if err := getRange(sCtx, ic.logger, client, *hr, out); err != nil {
		stream.Send(cStructs.TaskResponse{
			Id:    tr.Id,
			Error: cStructs.TaskError{Msg: err.Error()},
			Final: true,
		})
		ic.logger.Error("[COSMOS-CLIENT] Error getting range (Get Transactions) ", zap.Error(err), zap.Stringer("taskID", tr.Id))
		close(out)
		return
	}
	close(out)

	for {
		select {
		case <-ctx.Done():
			return
		case <-fin:
			ic.logger.Debug("[COSMOS-CLIENT] Finished sending all", zap.Stringer("taskID", tr.Id), zap.Any("heights", hr))
			return
		}
	}
}

// GetBlock gets block
func (ic *IndexerClient) GetBlock(ctx context.Context, tr cStructs.TaskRequest, stream *cStructs.StreamAccess, client GRPC) {
	timer := metrics.NewTimer(getBlockDuration)
	defer timer.ObserveDuration()

	hr := &structs.HeightHash{}
	err := json.Unmarshal(tr.Payload, hr)
	if err != nil {
		stream.Send(cStructs.TaskResponse{
			Id:    tr.Id,
			Error: cStructs.TaskError{Msg: "Cannot unmarshal payload"},
			Final: true,
		})
		return
	}

	block, err := client.GetBlock(ctx, *hr)
	if err != nil {
		ic.logger.Error("Error getting block", zap.Error(err))
		stream.Send(cStructs.TaskResponse{
			Id:    tr.Id,
			Error: cStructs.TaskError{Msg: "Error getting block data " + err.Error()},
			Final: true,
		})
		return
	}

	out := make(chan cStructs.OutResp, 1)
	out <- cStructs.OutResp{
		ID:      tr.Id,
		Type:    "Block",
		Payload: block,
	}
	close(out)

	sendResp(ctx, tr.Id, out, ic.logger, stream, nil)
}

// GetAccountBalance gets account balance
func (ic *IndexerClient) GetAccountBalance(ctx context.Context, tr cStructs.TaskRequest, stream *cStructs.StreamAccess, client GRPC) {
	timer := metrics.NewTimer(getAccountBalanceDuration)
	defer timer.ObserveDuration()

	ha := &structs.HeightAccount{}
	err := json.Unmarshal(tr.Payload, ha)
	if err != nil {
		stream.Send(cStructs.TaskResponse{
			Id:    tr.Id,
			Error: cStructs.TaskError{Msg: "Cannot unmarshal payload"},
			Final: true,
		})
		return
	}

	sCtx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	blnc, err := client.GetAccountBalance(sCtx, *ha)
	if err != nil {
		ic.logger.Error("Error getting account balance", zap.Error(err))
		stream.Send(cStructs.TaskResponse{
			Id:    tr.Id,
			Error: cStructs.TaskError{Msg: "Error getting account balance data " + err.Error()},
			Final: true,
		})
		return
	}

	out := make(chan cStructs.OutResp, 1)
	out <- cStructs.OutResp{
		ID:      tr.Id,
		Type:    "AccountBalance",
		Payload: blnc,
	}
	close(out)

	sendResp(ctx, tr.Id, out, ic.logger, stream, nil)
}

// GetAccountDelegations gets account delegations
func (ic *IndexerClient) GetAccountDelegations(ctx context.Context, tr cStructs.TaskRequest, stream *cStructs.StreamAccess, client GRPC) {
	timer := metrics.NewTimer(getAccountDelegationsDuration)
	defer timer.ObserveDuration()

	ha := &structs.HeightAccount{}
	err := json.Unmarshal(tr.Payload, ha)
	if err != nil {
		stream.Send(cStructs.TaskResponse{
			Id:    tr.Id,
			Error: cStructs.TaskError{Msg: "Cannot unmarshal payload"},
			Final: true,
		})
		return
	}

	sCtx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	blnc, err := client.GetAccountDelegations(sCtx, *ha)
	if err != nil {
		ic.logger.Error("Error getting account balance", zap.Error(err))
		stream.Send(cStructs.TaskResponse{
			Id:    tr.Id,
			Error: cStructs.TaskError{Msg: "Error getting account balance data " + err.Error()},
			Final: true,
		})
		return
	}

	out := make(chan cStructs.OutResp, 1)
	out <- cStructs.OutResp{
		ID:      tr.Id,
		Type:    "AccountDelegations",
		Payload: blnc,
	}
	close(out)

	sendResp(ctx, tr.Id, out, ic.logger, stream, nil)
}

// GetReward gets reward
func (ic *IndexerClient) GetReward(ctx context.Context, tr cStructs.TaskRequest, stream *cStructs.StreamAccess, client GRPC) {
	timer := metrics.NewTimer(getRewardDuration)
	defer timer.ObserveDuration()

	ha := &structs.HeightAccount{}
	err := json.Unmarshal(tr.Payload, ha)
	if err != nil {
		stream.Send(cStructs.TaskResponse{
			Id:    tr.Id,
			Error: cStructs.TaskError{Msg: "Cannot unmarshal payload"},
			Final: true,
		})
		return
	}

	sCtx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	reward, err := client.GetReward(sCtx, *ha)
	if err != nil {
		ic.logger.Error("Error getting reward", zap.Error(err))
		stream.Send(cStructs.TaskResponse{
			Id:    tr.Id,
			Error: cStructs.TaskError{Msg: "Error getting reward data " + err.Error()},
			Final: true,
		})
		return
	}

	out := make(chan cStructs.OutResp, 1)
	out <- cStructs.OutResp{
		ID:      tr.Id,
		Type:    "Reward",
		Payload: reward,
	}
	close(out)

	sendResp(ctx, tr.Id, out, ic.logger, stream, nil)
}

// GetLatest gets latest transactions and blocks.
// It gets latest transaction, then diff it with
func (ic *IndexerClient) GetLatest(ctx context.Context, tr cStructs.TaskRequest, stream *cStructs.StreamAccess, client GRPC) {
	timer := metrics.NewTimer(getLatestDuration)
	defer timer.ObserveDuration()

	ldr := &structs.LatestDataRequest{}
	err := json.Unmarshal(tr.Payload, ldr)
	if err != nil {
		stream.Send(cStructs.TaskResponse{Id: tr.Id, Error: cStructs.TaskError{Msg: "Cannot unmarshal payload"}, Final: true})
	}

	sCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// (lukanus): Get latest block (height = 0)
	block, err := client.GetBlock(sCtx, structs.HeightHash{})
	if err != nil {
		stream.Send(cStructs.TaskResponse{Id: tr.Id, Error: cStructs.TaskError{Msg: "Error getting block data " + err.Error()}, Final: true})
		return
	}

	hr := getLastHeightRange(ldr.LastHeight, ic.maximumHeightsToGet, block.Height)

	out := make(chan cStructs.OutResp, page*2+1)
	fin := make(chan bool, 2)

	// (lukanus): in separate goroutine take transaction format wrap it in transport message and send
	go sendResp(sCtx, tr.Id, out, ic.logger, stream, fin)

	ic.logger.Debug("[COSMOS-CLIENT] Getting Range", zap.Stringer("taskID", tr.Id), zap.Uint64("start", hr.StartHeight), zap.Uint64("end", hr.EndHeight))
	if err := getRange(sCtx, ic.logger, ic.grpc, hr, out); err != nil {
		stream.Send(cStructs.TaskResponse{
			Id:    tr.Id,
			Error: cStructs.TaskError{Msg: err.Error()},
			Final: true,
		})
		ic.logger.Error("[COSMOS-CLIENT] Error getting range (Get Transactions) ", zap.Error(err), zap.Stringer("taskID", tr.Id))
		close(out)
		return
	}
	close(out)

	for {
		select {
		case <-sCtx.Done():
			return
		case <-fin:
			ic.logger.Debug("[COSMOS-CLIENT] Finished sending all", zap.Stringer("taskID", tr.Id), zap.Any("heights", hr))
			return
		}
	}
}

// getLastHeightRange - based current state
func getLastHeightRange(lastKnownHeight, maximumHeightsToGet, lastBlockFromNetwork uint64) structs.HeightRange {
	// (lukanus): When nothing is scraped we want to get only X number of last requests
	if lastKnownHeight == 0 {
		lastX := lastBlockFromNetwork - maximumHeightsToGet
		if lastX > 0 {
			return structs.HeightRange{
				StartHeight: lastX,
				EndHeight:   lastBlockFromNetwork,
			}
		}
	}

	if maximumHeightsToGet < lastBlockFromNetwork-lastKnownHeight {
		return structs.HeightRange{
			StartHeight: lastBlockFromNetwork - maximumHeightsToGet,
			EndHeight:   lastBlockFromNetwork,
		}
	}

	return structs.HeightRange{
		StartHeight: lastKnownHeight,
		EndHeight:   lastBlockFromNetwork,
	}
}

func blockAndTx(ctx context.Context, logger *zap.Logger, client GRPC, height uint64) (block structs.Block, txs []structs.Transaction, err error) {
	defer logger.Sync()
	logger.Debug("[COSMOS-CLIENT] Getting block", zap.Uint64("block", height))
	block, err = client.GetBlock(ctx, structs.HeightHash{Height: uint64(height)})
	if err != nil {
		logger.Debug("[COSMOS-CLIENT] Err Getting block", zap.Uint64("block", height), zap.Error(err), zap.Uint64("txs", block.NumberOfTransactions))
		return block, nil, fmt.Errorf("error fetching block: %d %w ", uint64(height), err)
	}

	if block.NumberOfTransactions > 0 {
		logger.Debug("[COSMOS-CLIENT] Getting txs", zap.Uint64("block", height), zap.Uint64("txs", block.NumberOfTransactions))
		txs, err = client.SearchTx(ctx, structs.HeightHash{Height: height}, block, page)

		logger.Debug("[COSMOS-CLIENT] txErr Getting txs", zap.Uint64("block", height), zap.Error(err), zap.Uint64("txs", block.NumberOfTransactions))
	}

	logger.Debug("[COSMOS-CLIENT] Got block", zap.Uint64("block", height), zap.Uint64("txs", block.NumberOfTransactions))
	return block, txs, err
}

func asyncBlockAndTx(ctx context.Context, logger *zap.Logger, wg *sync.WaitGroup, client GRPC, cinn chan hBTx) {
	defer wg.Done()
	for in := range cinn {
		b, txs, err := blockAndTx(ctx, logger, client, in.Height)
		if err != nil {
			in.Ch <- cStructs.OutResp{
				ID:    b.ID,
				Error: err,
				Type:  "Error",
			}
			return
		}
		in.Ch <- cStructs.OutResp{
			ID:      b.ID,
			Type:    "Block",
			Payload: b,
		}
		if txs != nil {
			for _, t := range txs {
				in.Ch <- cStructs.OutResp{
					ID:      t.ID,
					Type:    "Transaction",
					Payload: t,
				}
			}
		}

		in.Ch <- cStructs.OutResp{
			ID:   b.ID,
			Type: "Partial",
		}
	}
}

type hBTx struct {
	Height uint64
	Last   bool
	Ch     chan cStructs.OutResp
}

// getRange gets given range of blocks and transactions
func getRange(ctx context.Context, logger *zap.Logger, client GRPC, hr structs.HeightRange, out chan cStructs.OutResp) (err error) {
	defer logger.Sync()

	chIn := oHBTxPool.Get()
	chOut := oHBTxPool.Get()

	errored := make(chan bool, 7)
	defer close(errored)

	wg := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go asyncBlockAndTx(ctx, logger, wg, client, chIn)
	}
	go populateRange(chIn, chOut, hr, errored)

RANGE_LOOP:
	for {
		select {
		// (lukanus): add timeout
		case o := <-chOut:
			if o.Last {
				logger.Debug("[COSMOS-CLIENT] Finished sending height", zap.Uint64("height", o.Height))
				break RANGE_LOOP
			}

		INNER_LOOP:
			for resp := range o.Ch {
				switch resp.Type {
				case "Partial":
					break INNER_LOOP
				case "Error":
					errored <- true // (lukanus): to close publisher and asyncBlockAndTx
					err = resp.Error
					out <- resp
					break RANGE_LOOP
				default:
					out <- resp
				}
			}
			oRespPool.Put(o.Ch)
		}
	}

	if err != nil { // (lukanus): discard everything on error, after error
		wg.Wait() // (lukanus): make sure there are no outstanding producers
	PURIFY_CHANNELS:
		for {
			select {
			case o := <-chOut:
				if o.Ch != nil {
				PURIFY_INNER_CHANNELS:
					for {
						select {
						case <-o.Ch:
						default:
							break PURIFY_INNER_CHANNELS
						}
					}
				}
				oRespPool.Put(o.Ch)
			default:
				break PURIFY_CHANNELS
			}
		}
	}
	oHBTxPool.Put(chOut)
	return err
}

func populateRange(in, out chan hBTx, hr structs.HeightRange, er chan bool) {
	height := hr.StartHeight

	for {
		hBTxO := hBTx{Height: height, Ch: oRespPool.Get()}
		select {
		case out <- hBTxO:
		case <-er:
			break
		}

		select {
		case in <- hBTxO:
		case <-er:
			break
		}

		height++
		if height > hr.EndHeight {
			select {
			case out <- hBTx{Last: true}:
			case <-er:
			}
			break
		}

	}
	close(in)
}

// sendResp constructs protocol response and send it out to transport
func sendResp(ctx context.Context, id uuid.UUID, in <-chan cStructs.OutResp, logger *zap.Logger, sender OutputSender, fin chan bool) {
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	order := uint64(0)

	var contextDone bool

SendLoop:
	for {
		select {
		case <-ctx.Done():
			contextDone = true
			break SendLoop
		case t, ok := <-in:
			if !ok && t.Type == "" {
				break SendLoop
			}
			b.Reset()

			err := enc.Encode(t.Payload)
			if err != nil {
				logger.Error("[COSMOS-CLIENT] Error encoding payload data", zap.Error(err))
			}

			tr := cStructs.TaskResponse{
				Id:      id,
				Type:    t.Type,
				Order:   order,
				Payload: make([]byte, b.Len()),
			}

			b.Read(tr.Payload)
			order++
			err = sender.Send(tr)
			if err != nil {
				logger.Error("[COSMOS-CLIENT] Error sending data", zap.Error(err))
			}
		}
	}

	err := sender.Send(cStructs.TaskResponse{
		Id:    id,
		Type:  "END",
		Order: order,
		Final: true,
	})

	if err != nil {
		logger.Error("[COSMOS-CLIENT] Error sending end", zap.Error(err))
	}

	if fin != nil {
		if !contextDone {
			fin <- true
		}
		close(fin)
	}

}
