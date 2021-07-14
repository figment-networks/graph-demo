package structs

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/google/uuid"
)

// StreamState the state of stream ;D
type StreamState int

const (
	StreamUnknown StreamState = iota
	StreamOnline
	StreamOffline
)

var ErrStreamIsNotOnline = errors.New("Stream is not Online")

// StreamAccess creates a proxy between code and transport level.
// The extra layer serves a function of access manager.
// Requests and Responses are processed by different goroutines, that doesn't need to know about connection state.
// This code prevents sending messages on closed channels after connection breakage.
type StreamAccess struct {
	Finish           chan bool
	State            StreamState
	StreamID         uuid.UUID
	ResponseListener chan TaskResponse
	respLock         sync.RWMutex

	RequestListener chan TaskRequest
	reqLock         sync.RWMutex
}

// NewStreamAccess is StreamAccess constructor
func NewStreamAccess() *StreamAccess {
	responsesCh := TaskResponseChanPool.Get()
	sID, _ := uuid.NewRandom()

	return &StreamAccess{
		Finish:           make(chan bool),
		State:            StreamOnline,
		StreamID:         sID,
		ResponseListener: responsesCh,
		RequestListener:  make(chan TaskRequest, 30),
	}
}

// Send sends TaskResponse back to manager (thread safe)
func (sa *StreamAccess) Send(tr TaskResponse) error {
	sa.respLock.RLock()
	defer sa.respLock.RUnlock()

	if sa.State != StreamOnline {
		return ErrStreamIsNotOnline
	}

	sa.ResponseListener <- tr
	return nil

}

// Req receive TaskRequest (thread safe)
func (sa *StreamAccess) Req(tr TaskRequest) error {
	sa.reqLock.RLock()
	defer sa.reqLock.RUnlock()
	if sa.State != StreamOnline {
		return ErrStreamIsNotOnline
	}

	sa.RequestListener <- tr
	return nil
}

// Close is closing access stream  (thread safe)
func (sa *StreamAccess) Close() error {
	sa.reqLock.Lock()
	sa.respLock.RLock()
	defer sa.respLock.RUnlock()
	defer sa.reqLock.Unlock()

	if sa.State == StreamOffline {
		return nil
	}
	close(sa.Finish)

	TaskResponseChanPool.Put(sa.ResponseListener)
	close(sa.RequestListener)
	return nil
}

// OutResp is enriched response format used internally in workers
type OutResp struct {
	ID      uuid.UUID
	Type    string
	Payload interface{} // to be encoded
	Error   error
}

// TaskRequest is the incoming request
type TaskRequest struct {
	Id      uuid.UUID
	Type    string
	Payload json.RawMessage
}

// TaskError is basic error format
type TaskError struct {
	Msg string
}

// TaskResponse is task ... response :)
type TaskResponse struct {
	Version string
	Id      uuid.UUID
	Type    string
	Order   uint64
	Final   bool
	Error   TaskError
	Payload json.RawMessage
}

// IndexerClienter is a client interface
type IndexerClienter interface {
	RegisterStream(ctx context.Context, stream *StreamAccess) error
	CloseStream(ctx context.Context, streamID uuid.UUID) error
}
