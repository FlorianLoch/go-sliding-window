package slidingwindow

import "sync"

type slidingWindowCore interface {
	Add(val float64)
	reduce(fn WeightFn) (sum float64, denominator float64)
	Size() int
	Count() int
}

type core struct {
	window     []float64
	size       int
	head       int
	windowFull bool
}

func newCore(windowSize int) *core {
	// Set windowSize to a sane minimum; this also helps to get around the edge case of windowSize being <= 0
	if windowSize < 2 {
		windowSize = 2
	}

	return &core{
		size:   windowSize,
		window: make([]float64, windowSize),
	}
}

func (s *core) Add(val float64) {
	s.window[s.head] = val

	s.head = (s.head + 1) % s.size

	// The window has been filled as soon as we have to reset the head to the first index of the slice
	if s.head == 0 {
		s.windowFull = true
	}
}

func (s *core) Count() int {
	if s.windowFull {
		return s.Size()
	}

	return s.head
}

func (s *core) Size() int {
	return s.size
}

func (s *core) reduce(weightFn WeightFn) (sum, denominator float64) {
	count := s.Count()

	offset := s.head
	steps := s.size

	if !s.windowFull {
		offset = 0
		steps = s.head
	}

	weight := 1.0

	for i := 0; i < steps; i++ {
		val := s.window[(offset+i)%s.size]

		// This is an optimization in order to avoid all the invocations in case of summing up or computing
		// a plain old average (in which case weight is 1). This is backed by a benchmark.
		if weightFn != nil {
			weight = weightFn(i, count)
		}

		sum += weight * val
		denominator += weight
	}

	return
}

type synchronizedCore struct {
	*core
	lock sync.RWMutex
}

func newSynchronizedCore(windowSize int) *synchronizedCore {
	return &synchronizedCore{
		core: newCore(windowSize),
	}
}

func (s *synchronizedCore) Add(val float64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.core.Add(val)
}

func (s *synchronizedCore) reduce(fn WeightFn) (sum float64, denominator float64) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.core.reduce(fn)
}

// No need to wrap Size() as this one just returns a static value, no chance of a data race

func (s *synchronizedCore) Count() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.core.Count()
}
