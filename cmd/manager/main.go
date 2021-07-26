package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/connectivity"
	connWS "github.com/figment-networks/graph-demo/connectivity/ws"
	"github.com/gorilla/websocket"

	"github.com/figment-networks/graph-demo/cmd/manager/config"
	"github.com/figment-networks/graph-demo/manager/client"
	"github.com/figment-networks/graph-demo/manager/scheduler"
	"github.com/figment-networks/graph-demo/manager/store"
	"github.com/figment-networks/graph-demo/manager/store/postgres"
	"github.com/figment-networks/graph-demo/manager/subscription"

	"github.com/figment-networks/graph-demo/manager/api"
	runnerWSAPI "github.com/figment-networks/graph-demo/manager/api/runner/transport/ws"
	workerWSAPI "github.com/figment-networks/graph-demo/manager/api/worker/transport/ws"

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
	cfg, wCfg, err := initConfig(configFlags.configPath)
	if err != nil {
		log.Fatal(fmt.Errorf("error initializing config [ERR: %+v]", err))
	}

	log.Println(" wCfg.WorkerAddrs", wCfg.WorkerAddrs)

	if cfg.AppEnv == "development" || cfg.AppEnv == "local" {
		logger.Init("console", "debug", []string{"stderr"})
	} else {
		logger.Init("json", "info", []string{"stderr"})
	}

	logger.Info(config.IdentityString())

	defer logger.Sync()

	log := logger.GetLogger()

	dbDriver, err := postgres.NewDriver(ctx, log, cfg.DatabaseURL)
	if err != nil {
		log.Error("Error while creating database driver", zap.Error(err))
		os.Exit(1)
	}

	st := store.NewStore(dbDriver)
	mux := http.NewServeMux()
	sc := subscription.NewSubscriptions()

	reg := connWS.NewRegistry()

	client := client.NewClient(log, st, sc)
	sched := scheduler.NewScheduler(log, client)

	serv := api.NewService(st)
	wProc := workerWSAPI.NewProcessHandler(log, serv, sched, reg)
	linkWorker(ctx, log, reg, wProc, mux)

	proc := runnerWSAPI.NewProcessHandler(log, serv, reg, sc)
	linkRunner(ctx, log, reg, proc, mux)

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

	go runHTTP(s, cfg.Address, log, exit)

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

func initConfig(path string) (config.Config, config.WorkerConfig, error) {
	cfg := &config.Config{}

	if path != "" {
		if err := config.FromFile(path, cfg); err != nil {
			return config.Config{}, config.WorkerConfig{}, err
		}
	}

	if err := config.FromEnv(cfg); err != nil {
		return config.Config{}, config.WorkerConfig{}, err
	}

	workerConfig, err := getWorkerConfig(cfg.WorkerConfigPath)
	if err != nil {
		return config.Config{}, workerConfig, err
	}

	return *cfg, workerConfig, nil
}

func getWorkerConfig(path string) (workerConfig config.WorkerConfig, err error) {
	if path == "" {
		return config.WorkerConfig{}, errors.New("Missing worker config file")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config.WorkerConfig{}, err
	}

	if err = json.Unmarshal(data, &workerConfig); err != nil {
		return config.WorkerConfig{}, err
	}

	return workerConfig, err

}

func runHTTP(s *http.Server, address string, logger *zap.Logger, exit chan<- string) {
	defer logger.Sync()

	logger.Info(fmt.Sprintf("[HTTP] Listening on %s", address))

	if err := s.ListenAndServe(); err != nil {
		logger.Error("[HTTP] failed to listen", zap.Error(err))
	}
	exit <- "http"
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var ErrConnectionClosed = errors.New("connection closed")

func linkWorker(ctx context.Context, l *zap.Logger, reg *connWS.Registry, callH connectivity.FunctionCallHandler, mux *http.ServeMux) {
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		uConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			l.Warn("Error upgrading connection", zap.Error(err))
			return
		}

		sess := connWS.NewSession(ctx, uConn, l, callH)
		reg.Add(sess.ID, sess)
		go sess.Recv()
		go sess.Req()
	})
}

func linkRunner(ctx context.Context, l *zap.Logger, reg *connWS.Registry, callH connectivity.FunctionCallHandler, mux *http.ServeMux) {
	mux.HandleFunc("/runner", func(w http.ResponseWriter, r *http.Request) {
		uConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			l.Warn("Error upgrading connection", zap.Error(err))
			return
		}

		sess := connWS.NewSession(ctx, uConn, l, callH)
		reg.Add(sess.ID, sess)
		go sess.Recv()
		go sess.Req()
	})
}
