package stats

import (
	"fmt"
	"reflect"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/core/util/reflectx"
)

// Mean returns the arithmetic mean (or average)-- per key, if keyed -- of the
// elements in a collection. It expects a PCollection<A> or PCollection<KV<A,B>>
// as input and returns a singleton PCollection<float64> or a
// PCollection<KV<A,float64>>, respectively. It can only be used for numbers,
// such as int, uint16, float32, etc.
//
// For example:
//
//    col := beam.Create(p, 1, 11, 7, 5, 10)
//    mean := stats.Mean(p, col)   // PCollection<float64> with 6.8 as the only element.
//
func Mean(p *beam.Pipeline, col beam.PCollection) beam.PCollection {
	p = p.Composite("stats.Mean")

	t := beam.FindCombineType(col)
	if !reflectx.IsNumber(t) || reflectx.IsComplex(t) {
		panic(fmt.Sprintf("Mean requires a non-complex number: %v", t))
	}

	return beam.Combine(p, &meanFn{}, col)
}

// TODO(herohde) 7/7/2017: the accumulator should be serializable with a Coder.

type meanAccum struct {
	Count int64
	Sum   float64
}

// meanFn is a combineFn that accumulates the count and sum of numbers to
// produce their mean. It assumes numbers are convertible to float64.
type meanFn struct{}

func (f *meanFn) CreateAccumulator() meanAccum {
	return meanAccum{}
}

func (f *meanFn) AddInput(a meanAccum, val beam.T) meanAccum {
	a.Count++
	a.Sum += reflect.ValueOf(val.(interface{})).Convert(reflectx.Float64).Interface().(float64)
	return a
}

func (f *meanFn) MergeAccumulators(list []meanAccum) meanAccum {
	var ret meanAccum
	for _, a := range list {
		ret.Count += a.Count
		ret.Sum += a.Sum
	}
	return ret
}

func (f *meanFn) ExtractOutput(a meanAccum) float64 {
	if a.Count == 0 {
		return 0
	}
	return a.Sum / float64(a.Count)
}
