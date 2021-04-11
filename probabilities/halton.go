package probabilities

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distmv"
	"gonum.org/v1/gonum/stat/samplemv"
)

type HaltonSampler struct {
	dimensions int
	iterations int
	iterator   int
	samples    *mat.Dense
}

func NewHaltonSampler(numIterations, numVariables int) *HaltonSampler {
	batch := mat.NewDense(numIterations, numVariables, nil)
	samplemv.Halton{
		Kind: samplemv.Owen,
		Q:    distmv.NewUnitUniform(numVariables, nil),
		Src:  nil,
	}.Sample(batch)

	return &HaltonSampler{
		dimensions: numVariables,
		iterations: numIterations,
		iterator:   0,
		samples:    batch,
	}
}

func (s *HaltonSampler) Sample() []float64 {
	if s.iterator >= s.iterations {
		panic(fmt.Sprintf("cannot surpass max iterations %d", s.iterations))
	}

	var probs []float64
	for j := 0; j < s.dimensions; j++ {
		probs = append(probs, s.samples.At(s.iterator, j))
	}

	s.iterator++
	return probs
}
