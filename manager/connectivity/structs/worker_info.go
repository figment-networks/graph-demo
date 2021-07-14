package structs

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type WorkerNetworkStatic struct {
	Workers map[string]WorkerInfoStatic `json:"workers"`
	All     int                         `json:"all"`
	Active  int                         `json:"active"`
}

type WorkerInfoStatic struct {
	*WorkerInfo
	TaskWorkerInfo TaskWorkerInfo `json:"tasks"`
}

type TaskWorkerInfo struct {
	WorkerID string    `json:"worker_id"`
	LastSend time.Time `json:"last_send"`

	StreamState StreamState `json:"stream_state"`
	StreamID    uuid.UUID   `json:"stream_id"`
	ResponseMap map[uuid.UUID]TaskWorkerInfoAwait
}

type TaskWorkerInfoAwait struct {
	Created        time.Time   `json:"created"`
	State          StreamState `json:"state"`
	ReceivedFinals int         `json:"received_finals"`
	Uids           []uuid.UUID `json:"expected"`
}

type TaskWorkerRecordInfo struct {
	Workers []TaskWorkerInfo `json:"workers"`
	All     int              `json:"all"`
	Active  int              `json:"active"`
}

type TaskWorkerRecord struct {
	WorkerID string
	Stream   *StreamAccess
	LastSend time.Time
	L        sync.RWMutex
}
