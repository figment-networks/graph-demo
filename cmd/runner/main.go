package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/runner/jsRuntime"
	"github.com/figment-networks/graph-demo/runner/requester"
	"github.com/figment-networks/graph-demo/runner/schema"
	"github.com/figment-networks/graph-demo/runner/store"
	"github.com/figment-networks/graph-demo/runner/store/memap"
	"github.com/hasura/go-graphql-client"
)

func main() {
	rcfg := &logger.RollbarConfig{
		AppEnv: "development",
	}
	logger.Init("console", "debug", []string{"stderr"}, rcfg)

	subgraph := struct {
		name   string
		path   string
		schema string
	}{
		"simple-example",
		"./runner/subgraphs/simple-example/simple-example.js",
		"./runner/subgraphs/simple-example/schema.graphql",
	}

	// For GraphQL queries
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
	}

	loader := jsRuntime.NewLoader(rqstr, sStore)
	logger.Info(fmt.Sprintf("Loading subgraph js file %s", subgraph.path))
	if err := loader.LoadJS(subgraph.name, subgraph.path); err != nil {
		logger.Error(fmt.Errorf("Loader.LoadJS() error = %v", err))
		return
	}

	go initGraphQLSubscription(loader)
}

func initGraphQLSubscription(loader *jsRuntime.Loader) error {
	// For GraphQL subscriptions
	// https://github.com/hasura/go-graphql-client
	client := graphql.NewSubscriptionClient("wss://0.0.0.0:5002/network/cosmos")
	defer client.Close()

	type subscription struct {
		NewEvent struct {
			Time    graphql.Int
			Type    graphql.String
			Content graphql.String
		}
	}

	query := subscription{}

	_, err := client.Subscribe(&query, nil, func(dataValue *json.RawMessage, errValue error) error {
		if errValue != nil {
			// if returns error, it will failback to `onError` event
			return nil
		}
		data := subscription{}
		err := json.Unmarshal(*dataValue, &data)
		if err != nil {
			return err
		}

		fmt.Println(query.NewEvent.Time)

		if err := loader.NewBlockEvent(jsRuntime.NewBlockEvent{"network": "cosmos", "height": 1234}); err != nil {
			logger.Error(fmt.Errorf("Loader.NewBlockEvent() error = %v", err))
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
