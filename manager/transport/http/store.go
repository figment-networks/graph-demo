package http

/*
// InsertTransactions is http handler for InsertTransactions method
func (c *Connector) InsertTransactions(w http.ResponseWriter, req *http.Request) {
	nv := client.NetworkVersion{Network: "", Version: "0.0.1"}
	s := strings.Split(req.URL.Path, "/")

	if len(s) > 0 {
		nv.Network = s[2]
	} else {
		nv.Network = req.URL.Path
	}

	err := c.cli.InsertTransactions(req.Context(), nv, req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetTransactions is http handler for GetTransactions method
func (c *Connector) GetTransactions(w http.ResponseWriter, req *http.Request) {

	enc := json.NewEncoder(w)
	strHeight := req.URL.Query().Get("start_height")
	intHeight, err := strconv.Atoi(strHeight)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(ValidationError{Msg: "Invalid start_height param: " + err.Error()})
		return
	}

	endHeight := req.URL.Query().Get("end_height")
	intEndHeight, _ := strconv.Atoi(endHeight)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(ValidationError{Msg: "Invalid end_height param: " + err.Error()})
		return
	}

	hash := req.URL.Query().Get("hash")
	network := req.URL.Query().Get("network")
	chainID := req.URL.Query().Get("chain_id")

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Minute)
	defer cancel()

	hr := shared.HeightRange{
		StartHeight: uint64(intHeight),
		EndHeight:   uint64(intEndHeight),
		Network:     network,
		ChainID:     chainID,
	}
	if hash != "" {
		hr.Hash = hash
	}
	transactions, err := c.cli.GetTransactions(ctx, client.NetworkVersion{Network: network, Version: "0.0.1", ChainID: chainID}, hr, 1000, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ValidationError{Msg: "Error getting transaction: " + err.Error()})
		return
	}

	log.Printf("Returning %d transactions", len(transactions))
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc.Encode(transactions)
}

// ScrapeLatest is http handler for ScrapeLatest method
func (c *Connector) ScrapeLatest(w http.ResponseWriter, req *http.Request) {
	ct := req.Header.Get("Content-Type")

	ldReq := &shared.LatestDataRequest{}
	if strings.Contains(ct, "json") {
		dec := json.NewDecoder(req.Body)
		err := dec.Decode(ldReq)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
	}
	ldResp, err := c.cli.ScrapeLatest(req.Context(), *ldReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.Encode(ldResp)
}
*/
