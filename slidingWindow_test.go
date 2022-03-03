package slidingwindow

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlidingWindow(t *testing.T) {
	t.Run("not-yet-full window", func(t *testing.T) {
		r := require.New(t)

		// Cast to concrete type, otherwise we cannot inspect the internals
		slider := New(3).(*slidingWindow)

		slider.Digest(1).Digest(2)

		r.Equal(2, slider.Count())
		r.Equal(3, slider.Size())
		r.Equal(2, slider.head)
		r.Equal(false, slider.windowFull)

		r.Equal(3.0, slider.Sum())
		r.Equal(1.5, slider.Avg())
		r.Equal(5.0/3.0, slider.WeightedAvg(PositionBasedWeight))
	})

	t.Run("full window", func(t *testing.T) {
		r := require.New(t)

		// Cast to concrete type, otherwise we cannot inspect the internals
		slider := New(3).(*slidingWindow)

		// We "over-fill" the window, we expect "1" to be dropped
		slider.Digest(1).Digest(2).Digest(3).Digest(4)

		r.Equal(3, slider.Count())
		r.Equal(3, slider.Size())
		r.Equal(1, slider.head)
		r.Equal(true, slider.windowFull)

		r.Equal(9.0, slider.Sum())
		r.Equal(3.0, slider.Avg())
		r.Equal(20.0/6.0, slider.WeightedAvg(PositionBasedWeight))
	})
}

func BenchmarkSlidingWindow_Sum(b *testing.B) {
	slider := New(100)

	for i := 0; i < 100; i++ {
		slider.Digest(float64(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		slider.Sum()
	}
}
