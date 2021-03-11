package main

import (
	"fmt"
	"github.com/sajari/regression"
	"gonum.org/v1/gonum/floats"
)

type Model struct {
	paths        []string
	observations []observation
	regression   *regression.Regression
	hasTrained   bool
}

type Coefficient struct {
	Path        string
	Coefficient float64
}

type observation struct {
	// responseTime is the observed response time in seconds.
	responseTime float64
	// pathProbabilities is the order of {@link Model.Paths} given.
	pathProbabilities []float64
}

func NewPathProbabilitiesModel(paths []string) *Model {
	r := new(regression.Regression)
	r.SetObserved("Response time, s")
	for i, path := range paths {
		r.SetVar(i, path)
	}

	return &Model{
		paths:        paths,
		observations: []observation{},
		regression:   r,
		hasTrained:   false,
	}
}

func (m *Model) AddObservation(responseTime float64, pathProbabilities []float64) {
	if m.hasTrained {
		panic("developer error: called AddObservation after training complete")
	}

	m.observations = append(m.observations, observation{
		responseTime:      responseTime,
		pathProbabilities: pathProbabilities,
	})
}

func (m *Model) Train() error {
	if m.hasTrained {
		panic("developer error: called Train after training complete")
	}

	var dataPoints regression.DataPoints
	for _, obs := range m.observations {
		dataPoints = append(dataPoints, regression.DataPoint(obs.responseTime, obs.pathProbabilities))
	}

	m.regression.Train(dataPoints...)

	err := m.regression.Run()
	if err != nil {
		return fmt.Errorf("train() failed on regression training with error: %w", m.regression.Run())
	}

	m.hasTrained = true
	return nil
}

func (m *Model) Coefficients() []Coefficient {
	if !m.hasTrained {
		panic("developer error: called Coefficients before training complete")
	}

	var coefficients []Coefficient
	for i, path := range m.paths {
		coefficients = append(coefficients, Coefficient{
			path,
			m.regression.Coeff(i + 1),
		})
	}

	return coefficients
}

// NormalisedCoefficients returns a set of coefficients which have been
// normalised between 0 and 1.
func (m *Model) NormalisedCoefficients() []Coefficient {
	if !m.hasTrained {
		panic("developer error: called NormalisedCoefficients before training complete")
	}

	normalised := m.Coefficients()

	// https://stats.stackexchange.com/a/70807
	var originalCoeffs []float64
	for _, coeff := range normalised {
		originalCoeffs = append(originalCoeffs, coeff.Coefficient)
	}
	min := floats.Min(originalCoeffs)
	max := floats.Max(originalCoeffs)
	coeffsRange := max - min

	for i := range normalised {
		normalised[i].Coefficient = (normalised[i].Coefficient - min) / coeffsRange
	}

	return normalised
}

// ComplementaryNormalisedCoefficients performs (1 - coeff) on each of the
// normalised coefficients, to represent
func (m *Model) ComplementaryNormalisedCoefficients() []Coefficient {
	normalised := m.NormalisedCoefficients()
	for i := range normalised {
		normalised[i].Coefficient = 1 - normalised[i].Coefficient
	}
	return normalised
}

func (m *Model) Formula() string {
	if !m.hasTrained {
		panic("developer error: called Formula before training complete")
	}

	return m.regression.Formula
}
