package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/runner/jsRuntime"
	"github.com/figment-networks/graph-demo/runner/requester"
	"github.com/figment-networks/graph-demo/runner/schema"
	"github.com/figment-networks/graph-demo/runner/store"
	"github.com/figment-networks/graph-demo/runner/store/memap"
	"github.com/hasura/go-graphql-client"
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
		"./runner/subgraphs/simple-example/simple-example.js",
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
	}

	// Init the javascript runtime
	loader := jsRuntime.NewLoader(rqstr, sStore)
	logger.Info(fmt.Sprintf("Loading subgraph js file %s", subgraph.path))
	if err := loader.LoadJS(subgraph.name, subgraph.path); err != nil {
		logger.Error(fmt.Errorf("Loader.LoadJS() error = %v", err))
		return
	}

	// For GraphQL subscriptions (new events from manager)
	// TODO manager ws endpoint
	subClient := graphql.NewSubscriptionClient("wss://0.0.0.0:5002/network/cosmos").
		WithLog(log.Println).
		OnError(func(subClient *graphql.SubscriptionClient, err error) error {
			logger.Error(fmt.Errorf("graphql.NewSubscriptionClient error = %v", err))
			return err
		})
	defer subClient.Close()

	initGraphQLSubscription(subClient, loader, logger.GetLogger())

	ctx := subClient.GetContext()
	_, cancel := context.WithCancel(ctx)
	osSig := make(chan os.Signal)
	exit := make(chan string, 2)
	signal.Notify(osSig, syscall.SIGTERM)
	signal.Notify(osSig, syscall.SIGINT)

	go subClient.Run()

RunLoop:
	for {
		select {
		case sig := <-osSig:
			logger.Info("Stopping runner... ", zap.String("signal", sig.String()))
			cancel()
			subClient.Close()
			break RunLoop
		case k := <-exit:
			logger.Info("Stopping runner... ", zap.String("reason", k))
			cancel()
			subClient.Close()
			break RunLoop
		}
	}
}

type subscription struct {
	NewEvent struct {
		Time    graphql.Int
		Type    graphql.String
		Content graphql.String
	}
}

// Example at https://github.com/hasura/go-graphql-client/blob/master/example/subscription/main.go
func initGraphQLSubscription(client *graphql.SubscriptionClient, loader *jsRuntime.Loader, logger *zap.Logger) error {
	query := subscription{}

	logger.Info("Establishing graphQL subscription")
	_, err := client.Subscribe(&query, nil, func(dataValue *json.RawMessage, errValue error) error {
		if errValue != nil {
			return errValue // falls back to `onError` event
		}
		data := subscription{}
		err := json.Unmarshal(*dataValue, &data)
		if err != nil {
			logger.Error("could not parse graphQL response error = %v", zap.Error(err))
			return err
		}

		// TODO based on event type, call different event handlers

		if err := loader.NewBlockEvent(jsRuntime.NewBlockEvent{"network": "cosmos", "height": data.NewEvent.Time}); err != nil {
			logger.Error("Loader.NewBlockEvent() error = %v", zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("subscription client handler error = %v", zap.Error(err))
		return err
	}

	return nil
}
