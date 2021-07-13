package main

import (
	"fmt"
	"net/http"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/runner/jsRuntime"
	"github.com/figment-networks/graph-demo/runner/requester"
	"github.com/figment-networks/graph-demo/runner/schema"
	"github.com/figment-networks/graph-demo/runner/store"
	"github.com/figment-networks/graph-demo/runner/store/memap"
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

	cli := &http.Client{}
	rqstr := requester.NewRqstr(cli)
	rqstr.AddDestination(requester.Destination{
		Name:    "cosmos",
		Kind:    "http",
		Address: "http://0.0.0.0:5001", // TODO manager address
	})

	schemas := schema.NewSchemas()
	logger.Info(fmt.Sprintf("Loading subgraph schema file %s", subgraph.schema))
	if err := schemas.LoadFromFile(subgraph.name, subgraph.schema); err != nil {
		logger.Error(fmt.Errorf("Loader.LoadFromFile() error = %v", err))
		return
	}

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

	// TODO should be called via "connectivity" between manager -> runner
	if err := loader.NewBlockEvent(jsRuntime.NewBlockEvent{"network": "cosmos", "height": 1234}); err != nil {
		logger.Error(fmt.Errorf("Loader.NewBlockEvent() error = %v", err))
	}

	/* 	ctx, cancel := context.WithCancel(context.Background())
		mux := http.NewServeMux()
		httpapi.AttachMux(*mux)

		server := &http.Server{
			Addr:         ":5000",
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		osSig := make(chan os.Signal)
		exit := make(chan string, 2)
		signal.Notify(osSig, syscall.SIGTERM)
		signal.Notify(osSig, syscall.SIGINT)

		go runHTTP(server, "5000", logger.GetLogger(), exit)

	RunLoop:
		for {
			select {
			case sig := <-osSig:
				logger.Info("Stopping runner... ", zap.String("signal", sig.String()))
				cancel()
				logger.Info("Canceled context, gracefully stopping http")
				err := server.Shutdown(ctx)
				if err != nil {
					logger.GetLogger().Error("Error stopping http server ", zap.Error(err))
				}
				break RunLoop
			case k := <-exit:
				logger.Info("Stopping runner... ", zap.String("reason", k))
				cancel()
				logger.Info("Canceled context, gracefully stopping http")
				err := server.Shutdown(ctx)
				if err != nil {
					logger.GetLogger().Error("Error stopping http server ", zap.Error(err))
				}
				break RunLoop
			}
		} */
}

/* func runHTTP(s *http.Server, port string, logger *zap.Logger, exit chan<- string) {
	defer logger.Sync()

	logger.Info(fmt.Sprintf("[HTTP] Listening on 0.0.0.0:%s", port))
	if err := s.ListenAndServe(); err != nil {
		logger.Error("[HTTP] failed to listen", zap.Error(err))
	}
	exit <- "http"
} */
