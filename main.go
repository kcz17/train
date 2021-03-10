package main

import (
	"github.com/kcz17/train/loadgenerator"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type Config struct {
	///////////////////////////////////////////////////////////////////////////
	// Key endpoints.
	///////////////////////////////////////////////////////////////////////////

	TargetHost      string
	TargetPort      string
	DimmerAdminHost string
	DimmerAdminPort string

	///////////////////////////////////////////////////////////////////////////
	// Load testing tool.
	///////////////////////////////////////////////////////////////////////////

	LoadTestingDriver string
	K6Host            string
	K6Port            string

	///////////////////////////////////////////////////////////////////////////
	// Load testing profile.
	///////////////////////////////////////////////////////////////////////////

	NumIterations      int
	MaxUsers           int
	RampUpSeconds      int
	PeakSeconds        int
	RampDownSeconds    int
	SecondsBetweenRuns int
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	config := initHardcodedConfig()
	paths := initHardcodedPaths()
	if len(paths) == 0 {
		log.Error("must have more than one path")
		return
	}

	sampler := NewProbabilitySampler(len(paths))
	model := NewPathProbabilitiesModel(paths)

	dimmer := NewDimmerAPI(config.DimmerAdminHost + ":" + config.DimmerAdminPort)
	load, err := loadgenerator.NewK6Generator(config.K6Host + ":" + config.K6Port)
	if err != nil {
		log.Fatalf("NewK6Generator() failed with err != nil; err = %v", err)
	}

	// Ensure the dimmer starts in a clean state.
	if err := load.Stop(); err != nil {
		log.Fatalf("encountered error while initially stopping load generator; err = %v", err)
	}
	dimmer.StopServer()
	dimmer.StopResponseTimeCollector()
	dimmer.ClearPathProbabilities()

	for i := 1; i <= config.NumIterations; i++ {
		log.Infof("Starting iteration %d of %d\n", i, config.NumIterations)

		// Sample a set of path probabilities associated with the paths.
		probabilities := sampler.Sample()

		// Set the probabilities with the server.
		var rules []PathProbabilityRule
		for i, path := range paths {
			rules = append(rules, PathProbabilityRule{
				Path:        path,
				Probability: probabilities[i],
			})
		}
		dimmer.SetPathProbabilities(rules)
		log.WithField("iteration", i).Debugf("Using probabilities: %+v\n", rules)

		// Perform load test.
		dimmer.StartServer()
		dimmer.StartResponseTimeCollector()

		if err := load.Start(); err != nil {
			log.WithField("iteration", i).Fatalf("encountered error while starting load generator; err = %v", err)
		}
		if err := load.Ramp(config.MaxUsers, config.RampUpSeconds); err != nil {
			log.WithField("iteration", i).Fatalf("encountered error while ramping up load; err = %v", err)
		}
		time.Sleep(time.Duration(config.PeakSeconds) * time.Second)
		if err := load.Ramp(0, config.RampDownSeconds); err != nil {
			log.WithField("iteration", i).Fatalf("encountered error while ramping down load; err = %v", err)
		}
		if err := load.Stop(); err != nil {
			log.WithField("iteration", i).Fatalf("encountered error while stopping load generator; err = %v", err)
		}

		// Briefly sleep after stopping the load generation to ensure requests
		// are adequately flushed before sending API requests.
		time.Sleep(2 * time.Second)
		dimmer.StopResponseTimeCollector()
		dimmer.StopServer()
		dimmer.ClearPathProbabilities()

		// Persist results.
		responseTimes := dimmer.GetResponseTimeCollectorStats()
		model.AddObservation(responseTimes.P95, probabilities)

		time.Sleep(time.Duration(config.SecondsBetweenRuns) * time.Second)
	}

	if err = model.Train(); err != nil {
		log.Fatalf("model.Train() failed with err != nil; err = %v", err)
	}

	log.Infof("Training complete!\nCoefficients: %+v", model.Coefficients())
}

func initHardcodedConfig() *Config {
	return &Config{
		TargetHost:         "146.169.42.31",
		TargetPort:         "30002",
		DimmerAdminHost:    "http://146.169.42.31",
		DimmerAdminPort:    "30003",
		LoadTestingDriver:  "k6",
		K6Host:             "localhost",
		K6Port:             "6565",
		NumIterations:      10,
		MaxUsers:           77,
		RampUpSeconds:      20,
		PeakSeconds:        120,
		RampDownSeconds:    20,
		SecondsBetweenRuns: 30,
	}
}

func initHardcodedPaths() []string {
	return []string{"recommender", "news.html", "news", "cart"}
}
