package main

import (
	"golang.org/x/exp/rand"
	"time"
)

type ProbabilitySampler struct {
	n      int
	random *rand.Rand
}

func NewProbabilitySampler(numVariables int) *ProbabilitySampler {
	// Set the random seed to the current time for sufficient uniqueness.
	randSeed := uint64(time.Now().UTC().UnixNano())

	return &ProbabilitySampler{
		n:      numVariables,
		random: rand.New(rand.NewSource(randSeed)),
	}
}

// Sample samples an array of probabilities.
func (s *ProbabilitySampler) Sample() []float64 {
	var probs []float64

	for i := 0; i < s.n; i++ {
		probs = append(probs, s.random.Float64())
	}

	return probs
}
