package slidingwindow

import (
	"sync"
)

type SlidingWindow interface {
	Digest(val float64) SlidingWindow
	Count() int
	Size() int
	Sum() float64
	Avg() float64
	WeightedAvg(WeightFn) float64
}

type WeightFn func(i, n int) float64

var EqualWeight WeightFn = nil

var PositionBasedWeight WeightFn = func(i, n int) float64 {
	// It is necessary to add 1 - otherwise the first value (with index 0) would basically be dropped.
	// In case of window size 1 that additionally would even lead to a division by zero.
	return float64(i + 1)
}

type slidingWindow struct {
	window     []float64
	size       int
	head       int
	windowFull bool
}

func New(windowSize int) SlidingWindow {
	// Set windowSize to a sane minimum; this also helps to get around the edge case of windowSize being <= 0
	if windowSize < 2 {
		windowSize = 2
	}

	return &slidingWindow{
		size:   windowSize,
		window: make([]float64, windowSize),
	}
}

func (s *slidingWindow) Digest(val float64) SlidingWindow {
	s.window[s.head] = val

	s.head = (s.head + 1) % s.size

	// As soon as we have to reset the head to the first field of the slice the window has been filled
	if s.head == 0 {
		s.windowFull = true
	}

	return s
}

func (s *slidingWindow) Count() int {
	if s.windowFull {
		return s.Size()
	}

	return s.head
}

func (s *slidingWindow) Size() int {
	return s.size
}

func (s *slidingWindow) reduce(weightFn WeightFn) (sum float64, denominator float64) {
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
		// a plain old average. It is backed by a benchmark.
		if weightFn != nil {
			weight = weightFn(i, count)
		}

		sum += weight * val
		denominator += weight
	}

	return
}

func (s *slidingWindow) Sum() float64 {
	sum, _ := s.reduce(EqualWeight)

	return sum
}

func (s *slidingWindow) Avg() float64 {
	return s.WeightedAvg(EqualWeight)
}

func (s *slidingWindow) WeightedAvg(weightFn WeightFn) float64 {
	if s.Count() == 0 {
		// In order to avoid division by zero below
		return 0
	}

	sum, denominator := s.reduce(weightFn)

	return sum / denominator
}

type synchronizedSlidingWindow struct {
	SlidingWindow
	lock sync.RWMutex
}

func NewSynchronized(windowSize int) SlidingWindow {
	return &synchronizedSlidingWindow{
		SlidingWindow: New(windowSize),
		lock:          sync.RWMutex{},
	}
}

func (s *synchronizedSlidingWindow) Digest(val float64) SlidingWindow {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.SlidingWindow.Digest(val)

	return s
}

func (s *synchronizedSlidingWindow) Count() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.SlidingWindow.Count()
}

// No need to wrap Size() as this one just returns a static value, no change of a race

func (s *synchronizedSlidingWindow) Sum() float64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.SlidingWindow.Sum()
}

func (s *synchronizedSlidingWindow) Avg() float64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.SlidingWindow.Avg()
}

func (s *synchronizedSlidingWindow) WeightedAvg(weightFn WeightFn) float64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.SlidingWindow.WeightedAvg(weightFn)
}
