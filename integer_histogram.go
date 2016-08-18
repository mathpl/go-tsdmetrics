package tsdmetrics

import "github.com/rcrowley/go-metrics"

type IntegerHistogram interface {
	Clear()
	Count() int64
	Max() int64
	Mean() int64
	Min() int64
	Percentile(float64) int64
	Percentiles([]float64) []int64
	Sample() metrics.Sample
	Snapshot() IntegerHistogram
	StdDev() int64
	Sum() int64
	Update(int64)
	Variance() int64
}

func NewIntegerHistogram(s metrics.Sample) IntegerHistogram {
	return &IntHistogram{sample: s}
}

// IntHistogram is the standard implementation of a Histogram and uses a
// Sample to bound its memory use.
type IntHistogram struct {
	sample metrics.Sample
}

// Clear clears the histogram and its sample.
func (h *IntHistogram) Clear() { h.sample.Clear() }

// Count returns the number of samples recorded since the histogram was last
// cleared.
func (h *IntHistogram) Count() int64 { return h.sample.Count() }

// Max returns the maximum value in the sample.
func (h *IntHistogram) Max() int64 { return h.sample.Max() }

// Mean returns the mean of the values in the sample.
func (h *IntHistogram) Mean() int64 { return int64(h.sample.Mean()) }

// Min returns the minimum value in the sample.
func (h *IntHistogram) Min() int64 { return h.sample.Min() }

// Percentile returns an arbitrary percentile of the values in the sample.
func (h *IntHistogram) Percentile(p float64) int64 {
	return int64(h.sample.Percentile(p))
}

// Percentiles returns a slice of arbitrary percentiles of the values in the
// sample.
func (h *IntHistogram) Percentiles(ps []float64) []int64 {
	vals := make([]int64, len(ps))
	for i, v := range h.sample.Percentiles(ps) {
		vals[i] = int64(v)
	}
	return vals
}

// Sample returns the Sample underlying the histogram.
func (h *IntHistogram) Sample() metrics.Sample { return h.sample }

// Snapshot returns a read-only copy of the histogram.
func (h *IntHistogram) Snapshot() IntegerHistogram {
	return &IntHistogramSnapshot{sample: h.sample.Snapshot().(*metrics.SampleSnapshot)}
}

// StdDev returns the standard deviation of the values in the sample.
func (h *IntHistogram) StdDev() int64 { return int64(h.sample.StdDev()) }

// Sum returns the sum in the sample.
func (h *IntHistogram) Sum() int64 { return int64(h.sample.Sum()) }

// Update samples a new value.
func (h *IntHistogram) Update(v int64) { h.sample.Update(v) }

// Variance returns the variance of the values in the sample.
func (h *IntHistogram) Variance() int64 { return int64(h.sample.Variance()) }

// IntHistogramSnapshot is a read-only copy of another Histogram.
type IntHistogramSnapshot struct {
	sample *metrics.SampleSnapshot
}

// Clear panics.
func (*IntHistogramSnapshot) Clear() {
	panic("Clear called on a IntHistogramSnapshot")
}

// Count returns the number of samples recorded at the time the snapshot was
// taken.
func (h *IntHistogramSnapshot) Count() int64 { return h.sample.Count() }

// Max returns the maximum value in the sample at the time the snapshot was
// taken.
func (h *IntHistogramSnapshot) Max() int64 { return h.sample.Max() }

// Mean returns the mean of the values in the sample at the time the snapshot
// was taken.
func (h *IntHistogramSnapshot) Mean() int64 { return int64(h.sample.Mean()) }

// Min returns the minimum value in the sample at the time the snapshot was
// taken.
func (h *IntHistogramSnapshot) Min() int64 { return h.sample.Min() }

// Percentile returns an arbitrary percentile of values in the sample at the
// time the snapshot was taken.
func (h *IntHistogramSnapshot) Percentile(p float64) int64 {
	return int64(h.sample.Percentile(p))
}

// Percentiles returns a slice of arbitrary percentiles of values in the sample
// at the time the snapshot was taken.
func (h *IntHistogramSnapshot) Percentiles(ps []float64) []int64 {
	vals := make([]int64, len(ps))
	for i, v := range h.sample.Percentiles(ps) {
		vals[i] = int64(v)
	}
	return vals
}

// Sample returns the Sample underlying the histogram.
func (h *IntHistogramSnapshot) Sample() metrics.Sample { return h.sample }

// Snapshot returns the snapshot.
func (h *IntHistogramSnapshot) Snapshot() IntegerHistogram { return h }

// StdDev returns the standard deviation of the values in the sample at the
// time the snapshot was taken.
func (h *IntHistogramSnapshot) StdDev() int64 { return int64(h.sample.StdDev()) }

// Sum returns the sum in the sample at the time the snapshot was taken.
func (h *IntHistogramSnapshot) Sum() int64 { return h.sample.Sum() }

// Update panics.
func (*IntHistogramSnapshot) Update(int64) {
	panic("Update called on a IntHistogramSnapshot")
}

// Variance returns the variance of inputs at the time the snapshot was taken.
func (h *IntHistogramSnapshot) Variance() int64 { return int64(h.sample.Variance()) }
