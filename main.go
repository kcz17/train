package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/kcz17/train/extensions"
	"github.com/kcz17/train/loadgenerator"
	"github.com/kcz17/train/probabilities"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"
)

type Config struct {
	Endpoints struct {
		TargetHost      *string `mapstructure:"targetHost" validate:"required"`
		TargetPort      *int    `mapstructure:"targetPort" validate:"required"`
		DimmerAdminHost *string `mapstructure:"dimmerAdminHost" validate:"required"`
		DimmerAdminPort *int    `mapstructure:"dimmerAdminPort" validate:"required"`
	} `mapstructure:"endpoints" validate:"required"`

	DimmableComponentPaths []string `mapstructure:"dimmableComponentPaths" validate:"required"`

	LoadGenerator struct {
		Driver *string `mapstructure:"driver" validate:"oneof=k6"`
		K6     struct {
			Host *string `mapstructure:"host" validate:"required"`
			Port *int    `mapstructure:"port" validate:"required"`
		} `mapstructure:"k6" validate:"required_if=Driver k6"`
	} `mapstructure:"loadGenerator" validate:"required"`

	LoadProfile struct {
		NumIterations      *int `mapstructure:"numIterations" validate:"required"`
		MaxUsers           *int `mapstructure:"maxUsers" validate:"required"`
		RampUpSeconds      *int `mapstructure:"rampUpSeconds" validate:"required"`
		PeakSeconds        *int `mapstructure:"peakSeconds" validate:"required"`
		RampDownSeconds    *int `mapstructure:"rampDownSeconds" validate:"required"`
		SecondsBetweenRuns *int `mapstructure:"secondsBetweenRuns" validate:"required"`
	} `mapstructure:"loadProfile" validate:"required"`

	Extensions struct {
		SockShopCartReseeding struct {
			Enabled       *bool   `mapstructure:"enabled" validate:"required"`
			Host          *string `mapstructure:"host" validate:"required_if=Enabled true"`
			Port          *int    `mapstructure:"port" validate:"required_if=Enabled true"`
			NumReseedRows *int    `mapstructure:"numReseedRows" validate:"required_if=Enabled true"`
		} `mapstructure:"sockShopCartReseeding"`
	} `mapstructure:"extensions"`
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	conf := readConfig()
	if len(conf.DimmableComponentPaths) == 0 {
		log.Error("must have more than one path")
		return
	}

	sampler := probabilities.NewHaltonSampler(*conf.LoadProfile.NumIterations, len(conf.DimmableComponentPaths))
	model := NewPathProbabilitiesModel(conf.DimmableComponentPaths)

	dimmer := NewDimmerAPIClient(fmt.Sprintf("%s:%d", *conf.Endpoints.DimmerAdminHost, *conf.Endpoints.DimmerAdminPort))
	load, err := loadgenerator.NewK6Generator(fmt.Sprintf("%s:%d", *conf.LoadGenerator.K6.Host, *conf.LoadGenerator.K6.Port))
	if err != nil {
		log.Fatalf("NewK6Generator() failed with err != nil; err = %v", err)
	}

	var extDBReseeder *extensions.ExtDBReseeder
	if *conf.Extensions.SockShopCartReseeding.Enabled {
		extDBReseeder = extensions.NewExtDBReseeder(
			fmt.Sprintf("%s:%d", *conf.Extensions.SockShopCartReseeding.Host, conf.Extensions.SockShopCartReseeding.Port),
			*conf.Extensions.SockShopCartReseeding.NumReseedRows,
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

	for i := 1; i <= *conf.LoadProfile.NumIterations; i++ {
		log.Infof("Starting iteration %d of %d\n", i, conf.LoadProfile.NumIterations)

		// Reseed the carts database
		if *conf.Extensions.SockShopCartReseeding.Enabled {
			if extDBReseeder == nil {
				panic("extDBReseeder should not be nil")
			}
			extDBReseeder.Reseed()
		}

		// Sample a set of path probabilities associated with the paths.
		probs := sampler.Sample()

		// Set the probabilities with the server.
		var rules []PathProbabilityRule
		for i, path := range conf.DimmableComponentPaths {
			rules = append(rules, PathProbabilityRule{
				Path:        path,
				Probability: probs[i],
			})
		}
		dimmer.SetPathProbabilities(rules)
		log.WithField("iteration", i).Debugf("Using probabilities: %+v\n", rules)

		// Perform load test.
		dimmer.StartTrainingMode()

		if err := load.Start(); err != nil {
			log.WithField("iteration", i).Fatalf("encountered error while starting load generator; err = %v", err)
		}
		if err := load.Ramp(*conf.LoadProfile.MaxUsers, *conf.LoadProfile.RampUpSeconds); err != nil {
			log.WithField("iteration", i).Fatalf("encountered error while ramping up load; err = %v", err)
		}
		time.Sleep(time.Duration(*conf.LoadProfile.PeakSeconds) * time.Second)
		if err := load.Ramp(0, *conf.LoadProfile.RampDownSeconds); err != nil {
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
		model.AddObservation(responseTimes.P95, probs)
		log.WithField("iteration", i).
			Infof("Added response time %vs with probs %+v", responseTimes.P95, probs)

		if i != *conf.LoadProfile.NumIterations {
			time.Sleep(time.Duration(*conf.LoadProfile.SecondsBetweenRuns) * time.Second)
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

func readConfig() *Config {
	viper.AutomaticEnv()

	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("error: config.yaml not found.\nerr = %s", err)
		} else {
			log.Fatalf("error when reading config file at config.yaml: err = %s", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("error occured while reading configuration file: err = %s", err)
	}

	validate := validator.New()
	err := validate.Struct(&config)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Printf("unable to validate config: err = %s", err)
		}

		log.Printf("encountered validation errors:\n")

		for _, err := range err.(validator.ValidationErrors) {
			fmt.Printf("\t%s\n", err.Error())
		}

		fmt.Println("Check your configuration file and try again.")
	}

	return &config
}
