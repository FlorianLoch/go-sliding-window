# go-sliding-window

A simple sliding window implementation intended for calculating [moving averages](https://en.wikipedia.org/wiki/Moving_average).

- Synchronized, a.k.a. thread-safe, if you want it to be.
- Supports calculation of weighted averages

## Usage

```bash
go get github.com/florianloch/go-sliding-window
```

```golang
import slidingwindow "github.com/florianloch/go-sliding-window"
```

## Example

```golang
window := slidingwindow.New(3, true)

window.Add(1.0).AddInt(2)  

require.Equal(2, window.Count())
require.Equal(3, window.Size())

require.Equal(3.0, window.Sum())
require.Equal(1.5, window.Avg())

// Custom functions can be used to adjust the weights
r.Equal(5.0/3.0, window.WeightedAvg(slidingwindow.PositionalWeight))
```