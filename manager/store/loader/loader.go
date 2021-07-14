package loader

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/figment-networks/graph-demo/manager/store/postgres"

	"github.com/figment-networks/indexing-engine/health"

	"go.uber.org/zap"
)

type healthMonitor interface {
	AddProber(ctx context.Context, p health.Prober)
}

type NC struct {
	Network string `json:"network"`
	ChainID string `json:"chain_id"`
}

type DatabaseConfig struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	NCS  []NC   `json:"for"`
}

type Database struct {
	config DatabaseConfig
	driver *postgres.Driver
}

type Loader struct {
	logger  *zap.Logger
	dbs     map[string]*Database
	dbsLock sync.RWMutex

	hm healthMonitor

	driversMap map[NC]*Database
	dmLock     sync.RWMutex
}

func NewLoader(logger *zap.Logger) (l *Loader) {
	return &Loader{
		logger:     logger,
		dbs:        make(map[string]*Database),
		driversMap: make(map[NC]*Database),
	}
}

func (l *Loader) Run() {
	//logger.Error(err)
}

func (l *Loader) LinkHealth(hM healthMonitor) {
	l.hm = hM
}

func (l *Loader) ConfigUpdate(ctx context.Context, dbConf DatabaseConfig) error {
	defer l.logger.Sync()

	l.dbsLock.Lock()
	defer l.dbsLock.Unlock()

	db, ok := l.dbs[dbConf.Name]
	if !ok {
		newDB, err := l.createDB(ctx, dbConf.Name, dbConf.URL)
		if err != nil {
			l.logger.Error("Error creating database ", zap.Error(err))
			return err
		}
		l.unlink(dbConf.NCS)

		newDB.config = dbConf
		l.dbs[dbConf.Name] = newDB

		l.relink(newDB, dbConf)
		return nil
	}

	if db.config.URL != dbConf.URL {
		newDB, err := l.createDB(ctx, dbConf.Name, dbConf.URL)
		if err != nil {
			return err
		}
		l.unlink(db.config.NCS)
		l.dbs[dbConf.Name] = newDB
		l.relink(newDB, dbConf)
		return nil
	}

	if len(db.config.NCS) != len(dbConf.NCS) {
		dbL := db
		l.relink(dbL, dbConf)
		return nil
	}

	for i, nc := range dbConf.NCS {
		if nc.ChainID != db.config.NCS[i].ChainID || nc.Network != db.config.NCS[i].Network {
			dbL := db
			l.relink(dbL, dbConf)
			return nil
		}
	}
	return nil
}

func (l *Loader) Get(nc NC) (*postgres.Driver, bool) {
	l.dmLock.RLock()
	defer l.dmLock.RUnlock()
	drv, ok := l.driversMap[nc]
	if !ok {
		return nil, false
	}
	return drv.driver, ok
}

func (l *Loader) unlink(ncs []NC) {
	l.dmLock.Lock()
	for _, nc := range ncs {
		delete(l.driversMap, nc)
	}
	l.dmLock.Unlock()
}

func (l *Loader) relink(db *Database, dbConf DatabaseConfig) {
	l.dmLock.Lock()
	for _, nc := range db.config.NCS {
		delete(l.driversMap, nc)
	}

	for _, nc := range dbConf.NCS {
		l.driversMap[nc] = db
	}
	l.dmLock.Unlock()
}

func (l *Loader) createDB(ctx context.Context, name, dbURL string) (*Database, error) {
	// connect to database
	db, err := l.connectPostgres(ctx, name, dbURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(60)
	db.SetConnMaxLifetime(time.Minute * 10)

	return &Database{driver: postgres.NewDriver(ctx, db)}, nil
}

func (l *Loader) connectPostgres(ctx context.Context, name, dbURL string) (*sql.DB, error) {
	defer l.logger.Sync()
	l.logger.Info("[DB] Connecting to database...", zap.String("name", name))
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	l.logger.Info("[DB] Ping successfull...")
	return db, nil
}
