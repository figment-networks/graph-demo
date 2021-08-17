
LDFLAGS      := -w -s
MODULE       := github.com/figment-networks/graph-demo
VERSION_FILE ?= ./VERSION


ifneq (,$(wildcard $(VERSION_FILE)))
VERSION ?= $(shell head -n 1 $(VERSION_FILE))
else
VERSION ?= n/a
endif

all: build build-migration build-worker build-runner

.PHONY: build
build: LDFLAGS += -X $(MODULE)/cmd/manager/config.Timestamp=$(shell date +%s)
build: LDFLAGS += -X $(MODULE)/cmd/manager/config.Version=$(VERSION)
build:
	$(info building manager binary as ./manager_bin)
	go build -o manager_bin -ldflags '$(LDFLAGS)' ./cmd/manager


.PHONY: build-migration
build-migration: LDFLAGS += -X $(MODULE)/cmd/manager-migration/config.Timestamp=$(shell date +%s)
build-migration: LDFLAGS += -X $(MODULE)/cmd/manager-migration/config.Version=$(VERSION)
build-migration:
	$(info building manager-migration binary as ./manager_migration_bin)
	go build -o manager_migration_bin -ldflags '$(LDFLAGS)' ./cmd/manager-migration

.PHONY: build-worker
build-worker: LDFLAGS += -X $(MODULE)/cmd/cosmos-worker/config.Timestamp=$(shell date +%s)
build-worker: LDFLAGS += -X $(MODULE)/cmd/cosmos-worker/config.Version=$(VERSION)
build-worker:
	$(info building worker binary as ./worker_bin)
	go build -o worker_bin -ldflags '$(LDFLAGS)' ./cmd/cosmos-worker

.PHONY: build-runner
build-runner: LDFLAGS += -X $(MODULE)/cmd/runner/config.Timestamp=$(shell date +%s)
build-runner: LDFLAGS += -X $(MODULE)/cmd/runner/config.Version=$(VERSION)
build-runner:
	$(info building runner binary as ./runner_bin)
	go build -o runner_bin -ldflags '$(LDFLAGS)' ./cmd/runner

