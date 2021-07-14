package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/figment-networks/graph-demo/manager/connectivity/structs"
	shared "github.com/figment-networks/graph-demo/manager/structs"
	"go.uber.org/zap"
)

type PSig struct {
	TaskID  string
	Network string
	ChainID string
	Version string
}

func NewRunning() *Running {
	return &Running{
		ProcessesSyncData:   make(map[PSig]ProcessSyncData),
		ProcessesLatestData: make(map[PSig]ProcessLatestData),
	}
}

type Running struct {
	lock sync.RWMutex

	ProcessesSyncData   map[PSig]ProcessSyncData
	ProcessesLatestData map[PSig]ProcessLatestData
}

type ProcessLatestData struct {
	Started   time.Time
	Finished  bool
	EndHeight uint64
	Error     error
	Resp      shared.LatestDataResponse
	Retries   uint64
}

type ProcessSyncData struct {
	Started   time.Time
	Finished  bool
	EndHeight uint64
	Error     error
	Resp      shared.SyncDataResponse
	Retries   uint64
}

func (hc *Client) SyncData(ctx context.Context, ldr shared.SyncDataRequest) (ldResp shared.SyncDataResponse, er error) {
	defer hc.recoverPanic()

	key := PSig{TaskID: ldr.TaskID, Network: ldr.Network, ChainID: ldr.ChainID, Version: ldr.Version}
	ldResp = shared.SyncDataResponse{}

	hc.logger.Debug("[Client] SyncData", zap.Any("received", ldr))

	hc.r.lock.Lock()
	p, ok := hc.r.ProcessesSyncData[key]
	hc.r.lock.Unlock()

	if ok {
		if p.Finished {
			hc.r.lock.Lock()
			delete(hc.r.ProcessesSyncData, key)
			hc.r.lock.Unlock()
			return p.Resp, p.Error
		}

		if p.Error != nil {
			hc.logger.Warn("[CLIENT] Last request errored", zap.Error(p.Error), zap.Uint64("height", ldr.LastHeight), zap.Duration("since", time.Since(p.Started)))
			hc.r.lock.Lock()
			delete(hc.r.ProcessesSyncData, key)
			hc.r.lock.Unlock()
			ldResp.LastHeight = ldr.LastHeight
			return ldResp, p.Error
		}

		hc.logger.Warn("[CLIENT] Last request is still processing", zap.Uint64("height", ldr.LastHeight), zap.Uint64("retry", p.Retries), zap.Duration("since", time.Since(p.Started)))
		ldResp.LastHeight = ldr.LastHeight
		p.Retries++
		hc.r.lock.Lock()
		hc.r.ProcessesSyncData[key] = p
		hc.r.lock.Unlock()
		if p.Retries > 20 {
			hc.logger.Warn("[CLIENT] still processing limit, retrying", zap.Uint64("height", ldr.LastHeight), zap.Duration("since", time.Since(p.Started)))
			hc.r.lock.Lock()
			delete(hc.r.ProcessesLatestData, key)
			hc.r.lock.Unlock()
		}

		return ldResp, p.Error
	}

	hR := getSyncRangeHeight(ldr.LastHeight, 300, ldr.FinalHeight)
	sresp, err := hc.syncData(ctx, hR, ldr)

	hc.r.lock.Lock()
	proc, ok := hc.r.ProcessesSyncData[key]
	if ok {
		proc.Resp = sresp
		proc.Finished = true
		if err != nil {
			proc.Error = err
		}
		hc.r.ProcessesSyncData[key] = proc
	}
	hc.r.lock.Unlock()

	return sresp, err
}

func (hc *Client) LatestData(ctx context.Context, ldr shared.LatestDataRequest) (ldResp shared.LatestDataResponse, er error) {
	defer hc.recoverPanic()

	key := PSig{TaskID: ldr.TaskID, Network: ldr.Network, ChainID: ldr.ChainID, Version: ldr.Version}
	ldResp = shared.LatestDataResponse{}

	hc.logger.Debug("[Client] LatestData", zap.Any("received", ldr))

	hc.r.lock.Lock()
	p, ok := hc.r.ProcessesLatestData[key]
	hc.r.lock.Unlock()

	if ok {
		if p.Finished {
			hc.r.lock.Lock()
			delete(hc.r.ProcessesLatestData, key)
			hc.r.lock.Unlock()
			return p.Resp, p.Error
		}

		if p.Error != nil {
			hc.logger.Warn("[CLIENT] Last request errored", zap.Error(p.Error), zap.Uint64("height", ldr.LastHeight), zap.Duration("since", time.Since(p.Started)))
			hc.r.lock.Lock()
			delete(hc.r.ProcessesLatestData, key)
			hc.r.lock.Unlock()
			ldResp.LastHeight = ldr.LastHeight
			return ldResp, p.Error
		}
		hc.logger.Warn("[CLIENT] Last request is still processing", zap.Uint64("height", ldr.LastHeight), zap.Duration("since", time.Since(p.Started)))
		ldResp.LastHeight = ldr.LastHeight

		p.Retries++
		hc.r.lock.Lock()
		hc.r.ProcessesLatestData[key] = p
		hc.r.lock.Unlock()

		if p.Retries > 20 {
			hc.logger.Warn("[CLIENT] still processing limit, retrying", zap.Uint64("height", ldr.LastHeight), zap.Duration("since", time.Since(p.Started)))
			hc.r.lock.Lock()
			delete(hc.r.ProcessesLatestData, key)
			hc.r.lock.Unlock()
		}

		return ldResp, p.Error
	}

	b, _ := json.Marshal(ldr)
	aw, err := hc.sender.Send([]structs.TaskRequest{{
		Network: ldr.Network,
		ChainID: ldr.ChainID,
		Version: ldr.Version,
		Type:    shared.ReqIDGetLatestMark,
		Payload: b,
	}})
	if err != nil {
		return ldResp, fmt.Errorf("error creating call: %w", err)
	}

	latestMark := &shared.LatestDataResponse{}
	select {
	case <-ctx.Done():
		return ldResp, errors.New("request timed out")
	case response := <-aw.Resp:
		hc.logger.Debug("[Client] Get LastMark received data:", zap.Any("response", response))

		if response.Error.Msg != "" {
			hc.logger.Warn("[CLIENT] Error getting last response ", zap.String("error", response.Error.Msg))
			return ldResp, fmt.Errorf("error getting response: %s", response.Error.Msg)
		}

		if response.Type != "LatestMark" {
			return ldResp, fmt.Errorf("error getting wrong LatestMark response type: %s", response.Type)
		}

		if err = json.Unmarshal(response.Payload, latestMark); err != nil {
			return ldResp, fmt.Errorf("error unmarshaling LatestMark response : %s", response.Error.Msg)
		}
	}

	if latestMark.LastHeight == ldr.LastHeight {
		return ldResp, nil // up to date
	}

	hc.r.lock.Lock()
	hc.r.ProcessesLatestData[key] = ProcessLatestData{
		Started:   time.Now(),
		Finished:  false,
		EndHeight: latestMark.LastHeight,
	}
	hc.r.lock.Unlock()

	hR := getLastHeightRange(ldr.LastHeight, 1000, latestMark.LastHeight)
	sresp, err := hc.syncData(ctx, hR, shared.SyncDataRequest{
		Network: ldr.Network,
		ChainID: ldr.ChainID,
		Version: ldr.Version,
		TaskID:  ldr.TaskID,

		LastHeight:  ldr.LastHeight,
		FinalHeight: latestMark.LastHeight,
		LastHash:    ldr.LastHash,
		LastEpoch:   ldr.LastEpoch,
		LastTime:    ldr.LastTime,
		Nonce:       ldr.Nonce,

		RetryCount: ldr.RetryCount,
		SelfCheck:  ldr.SelfCheck,
	})

	fResp := shared.LatestDataResponse{
		LastHeight: sresp.LastHeight,
		LastHash:   sresp.LastHash,
		LastEpoch:  sresp.LastEpoch,
		LastTime:   sresp.LastTime,
		Nonce:      sresp.Nonce,
		Error:      sresp.Error,
	}

	hc.r.lock.Lock()
	proc, ok := hc.r.ProcessesLatestData[key]
	if ok {
		proc.Resp = fResp
		proc.Finished = true
		if err != nil {
			proc.Error = err
		}
		hc.r.ProcessesLatestData[key] = proc
	}
	hc.r.lock.Unlock()

	return fResp, err
}

func (hc *Client) syncData(ctx context.Context, hR shared.HeightRange, sdr shared.SyncDataRequest) (sdResp shared.SyncDataResponse, er error) {
	// (lukanus): Set last initially in case of failure
	sdResp.LastHeight = sdr.LastHeight

	reqs, err := makeRequests(hR, sdr, 20)
	if err != nil {
		return sdResp, fmt.Errorf("error preparing data in syncData: %w", err)
	}
	buff := &bytes.Buffer{}
	dec := json.NewDecoder(buff)
	hc.logger.Debug("REQS", zap.Any("reqs", reqs), zap.Any("sdr", sdr))
	hc.logger.Sync()

	respAwait, err := hc.sender.Send(reqs)
	if err != nil {
		return sdResp, fmt.Errorf("error sending data in getTransaction: %w", err)
	}
	defer respAwait.Close()

	var transactionCounterInt, blockCounterInt uint64

	newSdResp := shared.SyncDataResponse{}
	var blockHeights []uint64
	var errorsAt []uint64
WaitForAllData:
	for {
		select {
		case <-ctx.Done():
			return sdResp, errors.New("request timed out")
		case response := <-respAwait.Resp:
			if response.Type == "Heights" {
				buff.Reset()
				buff.ReadFrom(bytes.NewReader(response.Payload))
				switch response.Type {
				case "Heights":
					higs := &shared.Heights{}
					err := dec.Decode(higs)
					if err != nil && response.Error.Msg == "" {
						return sdResp, fmt.Errorf("error decoding response: %w", err)
					}
					blockHeights = append(blockHeights, higs.Heights...)
					errorsAt = append(errorsAt, higs.ErrorAt...)
					if newSdResp.LastTime.IsZero() || newSdResp.LastHeight <= higs.LatestData.LastHeight {
						newSdResp.LastEpoch = higs.LatestData.LastEpoch
						newSdResp.LastHash = higs.LatestData.LastHash
						newSdResp.LastHeight = higs.LatestData.LastHeight
						newSdResp.LastTime = higs.LatestData.LastTime
					}

					transactionCounterInt += higs.NumberOfTx
					blockCounterInt += higs.NumberOfHeights
				}
			}

			if response.Error.Msg != "" {
				sdResp.Error = []byte(response.Error.Msg)
				err = fmt.Errorf("error getting response: %s", response.Error.Msg)
			}

			if response.Final {
				hc.logger.Info("[Client] Received All syncData",
					zap.Any("network", sdr.Network),
					zap.Any("chain_id", sdr.ChainID),
					zap.Uint64("transaction_count", transactionCounterInt),
					zap.Uint64("block_count", blockCounterInt),
				)
				break WaitForAllData
			}
		}
	}

	if len(errorsAt) > 0 {
		sort.Slice(errorsAt, func(i, j int) bool { return errorsAt[i] >= errorsAt[j] })
		sdResp.LastHeight = errorsAt[0] - 1
	}

	if err != nil {
		return sdResp, err
	}

	sort.Slice(blockHeights, func(i, j int) bool { return blockHeights[i] < blockHeights[j] })
	for i, k := range blockHeights {
		if i > 0 && k != blockHeights[i-1]+1 {
			sdResp.LastHeight = blockHeights[i-1] // move processed to before first missing
			return sdResp, ErrIntegrityCheckFailed
		}
	}

	sdResp.LastEpoch = newSdResp.LastEpoch
	sdResp.LastTime = newSdResp.LastTime
	sdResp.LastHeight = newSdResp.LastHeight
	sdResp.LastHash = newSdResp.LastHash

	return sdResp, nil

}

func makeRequests(hR shared.HeightRange, sdr shared.SyncDataRequest, batchLimit uint64) (req []structs.TaskRequest, err error) {

	diff := float64(hR.EndHeight - hR.StartHeight)
	if diff == 0 {
		if hR.EndHeight == 0 {
			return nil, errors.New("no transaction to get, bad request")
		}

		b, _ := json.Marshal(shared.HeightRange{
			Network:     sdr.Network,
			ChainID:     sdr.ChainID,
			StartHeight: hR.StartHeight,
			EndHeight:   hR.EndHeight,
			Hash:        hR.Hash,
		})

		req = append(req, structs.TaskRequest{
			Network: sdr.Network,
			ChainID: sdr.ChainID,
			Version: sdr.Version,
			Type:    shared.ReqIDGetTransactions,
			Payload: b,
		})
		return req, nil
	}

	var i uint64
	for {
		startH := hR.StartHeight + i*batchLimit
		endH := hR.StartHeight + i*batchLimit + batchLimit - 1

		if hR.EndHeight > 0 && endH > hR.EndHeight {
			endH = hR.EndHeight
		}

		b, _ := json.Marshal(shared.HeightRange{
			StartHeight: startH,
			EndHeight:   endH,
			Hash:        hR.Hash,
			Network:     sdr.Network,
			ChainID:     sdr.ChainID,
		})

		req = append(req, structs.TaskRequest{
			Network: sdr.Network,
			ChainID: sdr.ChainID,
			Version: sdr.Version,
			Type:    shared.ReqIDGetTransactions,
			Payload: b,
		})

		i++
		if hR.EndHeight == endH || hR.EndHeight == 0 {
			break
		}
	}
	return req, nil

}

// getLastHeightRange - based current state
func getLastHeightRange(lastKnownHeight, maximumHeightsToGet, lastBlockFromNetwork uint64) shared.HeightRange {
	// (lukanus): When nothing is scraped we want to get only X number of last requests
	if lastKnownHeight == 0 {
		lastX := lastBlockFromNetwork - maximumHeightsToGet
		if lastX > 0 {
			return shared.HeightRange{
				StartHeight: lastX,
				EndHeight:   lastBlockFromNetwork,
			}
		}
	}

	if maximumHeightsToGet < lastBlockFromNetwork-lastKnownHeight {
		return shared.HeightRange{
			StartHeight: lastBlockFromNetwork - maximumHeightsToGet,
			EndHeight:   lastBlockFromNetwork,
		}
	}

	return shared.HeightRange{
		StartHeight: lastKnownHeight,
		EndHeight:   lastBlockFromNetwork,
	}
}

func getSyncRangeHeight(lastKnownHeight, maximumHeightsToGet, lastBlockInRange uint64) shared.HeightRange {
	if lastKnownHeight+maximumHeightsToGet > lastBlockInRange {
		return shared.HeightRange{
			StartHeight: lastKnownHeight,
			EndHeight:   lastBlockInRange,
		}
	}

	return shared.HeightRange{
		StartHeight: lastKnownHeight,
		EndHeight:   lastKnownHeight + maximumHeightsToGet,
	}
}
