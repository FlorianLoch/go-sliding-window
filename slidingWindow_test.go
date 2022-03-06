package slidingwindow

import (
	"math"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlidingWindow(t *testing.T) {
	t.Run("empty window", func(t *testing.T) {
		r := require.New(t)

		slider := New(1, true)

		// Cast to concrete type, otherwise we cannot inspect the internals
		core := slider.slidingWindowCore.(*synchronizedCore)

		r.Equal(0, slider.Count())
		// We expect the windowSize to be adjusted 2
		r.Equal(2, slider.Size())
		r.Equal(0, core.head)
		r.Equal(false, core.windowFull)

		r.Equal(0.0, slider.Sum())
		r.Equal(0.0, slider.Avg())
		r.Equal(0.0, slider.WeightedAvg(PositionalWeight))
	})

	t.Run("not-yet-full window", func(t *testing.T) {
		r := require.New(t)

		slider := New(3, true)

		// Cast to concrete type, otherwise we cannot inspect the internals
		core := slider.slidingWindowCore.(*synchronizedCore)

		slider.AddInt(1).AddInt(2)

		r.Equal(2, slider.Count())
		r.Equal(3, slider.Size())
		r.Equal(2, core.head)
		r.Equal(false, core.windowFull)

		r.Equal(3.0, slider.Sum())
		r.Equal(1.5, slider.Avg())
		r.Equal(5.0/3.0, slider.WeightedAvg(PositionalWeight))
	})

	t.Run("full window", func(t *testing.T) {
		r := require.New(t)

		slider := New(3, true)

		// We "over-fill" the window, we expect "1" to be dropped
		slider.Add(1).Add(2).Add(3).Add(4)

		// Cast to concrete type, otherwise we cannot inspect the internals
		core := slider.slidingWindowCore.(*synchronizedCore)

		r.Equal(3, slider.Count())
		r.Equal(3, slider.Size())
		r.Equal(1, core.head)
		r.Equal(true, core.windowFull)

		r.Equal(9.0, slider.Sum())
		r.Equal(3.0, slider.Avg())
		r.Equal(20.0/6.0, slider.WeightedAvg(PositionalWeight))

		// Adding NaN is not so nice...
		slider.Add(math.NaN())

		r.Equal(3, slider.Count())
		r.Equal(3, slider.Size())
		r.Equal(2, core.head)
		r.Equal(true, core.windowFull)

		r.True(math.IsNaN(slider.Sum()))
		r.True(math.IsNaN(slider.Avg()))
		r.True(math.IsNaN(slider.WeightedAvg(PositionalWeight)))
	})

	t.Run("synchronises multiple goroutines", func(t *testing.T) {
		// To be fair, the value of this test is quite limited
		const iterations = 1e5

		wg := sync.WaitGroup{}
		wg.Add(2)

		r := require.New(t)

		slider := New(2, true)

		// Pre-fill the window
		slider.Add(2).Add(3)

		go func() {
			// Actual result should always stay the same as values get replaces with
			// equal values
			for i := 0; i < iterations; i++ {
				slider.Add(2).Add(3)
			}

			wg.Done()
		}()

		go func() {
			for i := 0; i < iterations; i++ {
				r.Equal(5.0, slider.Sum())
				r.Equal(2.5, slider.Avg())
			}

			wg.Done()
		}()

		wg.Wait()
	})
}

func BenchmarkSlidingWindow_SynchronizedVsNonSynchronized(b *testing.B) {
	slider := New(100, false)

	b.Run("digest", func(b *testing.B) {
		for i := 0; i < 100; i++ {
			slider.Add(float64(i))
		}
	})

	b.Run("sum", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slider.Sum()
		}
	})

	slider = New(100, true)

	b.Run("synchronized:digest", func(b *testing.B) {
		for i := 0; i < 100; i++ {
			slider.Add(float64(i))
		}
	})

	b.Run("synchronized:sum", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slider.Sum()
		}
	})
}
