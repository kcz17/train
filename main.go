package main

import (
	"github.com/kcz17/train/loadgenerator"
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

	MaxUsers           int
	RampUpSeconds      int
	PeakSeconds        int
	RampDownSeconds    int
	SecondsBetweenRuns int
}

func main() {
	config := initHardcodedConfig()

	// Get iterations from flag
	// Maybe read in previous iterations?

	// Initialise the driver
	loadGenerator := loadgenerator.NewK6Generator(config.K6Host + ":" + config.K6Port)
	// Initialise the aggregator
	// Initialise the probabilities

	// During each test
	//  Sample and set from dirchlet distribution
	//  Perform ramp-up
	//  Perform peak seconds
	//  Perform ramp down seconds
	//  Wait

	// Use https://github.com/sajari/regression
}

func initHardcodedConfig() *Config {
	return &Config{
		TargetHost:         "146.169.42.31",
		TargetPort:         "30002",
		DimmerAdminHost:    "146.169.42.31",
		DimmerAdminPort:    "30003",
		LoadTestingDriver:  "k6",
		K6Host:             "localhost",
		K6Port:             "6565",
		MaxUsers:           77,
		RampUpSeconds:      20,
		PeakSeconds:        120,
		RampDownSeconds:    20,
		SecondsBetweenRuns: 30,
	}
}
