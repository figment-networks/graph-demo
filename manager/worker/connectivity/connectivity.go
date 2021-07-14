package connectivity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// WorkerConnections is connection controller for worker
type WorkerConnections struct {
	network                 string
	chainID                 string
	version                 string
	workerID                string
	workerAccessibleAddress string
	managerAddresses        map[string]string
	managerAddressesLock    sync.RWMutex
}

// NewWorkerConnections is WorkerConnections constructor
func NewWorkerConnections(id, address, network, chainID, version string) *WorkerConnections {
	return &WorkerConnections{
		network:                 network,
		version:                 version,
		workerID:                id,
		chainID:                 chainID,
		workerAccessibleAddress: address,
		managerAddresses:        make(map[string]string),
	}
}

// AddManager dynamically adds manager to the list
func (wc *WorkerConnections) AddManager(managerAddress string) {
	wc.managerAddressesLock.Lock()
	defer wc.managerAddressesLock.Unlock()

	wc.managerAddresses[managerAddress] = managerAddress
}

// RemoveManager dynamically removes manager to the list
func (wc *WorkerConnections) RemoveManager(managerAddress string) {
	wc.managerAddressesLock.Lock()
	defer wc.managerAddressesLock.Unlock()
	delete(wc.managerAddresses, managerAddress)
}

type WorkerResponse struct {
	ID           string                 `json:"id"`
	Network      string                 `json:"network"`
	ChainID      string                 `json:"chain_id"`
	Connectivity WorkerInfoConnectivity `json:"connectivity"`
}

type WorkerInfoConnectivity struct {
	Version string `json:"version"`
	Type    string `json:"type"`
	Address string `json:"address"`
}

func (wc *WorkerConnections) gerWorkerInfo() WorkerResponse {
	return WorkerResponse{
		ID:      wc.workerID,
		Network: wc.network,
		ChainID: wc.chainID,
		Connectivity: WorkerInfoConnectivity{
			Version: wc.version,
			Type:    "grpc",
			Address: wc.workerAccessibleAddress,
		},
	}
}

// Run controls the registration of worker in manager. Every tick it sends it's identity (with address and network type) to every configured address.
func (wc *WorkerConnections) Run(ctx context.Context, logger *zap.Logger, dur time.Duration) {
	defer logger.Sync()

	tckr := time.NewTicker(dur)
	client := &http.Client{}

	wInfo, _ := json.Marshal(wc.gerWorkerInfo())
	readr := bytes.NewReader(wInfo)

	for {
		select {
		case <-ctx.Done():
			tckr.Stop()
			return
		case <-tckr.C:
			wc.managerAddressesLock.RLock()
			for _, addr := range wc.managerAddresses {
				readr.Seek(0, 0)
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+addr, readr)
				if err != nil {
					logger.Error(fmt.Sprintf("Error creating request %s", err.Error()), zap.String("address", addr))
					continue
				}
				resp, err := client.Do(req)
				if err != nil {
					logger.Error(fmt.Sprintf("Error connecting to manager on %s, %s", addr, err.Error()), zap.String("address", addr))
					continue
				}
				if resp.StatusCode > 399 {
					logger.Error(fmt.Sprintf("Error returned from manager", addr), zap.String("address", addr), zap.String("status", resp.Status))
					continue
				}
				resp.Body.Close()
			}
			wc.managerAddressesLock.RUnlock()
		}
	}
}
