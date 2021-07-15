package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/figment-networks/graph-demo/manager/client"
	"github.com/figment-networks/graph-demo/manager/conn/ws"
	"github.com/figment-networks/graph-demo/manager/connectivity"
	"github.com/figment-networks/graph-demo/manager/status"
	grpcTransport "github.com/figment-networks/graph-demo/manager/transport/grpc"
	wsTransport "github.com/figment-networks/graph-demo/manager/transport/ws"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/cmd/manager/config"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type flags struct {
	configPath  string
	showVersion bool
}

var configFlags = flags{}

func init() {
	flag.BoolVar(&configFlags.showVersion, "v", false, "Show application version")
	flag.StringVar(&configFlags.configPath, "config", "", "Path to config")
	flag.Parse()
}

func main() {

	ctx := context.Background()

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

	logger.Info(config.IdentityString())

	defer logger.Sync()

	// Initialize manager
	mID, _ := uuid.NewRandom()
	connManager := connectivity.NewManager(mID.String(), logger.GetLogger())

	mux := http.NewServeMux()
	stat := status.NewStatus(connManager)
	stat.AttachToMux(mux)

	conn := ws.NewConn(logger.GetLogger())
	stat.Handler(conn)
	conn.AttachToMux(mux)

	// setup grpc transport
	grpcCli := grpcTransport.NewClient(cfg.GrpcMaxRecvSize, cfg.GrpcMaxSendSize)
	connManager.AddTransport(grpcCli)

	hClient := client.NewClient(logger.GetLogger())
	hClient.LinkSender(connManager)

	WSTransport := wsTransport.NewConnector(hClient)
	WSTransport.Handler(conn)

	//HTTPTransport := httpTransport.NewConnector(hClient)
	//HTTPTransport.AttachToHandler(mux)

	connManager.AttachToMux(mux)

	s := &http.Server{
		Addr:    cfg.Address,
		Handler: mux,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	osSig := make(chan os.Signal)
	exit := make(chan string, 2)
	signal.Notify(osSig, syscall.SIGTERM)
	signal.Notify(osSig, syscall.SIGINT)

	go runHTTP(s, cfg.Address, logger.GetLogger(), exit)

RunLoop:
	for {
		select {
		case <-osSig:
			s.Shutdown(ctx)
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
