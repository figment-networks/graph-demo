package structs

// TaskResponseChanPool is default pool of channels TaskResponse
var TaskResponseChanPool = NewtaskResponsePool(40)

type taskResponsePool struct {
	pool chan chan TaskResponse
}

// NewtaskResponsePool is taskResponsePool constructor
func NewtaskResponsePool(size int) *taskResponsePool {
	return &taskResponsePool{make(chan chan TaskResponse, size)}
}

// Get new channel from pool or create new
func (tr *taskResponsePool) Get() chan TaskResponse {
	select {
	case t := <-tr.pool:
		return t
	default:
	}

	return make(chan TaskResponse, 40)
}

// Put channel back to the pool or discard
func (tr *taskResponsePool) Put(t chan TaskResponse) {
	select {
	case tr.pool <- t:
	default:
		close(t)
	}
}
