package main

import (
	"context"

	"github.com/figment-networks/graph-demo/cmd/common/logger"

	"net/http"
)

// attachDynamic attaches handler for dynamic change of parameters
func attachDynamic(ctx context.Context, mux *http.ServeMux) {
	mux.HandleFunc("/loglevel", func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()
		l := q.Get("level")
		if l == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"level parameter has to be set"}`))
			return
		}

		if l != "panic" && l != "fatal" && l != "error" && l != "warn" && l != "info" && l != "debug" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"level parameter has to be one of:  panic,fatal,error,warn,info,debug"}`))
			return
		}

		logger.SetLevel(l)
		w.WriteHeader(http.StatusOK)
	})

}
