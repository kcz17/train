package probabilities

type Sampler interface {
	Sample() []float64
}
