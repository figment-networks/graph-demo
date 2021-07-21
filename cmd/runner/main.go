package main

import (
<<<<<<< HEAD
	"flag"
	"log"
	"net/http"
=======
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
>>>>>>> bca17a11c51b4e4f4f8d47ff80093a7fdd74ec7b
	"time"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/cmd/runner/config"
	"github.com/figment-networks/graph-demo/runner/api/service"
	transportHTTP "github.com/figment-networks/graph-demo/runner/api/transport/http"
<<<<<<< HEAD
=======
	runnerClient "github.com/figment-networks/graph-demo/runner/client"
	clientWS "github.com/figment-networks/graph-demo/runner/client/transport/ws"
	"github.com/figment-networks/graph-demo/runner/jsRuntime"
	"github.com/figment-networks/graph-demo/runner/requester"
	"github.com/figment-networks/graph-demo/runner/schema"
	"github.com/figment-networks/graph-demo/runner/store"
	"github.com/figment-networks/graph-demo/runner/store/memap"

>>>>>>> bca17a11c51b4e4f4f8d47ff80093a7fdd74ec7b
	"go.uber.org/zap"
)

type flags struct {
	configPath string
}

var configFlags = flags{}

func init() {
<<<<<<< HEAD
	flag.StringVar(&configFlags.configPath, "config", "", "path to config.json file")
=======
	flag.StringVar(&configFlags.configPath, "config", "", "Path to config")
>>>>>>> bca17a11c51b4e4f4f8d47ff80093a7fdd74ec7b
	flag.Parse()
}

func main() {
<<<<<<< HEAD
	cfg, err := getConfig(configFlags.configPath)
	if err != nil {
		log.Fatalf("error initializing config [ERR: %v]", err.Error())
=======
	logger.Init("console", "debug", []string{"stderr"})

	// TODO(l): read from config
	subgraph := struct {
		name   string
		path   string
		schema string
	}{
		"simple-example",
		"../../runner/subgraphs/simple-example/generated/mapping.js",
		"../../runner/subgraphs/simple-example/schema.graphql",
	}

	// Initialize configuration
	cfg, err := initConfig(configFlags.configPath)
	if err != nil {
		log.Fatal(fmt.Errorf("error initializing config [ERR: %+v]", err))
>>>>>>> bca17a11c51b4e4f4f8d47ff80093a7fdd74ec7b
	}

	if cfg.AppEnv == "development" || cfg.AppEnv == "local" {
		logger.Init("console", "debug", []string{"stderr"})
	} else {
		logger.Init("json", "info", []string{"stderr"})
	}
<<<<<<< HEAD

	logger.Info(config.IdentityString())
	defer logger.Sync()
=======
	defer logger.Sync()

	logger.Info(config.IdentityString())
	l := logger.GetLogger()

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

	rqstr := requester.NewRqstr()

	// Init the javascript runtime
	loader := jsRuntime.NewLoader(l, rqstr, sStore)

	logger.Info(fmt.Sprintf("Loading subgraph js file %s", subgraph.path))
	if err := loader.LoadJS(subgraph.name, subgraph.path); err != nil {
		logger.Error(fmt.Errorf("Loader.LoadJS() error = %v", err))
		return
	}

	ngc := runnerClient.NewNetworkGraphClient(l, loader)

	// Cosmos configuration
	wst := clientWS.NewNetworkGraphWSTransport(l)
	if err := wst.Connect(context.Background(), cfg.ManagerURL, ngc); err != nil {
		l.Fatal("error conectiong to websocket", zap.Error(err))
	}

	rqstr.AddDestination("cosmos", wst)
	/*
			requester.Destination{
			Name:    "cosmos",
			Kind:    "http",
			Address: "http://0.0.0.0:5001/network/cosmos", // TODO manager address
		})
	*/
>>>>>>> bca17a11c51b4e4f4f8d47ff80093a7fdd74ec7b

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

<<<<<<< HEAD
	logger.Info("[HTTP] Listening on", zap.String("address", cfg.Address), zap.String("port", cfg.HTTPPort))
	if err := s.ListenAndServe(); err != nil {
		logger.GetLogger().Error("[GRPC] Error while listening ", zap.String("address", cfg.Address), zap.String("port", cfg.HTTPPort), zap.Error(err))
	}

}

func getConfig(path string) (cfg *config.Config, err error) {
	cfg = &config.Config{}
	if path != "" {
		if err := config.FromFile(path, cfg); err != nil {
			return nil, err
=======
	/*
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
	*/

	osSig := make(chan os.Signal)
	exit := make(chan string, 2)
	signal.Notify(osSig, syscall.SIGTERM)
	signal.Notify(osSig, syscall.SIGINT)

	go runHTTP(s, cfg.Address, logger.GetLogger(), exit)

RunLoop:
	for {
		select {
		case <-osSig:
			s.Shutdown(context.Background())
			break RunLoop
		case <-exit:
			break RunLoop
		}
	}
}

func initConfig(path string) (config.Config, error) {
	cfg := &config.Config{}

	if path != "" {
		if err := config.FromFile(path, cfg); err != nil {
			return *cfg, err
>>>>>>> bca17a11c51b4e4f4f8d47ff80093a7fdd74ec7b
		}
	}

	if err := config.FromEnv(cfg); err != nil {
<<<<<<< HEAD
		return nil, err
	}

	return cfg, nil
=======
		return *cfg, err
	}

	return *cfg, nil
}

func runHTTP(s *http.Server, address string, logger *zap.Logger, exit chan<- string) {
	defer logger.Sync()

	logger.Info(fmt.Sprintf("[HTTP] Listening on %s", address))

	if err := s.ListenAndServe(); err != nil {
		logger.Error("[HTTP] failed to listen", zap.Error(err))
	}
	exit <- "http"
>>>>>>> bca17a11c51b4e4f4f8d47ff80093a7fdd74ec7b
}
