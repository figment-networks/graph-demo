package manager

import (
	"context"

	"github.com/figment-networks/graph-demo/manager/scheduler"
)

type Manager struct {
	schedulers []*scheduler.Scheduler
}

func New(schedulers []*scheduler.Scheduler) *Manager {
	return &Manager{
		schedulers: schedulers,
	}
}

func (m *Manager) RunScheduler(ctx context.Context, height uint64) {
	for _, scheduler := range m.schedulers {
		go scheduler.Start(ctx, height)
	}
}
