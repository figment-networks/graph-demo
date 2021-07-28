package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/cmd/runner/config"
	"github.com/figment-networks/graph-demo/runner/api/service"
	transportHTTP "github.com/figment-networks/graph-demo/runner/api/transport/http"
	runnerClient "github.com/figment-networks/graph-demo/runner/client"
	clientWS "github.com/figment-networks/graph-demo/runner/client/transport/ws"
	"github.com/figment-networks/graph-demo/runner/requester"
	"github.com/figment-networks/graph-demo/runner/runtime"
	"github.com/figment-networks/graph-demo/runner/schema"
	"github.com/figment-networks/graph-demo/runner/store/memap"

	"go.uber.org/zap"
)

type flags struct {
	configPath string
}

var configFlags = flags{}

func init() {
	flag.StringVar(&configFlags.configPath, "config", "", "Path to config")
	flag.Parse()
}

func main() {
	logger.Init("console", "debug", []string{"stderr"})

	// Initialize configuration
	cfg, err := initConfig(configFlags.configPath)
	if err != nil {
		log.Fatal(fmt.Errorf("error initializing config [ERR: %+v]", err))
	}

	if cfg.AppEnv == "development" || cfg.AppEnv == "local" {
		logger.Init("console", "debug", []string{"stderr"})
	} else {
		logger.Init("json", "info", []string{"stderr"})
	}
	defer logger.Sync()

	logger.Info(config.IdentityString())
	l := logger.GetLogger()

	// Using in-memory store. Create entity collections.
	sStore := memap.NewSubgraphStore()

	rqstr := requester.NewRqstr()

	// Init the javascript runtime
	loader := runtime.NewLoader(l, rqstr, sStore)

	ngc := runnerClient.NewNetworkGraphClient(l, loader)

	// Cosmos configuration
	wst := clientWS.NewNetworkGraphWSTransport(l)
	if err := wst.Connect(context.Background(), cfg.ManagerURL, ngc); err != nil {
		l.Fatal("error conectiong to websocket", zap.Error(err))
	}

	rqstr.AddDestination("cosmos", wst)

	// Load GraphQL schema for subgraph
	schemas := schema.NewSchemas(sStore, loader)
	for _, path := range strings.Split(cfg.Subgraphs, ",") {
		if err := schemas.LoadFromSubgraphYaml(path); err != nil {
			logger.Error(fmt.Errorf("Loader.LoadFromFile() error = %v", err))
			return
		}
	}

	mux := http.NewServeMux()
	svc := service.New(sStore)
	handler := transportHTTP.NewHandler(svc)
	handler.AttachMux(mux)

	s := &http.Server{
		Addr:         cfg.Address + ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 40 * time.Second,
	}

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
		}
	}

	if err := config.FromEnv(cfg); err != nil {
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
}
