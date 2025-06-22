package httpreporter

// Semaphore is a semaphore structure
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a semaphore with a buffered channel of maxReq capacity
func NewSemaphore(maxReq int) *Semaphore {
	return &Semaphore{
		ch: make(chan struct{}, maxReq),
	}
}

// Acquire sends an empty struct to ch channel when goroutine starts
func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

// Release removes an empty struct from ch channel when goroutine finishes
func (s *Semaphore) Release() {
	<-s.ch
}
