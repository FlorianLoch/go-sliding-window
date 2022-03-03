package slidingwindow

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
	// In case of window size 1 that would even lead to a division by zero.
	return float64(i + 1)
}

type slidingWindow struct {
	window     []float64
	size       int
	head       int
	windowFull bool
}

func New(windowSize int) SlidingWindow {
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
	sum, denominator := s.reduce(weightFn)

	return sum / denominator
}

// TODO: Add thread safe implementation via decorator pattern
