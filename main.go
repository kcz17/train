package main

import (
	"github.com/kcz17/train/extensions"
	"github.com/kcz17/train/loadgenerator"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
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

	///////////////////////////////////////////////////////////////////////////
	// Extension: Sock Shop cart reset endpoints.
	///////////////////////////////////////////////////////////////////////////
	ExtDBReseedEnabled bool
	ExtDBReseedHost    string
	ExtDBReseedPort    string
	ExtDBReseedNumber  int
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
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

	dimmer := NewDimmerAPIClient(config.DimmerAdminHost + ":" + config.DimmerAdminPort)
	load, err := loadgenerator.NewK6Generator(config.K6Host + ":" + config.K6Port)
	if err != nil {
		log.Fatalf("NewK6Generator() failed with err != nil; err = %v", err)
	}

	var extDBReseeder *extensions.ExtDBReseeder
	if config.ExtDBReseedEnabled {
		extDBReseeder = extensions.NewExtDBReseeder(
			config.ExtDBReseedHost+":"+config.ExtDBReseedPort,
			config.ExtDBReseedNumber,
		)
	}

	// Ensure the dimmer starts in a clean state.
	if err := load.Stop(); err != nil {
		// Ignore the error if the test execution was already paused.
		if !strings.Contains(err.Error(), "cannot pause the externally controlled executor before it has started") ||
			!strings.Contains(err.Error(), "Pause error: test execution was already paused") {
			log.Fatalf("encountered error while initially stopping load generator; err = %v", err)
		}
	}
	dimmer.StopTrainingMode()
	dimmer.ClearPathProbabilities()

	for i := 1; i <= config.NumIterations; i++ {
		log.Infof("Starting iteration %d of %d\n", i, config.NumIterations)

		// Reseed the carts database
		if config.ExtDBReseedEnabled {
			if extDBReseeder == nil {
				panic("extDBReseeder should not be nil")
			}
			extDBReseeder.Reseed()
		}

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
		dimmer.StartTrainingMode()

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

		// Retrieve the response time collector stats before stopping the
		// collector, as stopping the collector will clear the stats.
		responseTimes := dimmer.GetTrainingModeStats()
		dimmer.StopTrainingMode()
		dimmer.ClearPathProbabilities()

		// Persist results.
		model.AddObservation(responseTimes.P95, probabilities)
		log.WithField("iteration", i).
			Infof("Added response time %vs with probabilities %+v", responseTimes.P95, probabilities)

		if i != config.NumIterations {
			time.Sleep(time.Duration(config.SecondsBetweenRuns) * time.Second)
		}
	}

	if err = model.Train(); err != nil {
		log.Fatalf("model.Train() failed with err != nil; err = %v", err)
	}

	log.Infof("Training complete!")
	log.Infof("Regression formula: %s", model.Formula())
	log.Infof("Normalised coefficients: %+v", model.NormalisedCoefficients())
	log.Infof("Complementary normalised coefficients: %+v", model.ComplementaryNormalisedCoefficients())

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
		NumIterations:      300,
		MaxUsers:           77,
		RampUpSeconds:      10,
		PeakSeconds:        50,
		RampDownSeconds:    10,
		SecondsBetweenRuns: 10,
		ExtDBReseedEnabled: true,
		ExtDBReseedHost:    "http://146.169.42.31",
		ExtDBReseedPort:    "30004",
		ExtDBReseedNumber:  200000,
	}
}

func initHardcodedPaths() []string {
	return []string{"recommender", "news", "cart"}
}
