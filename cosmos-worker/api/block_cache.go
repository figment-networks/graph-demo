package api

import (
	"sync"

	"github.com/figment-networks/graph-demo/manager/structs"
)

// SimpleBlockCache simple in memory block cache to store latest blocks
type SimpleBlockCache struct {
	space  map[uint64]structs.Block
	blocks chan *structs.Block
	l      sync.RWMutex
}

// NewSimpleBlockCache a SimpleBlockCache constructor
func NewSimpleBlockCache(cap int) *SimpleBlockCache {
	return &SimpleBlockCache{
		space:  make(map[uint64]structs.Block),
		blocks: make(chan *structs.Block, cap),
	}
}

// Add blocks to the cache (thread safe)
func (sbc *SimpleBlockCache) Add(bl structs.Block) {
	sbc.l.Lock()
	defer sbc.l.Unlock()

	_, ok := sbc.space[bl.Height]
	if ok {
		return
	}

	sbc.space[bl.Height] = bl
	select {
	case sbc.blocks <- &bl:
	default:
		oldBlock := <-sbc.blocks
		if oldBlock != nil {
			delete(sbc.space, oldBlock.Height)
		}
		sbc.blocks <- &bl
	}

}

// Get and check block of given height (thread safe)
func (sbc *SimpleBlockCache) Get(height uint64) (bl structs.Block, ok bool) {
	sbc.l.RLock()
	defer sbc.l.RUnlock()

	bl, ok = sbc.space[bl.Height]
	return bl, ok
}
