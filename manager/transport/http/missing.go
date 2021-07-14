package http

import (
	"net/http"
)

// MissingTransactionsResponse the response
type MissingTransactionsResponse struct {
	MissingTransactions [][2]uint64 `json:"missing_transactions"`
	MissingBlocks       [][2]uint64 `json:"missing_blocks"`
}

// CheckMissingTransactions is http handler for CheckMissingTransactions method
func (c *Connector) CheckMissingTransactions(w http.ResponseWriter, req *http.Request) {
	/*
		strHeight := req.URL.Query().Get("start_height")
		intHeight, _ := strconv.ParseUint(strHeight, 10, 64)

		endHeight := req.URL.Query().Get("end_height")
		intEndHeight, _ := strconv.ParseUint(endHeight, 10, 64)

		w.Header().Add("Content-Type", "application/json")

		if intHeight == 0 || intEndHeight == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"start_height and end_height query params are not set properly"}`))
			return
		}

		network := req.URL.Query().Get("network")
		chainID := req.URL.Query().Get("chain_id")

		if network == "" || chainID == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"network and chain_id parameters are required"}`))
			return
		}

		var err error
		mtr := MissingTransactionsResponse{}
		mtr.MissingBlocks, mtr.MissingTransactions, err = c.cli.CheckMissingTransactions(req.Context(), client.NetworkVersion{Network: network, Version: "0.0.1", ChainID: chainID}, shared.HeightRange{StartHeight: intHeight, EndHeight: intEndHeight}, client.MissingDiffTypeSQLHybrid, 999)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		enc.Encode(mtr)
	*/
}

// GetRunningTransactions gets currently running transactions
func (c *Connector) GetRunningTransactions(w http.ResponseWriter, req *http.Request) {
	/*
		run, err := c.cli.GetRunningTransactions(req.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		enc := json.NewEncoder(w)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc.Encode(run)*/

}

// StopRunningTransactions stops currently running transactions
func (c *Connector) StopRunningTransactions(w http.ResponseWriter, req *http.Request) {
	/*
		clean := (req.URL.Query().Get("clean") != "")
		network := req.URL.Query().Get("network")
		chainID := req.URL.Query().Get("chain_id")

		if network == "" || chainID == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"network and chain_id parameters are required"}`))
			return
		}

		if err := c.cli.StopRunningTransactions(req.Context(), client.NetworkVersion{Network: network, Version: "0.0.1", ChainID: chainID}, clean); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"state":"success"}`))
	*/
}

// GetMissingTransactions is http handler for GetMissingTransactions method
func (c *Connector) GetMissingTransactions(w http.ResponseWriter, req *http.Request) {
	/*
		strHeight := req.URL.Query().Get("start_height")
		intHeight, _ := strconv.ParseUint(strHeight, 10, 64)

		endHeight := req.URL.Query().Get("end_height")
		intEndHeight, _ := strconv.ParseUint(endHeight, 10, 64)

		async := req.URL.Query().Get("async")

		force := (req.URL.Query().Get("force") != "")

		overwriteAll := (req.URL.Query().Get("overwrite_all") != "")
		simplified := (req.URL.Query().Get("simplified") != "")

		network := req.URL.Query().Get("network")
		chainID := req.URL.Query().Get("chain_id")

		if network == "" || chainID == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"network and chain_id parameters are required"}`))
			return
		}

		if intHeight == 0 || intEndHeight == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("start_height and end_height query params are not set properly"))
			return
		}

		if async == "" {
			_, err := c.cli.GetMissingTransactions(req.Context(), client.NetworkVersion{Network: network, Version: "0.0.1", ChainID: chainID},
				shared.HeightRange{
					StartHeight: intHeight,
					EndHeight:   intEndHeight,
					Network:     network,
					ChainID:     chainID},
				client.GetMissingTxParams{
					Window:       1000,
					Async:        false,
					Force:        force,
					OverwriteAll: overwriteAll,
					Simplified:   simplified,
				})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		run, err := c.cli.GetMissingTransactions(req.Context(), client.NetworkVersion{Network: network, Version: "0.0.1", ChainID: chainID},
			shared.HeightRange{
				StartHeight: intHeight,
				EndHeight:   intEndHeight,
				Network:     network,
				ChainID:     chainID},
			client.GetMissingTxParams{
				Window:       1000,
				Async:        true,
				Force:        force,
				OverwriteAll: overwriteAll,
				Simplified:   simplified,
			})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		enc := json.NewEncoder(w)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if simplified {
			enc.Encode(client.Run{ // remove whole progress
				NV:               run.NV,
				HeightRange:      run.HeightRange,
				ProgressSummary:  run.ProgressSummary,
				Success:          run.Success,
				Finished:         run.Finished,
				StartedTime:      run.StartedTime,
				FinishTime:       run.FinishTime,
				LastProgressTime: run.LastProgressTime,
			})
			return
		}

		enc.Encode(run)
	*/
}
