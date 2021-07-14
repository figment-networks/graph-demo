package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/figment-networks/graph-demo/cmd/manager-migration/config"
	"github.com/figment-networks/graph-demo/manager/store/loader"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type flags struct {
	configPath    string
	migrationPath string
	version       uint
	verbose       bool
}

var configFlags = flags{}

func init() {
	flag.StringVar(&configFlags.configPath, "config", "", "Path to config")
	flag.BoolVar(&configFlags.verbose, "verbose", true, "Verbosity of logs during run")
	flag.UintVar(&configFlags.version, "version", 0, "Version parameter sets db changes to specified revision (up or down)")
	flag.StringVar(&configFlags.migrationPath, "path", "./migrations", "Path to migration folder")
	flag.Parse()
}

func main() {
	log.SetOutput(os.Stdout)
	// Initialize configuration
	cfg, err := initConfig(configFlags.configPath)
	if err != nil {
		log.Fatal(fmt.Errorf("error initializing config [ERR: %+v]", err))
	}

	if cfg.EnvDBConfig == "" && cfg.DatabaseURL == "" {
		log.Fatal(err)
	}

	if cfg.EnvDBConfig != "" {
		rc := &RunConfig{}
		dec := json.NewDecoder(strings.NewReader(cfg.EnvDBConfig))
		if err = dec.Decode(rc); err != nil {
			log.Fatal(err)
		}

		for _, conn := range rc.Connections {
			if configFlags.verbose {
				log.Println("Using migrations from: ", configFlags.migrationPath, " ", conn.Name)
			}

			err := migrateDB(conn.URL)
			if err != nil {
				log.Fatal(err)
			}
		}
		return
	} else if cfg.DatabaseURL != "" {

		if configFlags.verbose {
			log.Println("Using migrations from: ", configFlags.migrationPath, " for db url")
		}
		err := migrateDB(cfg.DatabaseURL)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func migrateDB(dburl string) error {

	db, err := sql.Open("postgres", dburl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	srcPath := fmt.Sprintf("file://%s", configFlags.migrationPath)

	if configFlags.verbose {
		log.Println("Using migrations from: ", configFlags.migrationPath)
	}

	if configFlags.version > 0 {
		if configFlags.verbose {
			log.Println("Migrating to version: ", configFlags.version)
		}
		err = migrateTo(srcPath, dburl, configFlags.version)
	} else {
		err = runMigrations(srcPath, dburl)
	}

	if err != nil {
		if err != migrate.ErrNoChange {
			return err
		}
		if configFlags.verbose {
			log.Println("No change")
		}
	}
	return nil
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

func runMigrations(srcPath, dbURL string) error {
	m, err := migrate.New(srcPath, dbURL)
	if err != nil {
		return err
	}

	defer m.Close()

	return m.Up()
}

func migrateTo(srcPath, dbURL string, version uint) error {
	m, err := migrate.New(srcPath, dbURL)
	if err != nil {
		return err
	}

	defer m.Close()

	return m.Migrate(version)
}

type RunConfig struct {
	Connections []loader.DatabaseConfig `json:"connections"`
}
