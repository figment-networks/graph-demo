package main

import (
	"fmt"
	"net/http"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/cmd/runner/config"
	"github.com/figment-networks/graph-demo/runner/api/service"
	transportHTTP "github.com/figment-networks/graph-demo/runner/api/transport/http"
	"go.uber.org/zap"
)

func main() {
	logger.Init("console", "debug", []string{"stderr"})

	subgraph := struct {
		name   string
		path   string
		schema string
	}{
		"simple-example",
		"./runner/subgraphs/simple-example/generated/mapping.js",
		"./runner/subgraphs/simple-example/schema.graphql",
	}

	// For GraphQL queries
	// TODO use a graphQL client lib?
	cli := &http.Client{}
	rqstr := requester.NewRqstr(cli)
	rqstr.AddDestination(requester.Destination{
		Name:    "cosmos",
		Kind:    "http",
		Address: "http://0.0.0.0:5001/network/cosmos", // TODO manager address
	})

	// Load GraphQL schema for subgraph
	schemas := schema.NewSchemas()
	logger.Info(fmt.Sprintf("Loading subgraph schema file %s", subgraph.schema))
	if err := schemas.LoadFromFile(subgraph.name, subgraph.schema); err != nil {
		logger.Error(fmt.Errorf("Loader.LoadFromFile() error = %v", err))
		return
	}

	// Using in-memory store. Create entity collections.
	sStore := memap.NewSubgraphStore()
	for _, sg := range schemas.Subgraphs {
		for _, ent := range sg.Entities {
			indexed := []store.NT{}
			for k, v := range ent.Params {
				indexed = append(indexed, store.NT{Name: k, Type: v.Type})
			}
			sStore.NewStore(subgraph.name, ent.Name, indexed)
		}
	logger.Info(config.IdentityString())
	defer logger.Sync()

	mux := http.NewServeMux()

	cli := http.DefaultClient
	svc := service.New(cli, cfg.ManagerURL)

	handler := transportHTTP.New(svc)
	handler.AttachMux(mux)

	s := &http.Server{
		Addr:         cfg.Address + ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 40 * time.Second,
	}

	// Init the javascript runtime
	loader := jsRuntime.NewLoader(rqstr, sStore)
	logger.Info(fmt.Sprintf("Loading subgraph js file %s", subgraph.path))
	if err := loader.LoadJS(subgraph.name, subgraph.path); err != nil {
		logger.Error(fmt.Errorf("Loader.LoadJS() error = %v", err))
		return
	}

	// TODO This is here just for testing until we get manager <> runner comms working
	evts := []jsRuntime.NewEvent{
		{
			Type: "block",
			Data: map[string]interface{}{"network": "cosmos", "height": 1234},
		},
		{
			Type: "transaction",
			Data: map[string]interface{}{"network": "cosmos", "height": 1234},
		},
	}
	for _, evt := range evts {
		if err := loader.NewEvent(evt); err != nil {
			logger.Error(fmt.Errorf("Loader.NewEvent() error = %v", err))
		}
	}
}
