package main

import (
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distmv"
	"time"
)

type ProbabilitySampler struct {
	dirchlet *distmv.Dirichlet
}

func NewProbabilitySampler(numVariables int) *ProbabilitySampler {
	// Set up the Dirichlet distribution with a number of 1.0s corresponding to
	// the number of variables according to https://stackoverflow.com/a/18662466/2102540.
	var alpha []float64
	for i := 0; i < numVariables; i++ {
		alpha = append(alpha, 1.0)
	}

	// Set the random seed to the current time for sufficient uniqueness.
	randSeed := uint64(time.Now().UTC().UnixNano())

	return &ProbabilitySampler{
		dirchlet: distmv.NewDirichlet(alpha, rand.NewSource(randSeed)),
	}
}

// Sample samples an array of probabilities adding up to 1.
func (s *ProbabilitySampler) Sample() []float64 {
	return s.dirchlet.Rand(nil)
}
