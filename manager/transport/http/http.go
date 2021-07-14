package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/figment-networks/graph-demo/manager/client"
)

// Connector is main HTTP connector for manager
type Connector struct {
	cli client.ClientContractor
}

// NewConnector is  Connector constructor
func NewConnector(cli client.ClientContractor) *Connector {
	return &Connector{cli}
}

// GetRewards calculates daily rewards for provided address
func (c *Connector) GetRewards(w http.ResponseWriter, req *http.Request) {
	network := req.URL.Query().Get("network")
	chainID := req.URL.Query().Get("chain_id")
	if network == "" || chainID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"network and chain_id parameters are required"}`))
		return
	}

	account := req.URL.Query().Get("account")
	if account == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"account parameter is required"}`))
		return
	}

	start := req.URL.Query().Get("start_time")
	end := req.URL.Query().Get("end_time")
	if start == "" || end == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"start_time and end_time parameters are required"}`))
		return
	}
	startTime, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"start_time parameter should be in the format \"2006-01-02T15:04:05.999999999Z07:00\""}`))
		return
	}

	endTime, err := time.Parse(time.RFC3339Nano, end)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"end_time parameter should be in the format \"2006-01-02T15:04:05.999999999Z07:00\""}`))
		return
	}

	if endTime.Before(startTime) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"start_time must be earlier than end_time"}`))
		return
	}

	var validators []string
	validatorQuery := req.URL.Query().Get("validators")
	if len(validatorQuery) > 0 {
		for _, v := range strings.Split(validatorQuery, ",") {
			validators = append(validators, strings.TrimSpace(v))
		}
	}

	resp, err := c.cli.GetRewards(req.Context(), client.NetworkVersion{Network: network, Version: "0.0.1", ChainID: chainID}, startTime, endTime, account, validators)
	if err == client.ErrNotAvailable {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(resp)

	w.Header().Add("Content-Type", "application/json")
}

// GetAccountBalance calculates daily balance for provided address
func (c *Connector) GetAccountBalance(w http.ResponseWriter, req *http.Request) {
	network := req.URL.Query().Get("network")
	chainID := req.URL.Query().Get("chain_id")
	if network == "" || chainID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"network and chain_id parameters are required"}`))
		return
	}

	account := req.URL.Query().Get("account")
	if account == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"account parameter is required"}`))
		return
	}

	start := req.URL.Query().Get("start_time")
	end := req.URL.Query().Get("end_time")
	if start == "" || end == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"start_time and end_time parameters are required"}`))
		return
	}
	startTime, errStart := time.Parse(time.RFC3339Nano, start)
	endTime, errEnd := time.Parse(time.RFC3339Nano, end)
	if errStart != nil || errEnd != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"time range parameters should be in the format \"2006-01-02T15:04:05.999999999Z07:00\""}`))
		return
	}

	resp, err := c.cli.GetAccountBalance(req.Context(), client.NetworkVersion{Network: network, Version: "0.0.1", ChainID: chainID}, startTime, endTime, account)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(resp)

	w.Header().Add("Content-Type", "application/json")
}

// GetRewardAPR calculates daily reward annual percentage rates for provided address
func (c *Connector) GetRewardAPR(w http.ResponseWriter, req *http.Request) {
	network := req.URL.Query().Get("network")
	chainID := req.URL.Query().Get("chain_id")
	if network == "" || chainID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"network and chain_id parameters are required"}`))
		return
	}

	account := req.URL.Query().Get("account")
	if account == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"account parameter is required"}`))
		return
	}

	var validators []string
	validatorQuery := req.URL.Query().Get("validators")
	if len(validatorQuery) > 0 {
		for _, v := range strings.Split(validatorQuery, ",") {
			validators = append(validators, strings.TrimSpace(v))
		}
	}

	start := req.URL.Query().Get("start_time")
	end := req.URL.Query().Get("end_time")
	if start == "" || end == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"start_time and end_time parameters are required"}`))
		return
	}
	startTime, errStart := time.Parse(time.RFC3339Nano, start)
	endTime, errEnd := time.Parse(time.RFC3339Nano, end)
	if errStart != nil || errEnd != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"time range parameters should be in the format \"2006-01-02T15:04:05.999999999Z07:00\""}`))
		return
	}

	resp, err := c.cli.GetRewardAPR(req.Context(), client.NetworkVersion{Network: network, Version: "0.0.1", ChainID: chainID}, startTime, endTime, account, validators)
	if err == client.ErrNotAvailable {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(resp)

	w.Header().Add("Content-Type", "application/json")
}

// AttachToHandler attaches handlers to http server's mux
func (c *Connector) AttachToHandler(mux *http.ServeMux) {
	mux.HandleFunc("/rewards", c.GetRewards)
	mux.HandleFunc("/account/balance", c.GetAccountBalance)
	mux.HandleFunc("/apr", c.GetRewardAPR)
}
