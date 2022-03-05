package slidingwindow

type WeightFn func(i, n int) float64

var EqualWeight WeightFn = nil

var PositionalWeight WeightFn = func(i, n int) float64 {
	// It is necessary to add 1 - otherwise the first value (with index 0) would basically be dropped.
	// In case of windowSize == 1 that additionally would lead to a division by zero.
	return float64(i + 1)
}

type SlidingWindow struct {
	slidingWindowCore
}

func New(windowSize int, synchronized bool) *SlidingWindow {
	if synchronized {
		return &SlidingWindow{
			slidingWindowCore: newSynchronizedCore(windowSize),
		}
	}

	return &SlidingWindow{
		slidingWindowCore: newCore(windowSize),
	}
}

func (s *SlidingWindow) Add(val float64) *SlidingWindow {
	s.slidingWindowCore.Add(val)

	return s
}

func (s *SlidingWindow) Sum() float64 {
	sum, _ := s.slidingWindowCore.reduce(EqualWeight)

	return sum
}

func (s *SlidingWindow) Avg() float64 {
	return s.WeightedAvg(EqualWeight)
}

func (s *SlidingWindow) WeightedAvg(weightFn WeightFn) float64 {
	if s.slidingWindowCore.Count() == 0 {
		// In order to avoid division by zero below
		return 0
	}

	sum, denominator := s.slidingWindowCore.reduce(weightFn)

	return sum / denominator
}
