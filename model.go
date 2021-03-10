package main

import (
	"fmt"
	"github.com/sajari/regression"
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
			m.regression.Coeff(i),
		})
	}

	return coefficients
}
