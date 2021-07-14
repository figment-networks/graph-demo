package status

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/figment-networks/graph-demo/manager/conn"
	"github.com/figment-networks/graph-demo/manager/connectivity/structs"
)

type ConnectivityIntf interface {
	GetAllWorkers() map[string]structs.WorkerNetworkStatic
}

type Status struct {
	conn ConnectivityIntf
}

func NewStatus(conn ConnectivityIntf) *Status {
	return &Status{conn: conn}
}

func (s *Status) AttachToMux(mux *http.ServeMux) {
	mux.HandleFunc("/get_workers", func(w http.ResponseWriter, r *http.Request) {
		m, err := json.Marshal(s.conn.GetAllWorkers())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Error marshaling data"}`))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(m)
	})
}

func (s *Status) Handler(connMux conn.WsConn) {
	connMux.Attach("get_workers", s.getWorkers)
}

func (s *Status) getWorkers(ctx context.Context, req conn.Request, response conn.Response) {
	enc := json.NewEncoder(response)
	err := enc.Encode(s.conn.GetAllWorkers())
	if err != nil {
		response.Send(nil, err)
	}
}
