package structs

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ErrStreamNotOnline = errors.New("Stream is not Online")

type StreamState int

const (
	StreamUnknown StreamState = iota
	StreamOnline
	StreamError
	StreamReconnecting
	StreamClosing
	StreamOffline
)

type ConnTransport interface {
	Run(ctx context.Context, logger *zap.Logger, stream *StreamAccess)
	Type() string
}

type WorkerCompositeKey struct {
	Network string
	ChainID string
	Version string
}

type SimpleWorkerInfo struct {
	Network string
	Version string
	ChainID string
	ID      string
}

type WorkerInfo struct {
	NodeSelfID     string             `json:"node_id"`
	Type           string             `json:"type"`
	ChainID        string             `json:"chain_id"`
	State          StreamState        `json:"state"`
	ConnectionInfo []WorkerConnection `json:"connection"`
	LastCheck      time.Time          `json:"last_check"`

	L sync.RWMutex `json:"-"`
}

func (wi *WorkerInfo) LastChecked() {
	wi.L.Lock()
	defer wi.L.Unlock()
	wi.LastCheck = time.Now()
}

func (wi *WorkerInfo) Clone() *WorkerInfo {
	wi.L.RLock()
	defer wi.L.RUnlock()

	wiC := &WorkerInfo{
		NodeSelfID: wi.NodeSelfID,
		Type:       wi.Type,
		ChainID:    wi.ChainID,
		State:      wi.State,
		LastCheck:  wi.LastCheck,
	}
	if len(wi.ConnectionInfo) > 0 {
		wiC.ConnectionInfo = make([]WorkerConnection, len(wi.ConnectionInfo))
		copy(wiC.ConnectionInfo, wi.ConnectionInfo)
	}
	return wiC
}

func (wi *WorkerInfo) SetState(ss StreamState) {
	wi.L.Lock()
	defer wi.L.Unlock()
	wi.State = ss
}

type WorkerConnection struct {
	Version   string          `json:"version"`
	Type      string          `json:"type"`
	Addresses []WorkerAddress `json:"addresses"`
}

type WorkerAddress struct {
	IP      net.IP `json:"ip"`
	Address string `json:"address"`
}

type TaskRequest struct {
	ID      uuid.UUID
	Network string
	ChainID string
	Version string

	Type    string
	Payload json.RawMessage
}

type TaskErrorType string

type TaskError struct {
	Msg  string
	Type TaskErrorType
}

type TaskResponse struct {
	ID      uuid.UUID
	Version string
	Type    string
	Order   int64
	Final   bool
	Error   TaskError
	Payload json.RawMessage
}

type ClientControl struct {
	Type string
	Resp chan ClientResponse
}

type ClientResponse struct {
	OK    bool
	Error string
	Time  time.Duration
}

type Await struct {
	sync.RWMutex
	Created        time.Time
	State          StreamState
	Resp           chan *TaskResponse
	ReceivedFinals int
	Uids           []uuid.UUID
}

func NewAwait(sendIDs []uuid.UUID) (aw *Await) {
	return &Await{
		Created: time.Now(),
		State:   StreamOnline,
		Uids:    sendIDs,
		Resp:    make(chan *TaskResponse, 400),
	}
}

func (aw *Await) Send(tr *TaskResponse) (bool, error) {
	aw.RLock()
	defer aw.RUnlock()

	if aw.State != StreamOnline {
		return false, errors.New("Cannot send recipient unavailable")
	}

	if tr.Final {
		aw.ReceivedFinals++
	}

	aw.Resp <- tr

	if len(aw.Uids) == aw.ReceivedFinals {
		// (lukanus)\ Temporary supressed
		// log.Printf("Received All %s ", time.Now().Sub(aw.Created).String())
		return true, nil
	}

	return false, nil
}

func (aw *Await) SetState(ss StreamState) {
	aw.Lock()
	defer aw.Unlock()
	aw.State = ss
}

func (aw *Await) Close() {
	aw.Lock()
	defer aw.Unlock()
	if aw.State != StreamOffline {
		return
	}

	aw.State = StreamOffline

Drain:
	for {
		select {
		case <-aw.Resp:
		default:
			break Drain
		}
	}
	close(aw.Resp)
	aw.Resp = nil
}

type IndexerClienter interface {
	RegisterStream(ctx context.Context, stream *StreamAccess) error
}

type StreamAccess struct {
	State           StreamState
	StreamID        uuid.UUID
	ResponseMap     map[uuid.UUID]*Await
	RequestListener chan TaskRequest

	ManagerID string

	ClientControl chan ClientControl

	CancelConnection context.CancelFunc

	WorkerInfo *WorkerInfo
	Transport  ConnTransport

	respLock sync.RWMutex
	reqLock  sync.RWMutex

	MapLock sync.RWMutex
}

func NewStreamAccess(transport ConnTransport, managerID string, conn *WorkerInfo) *StreamAccess {
	sID, _ := uuid.NewRandom()

	return &StreamAccess{
		StreamID:  sID,
		State:     StreamUnknown,
		ManagerID: managerID,

		Transport:     transport,
		WorkerInfo:    conn,
		ClientControl: make(chan ClientControl, 5),

		ResponseMap:     make(map[uuid.UUID]*Await),
		RequestListener: make(chan TaskRequest, 100),
	}
}

func (sa *StreamAccess) Run(ctx context.Context, logger *zap.Logger) error {
	sa.MapLock.Lock()
	if sa.State == StreamReconnecting {
		sa.MapLock.Unlock()
		return errors.New("already reconnecting")
	}

	sa.State = StreamReconnecting
	sa.MapLock.Unlock()

	if sa.CancelConnection != nil {
		sa.CancelConnection()
	}

	var nCtx context.Context
	nCtx, sa.CancelConnection = context.WithCancel(ctx)
	go sa.Transport.Run(nCtx, logger, sa)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := sa.Ping(ctx)

	return err
}

func (sa *StreamAccess) Reconnect(ctx context.Context, logger *zap.Logger) error {
	return sa.Run(ctx, logger)
}

func (sa *StreamAccess) Recv(tr *TaskResponse) error {

	sa.respLock.RLock()
	defer sa.respLock.RUnlock()

	sa.MapLock.Lock()
	resAwait, ok := sa.ResponseMap[tr.ID]
	sa.MapLock.Unlock()

	if !ok {
		return errors.New("No such requests registered")
	}

	all, err := resAwait.Send(tr)
	if err != nil {
		resAwait.SetState(StreamError)
		return err
	}

	sa.MapLock.Lock()
	resAwait.SetState(StreamOnline)

	if all {
		for _, u := range resAwait.Uids {
			delete(sa.ResponseMap, u)
		}
	}
	sa.MapLock.Unlock()
	return nil
}

func (sa *StreamAccess) Req(tr TaskRequest, aw *Await) error {
	sa.reqLock.RLock()
	defer sa.reqLock.RUnlock()

	if sa.State != StreamOnline {
		return ErrStreamNotOnline
	}

	sa.MapLock.Lock()
	sa.ResponseMap[tr.ID] = aw
	sa.MapLock.Unlock()

	sa.RequestListener <- tr

	return nil
}

func (sa *StreamAccess) Ping(ctx context.Context) (time.Duration, error) {
	resp := make(chan ClientResponse, 1) //(lukanus): this can be only closed after write

	select {
	case sa.ClientControl <- ClientControl{
		Type: "PING",
		Resp: resp,
	}:
	default:
		return 0, errors.New("Cannot send PING")
	}

	for {
		select {
		case <-ctx.Done(): // timeout
			/*	sa.reqLock.Lock()
				sa.respLock.Lock()
				defer sa.respLock.Unlock()
				defer sa.reqLock.Unlock()*/
			sa.WorkerInfo.SetState(StreamOffline)

			sa.MapLock.Lock()
			sa.State = StreamOffline
			sa.MapLock.Unlock()
			return 0, errors.New("Ping TIMEOUT")
		case a := <-resp:
			/*	sa.reqLock.Lock()
				sa.respLock.Lock()
				defer sa.respLock.Unlock()
				defer sa.reqLock.Unlock()*/

			sa.WorkerInfo.SetState(StreamOnline)

			sa.MapLock.Lock()
			sa.State = StreamOnline
			sa.MapLock.Unlock()
			return a.Time, nil
		}
	}
}

func (sa *StreamAccess) Close() error {
	sa.reqLock.Lock()
	sa.respLock.Lock()
	defer sa.respLock.Unlock()
	defer sa.reqLock.Unlock()

	if sa.State == StreamOffline {
		return nil
	}

	sa.MapLock.Lock()
	sa.State = StreamOffline
	sa.CancelConnection()

	if sa.WorkerInfo != nil {
		sa.WorkerInfo.SetState(StreamOffline)
	}

	for id, aw := range sa.ResponseMap {
		// (lukanus): Close all awaits, they won't get full response anyway.
		aw.Close()
		delete(sa.ResponseMap, id)
	}
	sa.MapLock.Unlock()

	return nil
}
