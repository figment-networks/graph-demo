package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/cmd/manager/config"
	"github.com/figment-networks/graph-demo/connectivity"
	connWS "github.com/figment-networks/graph-demo/connectivity/ws"
	"github.com/figment-networks/graph-demo/manager/api"
	runnerHTTP "github.com/figment-networks/graph-demo/manager/api/runner/transport/http"
	runnerWSAPI "github.com/figment-networks/graph-demo/manager/api/runner/transport/ws"
	workerWSAPI "github.com/figment-networks/graph-demo/manager/api/worker/transport/ws"
	"github.com/figment-networks/graph-demo/manager/client"
	"github.com/figment-networks/graph-demo/manager/scheduler"
	"github.com/figment-networks/graph-demo/manager/store"
	"github.com/figment-networks/graph-demo/manager/store/postgres"
	"github.com/figment-networks/graph-demo/manager/subscription"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
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

	log := logger.GetLogger()

	dbDriver, err := postgres.NewDriver(ctx, log, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Error while creating database driver", zap.Error(err))

	}

	st := store.NewStore(dbDriver)
	mux := http.NewServeMux()
	sc := subscription.NewSubscriptions()

	reg := connWS.NewRegistry()
	client := client.NewClient(log, st, sc)

	lhs := strings.Split(cfg.LowestHeights, ",")
	lheights := make(map[string]uint64)
	for _, l := range lhs {
		ls := strings.Split(l, ":")
		if len(ls) > 1 {
			lheights[ls[0]], err = strconv.ParseUint(ls[1], 10, 64)
			if err != nil {
				log.Fatal("Error while creating database driver", zap.Error(err))
			}
		}
	}

	sched := scheduler.NewScheduler(log, client, lheights)

	serv := api.NewService(st)
	wProc := workerWSAPI.NewProcessHandler(log, serv, sched, reg)
	linkWorker(ctx, log, reg, wProc, mux)

	proc := runnerWSAPI.NewProcessHandler(log, serv, reg, sc)
	linkRunner(ctx, log, reg, proc, mux)

	ms := api.NewService(st)
	handler := runnerHTTP.NewHandler(ms)
	handler.AttachMux(mux)

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

func initConfig(path string) (config.Config, error) {
	cfg := &config.Config{}

	if path != "" {
		if err := config.FromFile(path, cfg); err != nil {
			return config.Config{}, err
		}
	}

	if err := config.FromEnv(cfg); err != nil {
		return config.Config{}, err
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
