package connectivity

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/figment-networks/graph-demo/manager/connectivity/structs"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrStreamOffline       = errors.New("stream is Offline")
	ErrNoWorkersAvailable  = errors.New("no workers available")
	ErrWorkerDoesNotExists = errors.New("workers does not exists")
)

type RoundRobinWorkers struct {
	trws map[string]*structs.TaskWorkerRecord

	next chan *structs.TaskWorkerRecord
	lock sync.RWMutex
}

func NewRoundRobinWorkers() *RoundRobinWorkers {
	return &RoundRobinWorkers{
		next: make(chan *structs.TaskWorkerRecord, 100),
		trws: make(map[string]*structs.TaskWorkerRecord),
	}
}

func (rrw *RoundRobinWorkers) AddWorker(id string, stream *structs.StreamAccess) error {
	rrw.lock.Lock()
	defer rrw.lock.Unlock()

	trw := &structs.TaskWorkerRecord{WorkerID: id, Stream: stream}
	select {
	case rrw.next <- trw:
		rrw.trws[id] = trw
	default:
		return ErrNoWorkersAvailable
	}

	return nil
}

func (rrw *RoundRobinWorkers) SendNext(tr structs.TaskRequest, aw *structs.Await) (failedWorkerID string, err error) {
	rrw.lock.RLock()
	defer rrw.lock.RUnlock()

	select {
	case twr := <-rrw.next:
		if err := twr.Stream.Req(tr, aw); err != nil {
			return twr.WorkerID, fmt.Errorf("Cannot send to worker channel: %w", err)
		}

		twr.L.Lock()
		twr.LastSend = time.Now()
		twr.L.Unlock()
		rrw.next <- twr
	default:
		return "", ErrNoWorkersAvailable
	}

	return "", nil
}

// GetWorkers returns current workers information
func (rrw *RoundRobinWorkers) GetWorkers() structs.TaskWorkerRecordInfo {
	rrw.lock.RLock()
	defer rrw.lock.RUnlock()

	twri := structs.TaskWorkerRecordInfo{
		All:    len(rrw.trws),
		Active: len(rrw.next),
	}
	for _, w := range rrw.trws {
		twri.Workers = append(twri.Workers, getWorker(w))
	}

	return twri
}

func (rrw *RoundRobinWorkers) GetWorker(id string) (twi structs.TaskWorkerInfo, ok bool) {
	rrw.lock.RLock()
	defer rrw.lock.RUnlock()

	t, ok := rrw.trws[id]
	if !ok {
		return twi, ok
	}

	return getWorker(t), ok
}

func getWorker(w *structs.TaskWorkerRecord) structs.TaskWorkerInfo {

	w.L.RLock()
	twr := structs.TaskWorkerInfo{
		WorkerID: w.WorkerID,
		LastSend: w.LastSend,
	}
	w.L.RUnlock()

	if w.Stream != nil {
		w.Stream.MapLock.RLock()
		twr.StreamID = w.Stream.StreamID
		twr.StreamState = w.Stream.State

		twr.ResponseMap = map[uuid.UUID]structs.TaskWorkerInfoAwait{}

		for k, r := range w.Stream.ResponseMap {
			if r != nil {
				twia := structs.TaskWorkerInfoAwait{
					Created:        r.Created,
					State:          r.State,
					ReceivedFinals: r.ReceivedFinals,
					Uids:           make([]uuid.UUID, len(r.Uids)),
				}
				copy(twia.Uids, r.Uids)
				twr.ResponseMap[k] = twia
			}
		}
		w.Stream.MapLock.RUnlock()
	}
	return twr
}

func (rrw *RoundRobinWorkers) Ping(ctx context.Context, id string) (time.Duration, error) {
	rrw.lock.RLock()
	t, ok := rrw.trws[id]
	rrw.lock.RUnlock()
	if !ok {
		return 0, ErrWorkerDoesNotExists
	}

	return t.Stream.Ping(ctx)
}

// BringOnline Brings worker Online, removing duplicates
func (rrw *RoundRobinWorkers) BringOnline(id string) error {
	rrw.lock.Lock()
	defer rrw.lock.Unlock()
	t, ok := rrw.trws[id]
	if !ok {
		return ErrWorkerDoesNotExists
	}
	// (lukanus): Remove duplicates
	removeFromChannel(rrw.next, id)

	select {
	case rrw.next <- t:
	default:
		return ErrNoWorkersAvailable
	}

	return nil
}

// SendToWoker sends task to worker
func (rrw *RoundRobinWorkers) SendToWorker(id string, tr structs.TaskRequest, aw *structs.Await) error {
	rrw.lock.RLock()
	twr, ok := rrw.trws[id]
	rrw.lock.RUnlock()
	if !ok {
		return ErrWorkerDoesNotExists
	}

	if err := twr.Stream.Req(tr, aw); err != nil {
		return fmt.Errorf("Cannot send to worker channel: %w", err)
	}
	twr.LastSend = time.Now()
	return nil
}

// Reconnect reconnects stream if exists
func (rrw *RoundRobinWorkers) Reconnect(ctx context.Context, logger *zap.Logger, id string) error {
	rrw.lock.RLock()
	t, ok := rrw.trws[id]
	rrw.lock.RUnlock()
	if !ok {
		return ErrWorkerDoesNotExists
	}
	return t.Stream.Reconnect(ctx, logger)
}

// Close closes worker of given id
func (rrw *RoundRobinWorkers) Close(id string) error {
	rrw.lock.Lock()
	defer rrw.lock.Unlock()

	t, ok := rrw.trws[id]
	if !ok {
		return ErrWorkerDoesNotExists
	}

	removeFromChannel(rrw.next, id)
	// t.stream.State = structs.StreamOffline

	return t.Stream.Close()
}

func removeFromChannel(next chan *structs.TaskWorkerRecord, id string) {
	inCh := len(next)
	for i := 0; i < inCh; i++ {
		w := <-next
		if w.WorkerID == id {
			continue
		}
		next <- w
	}
}
