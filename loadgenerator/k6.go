package loadgenerator

import (
	"context"
	"fmt"
	v1 "github.com/loadimpact/k6/api/v1"
	"github.com/loadimpact/k6/api/v1/client"
	log "github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v3"
	"math"
	"time"
)

type K6Generator struct {
	client       *client.Client
	currentUsers int
}

func NewK6Generator(addr string) (*K6Generator, error) {
	c, err := client.New(addr)
	if err != nil {
		return nil, fmt.Errorf("NewK6Generator(addr = %s) encountered error on client.New(addr); err = %w", addr, err)
	}

	return &K6Generator{
		client:       c,
		currentUsers: 0,
	}, err
}

func (k *K6Generator) Start() error {
	_, err := k.client.SetStatus(context.Background(), v1.Status{Paused: null.BoolFrom(false)})
	if err != nil {
		return fmt.Errorf("start() encountered error on client.SetStatus; err = %w", err)
	}
	log.Debugf("Start() complete\n")
	return nil
}

func (k *K6Generator) Stop() error {
	_, err := k.client.SetStatus(context.Background(), v1.Status{Paused: null.BoolFrom(true)})
	if err != nil {
		return fmt.Errorf("stop() encountered error on client.SetStatus; err = %w", err)
	}
	log.Debugf("Stop() complete\n")
	return nil
}

func (k *K6Generator) SetUsers(users int) error {
	_, err := k.client.SetStatus(context.Background(), v1.Status{VUs: null.IntFrom(int64(users))})
	if err != nil {
		return fmt.Errorf("setUsers(users = %d) encountered error on client.SetStatus; err = %w", users, err)
	}
	log.Debugf("SetUsers(%d) complete\n", users)
	return nil
}

func (k *K6Generator) Ramp(target int, durationSeconds int) error {
	// Figure out how many users to ramp per second.
	usersPerSecond := (target - k.currentUsers) / durationSeconds

	// Ramp up every second. Users cannot be fractional, so we keep an internal
	// counter of fractional users being ramped up by usersPerSecond every
	// second, but we send a rounded number to the k6 client.
	remaining := target - k.currentUsers
	iterationTarget := k.currentUsers
	for range time.Tick(1 * time.Second) {
		iterationTarget += usersPerSecond

		// Send the rounded number to the k6 client.
		roundedIterationTarget := int(math.Round(float64(iterationTarget)))
		if err := k.SetUsers(roundedIterationTarget); err != nil {
			return fmt.Errorf("ramp() encountered an error on client.SetStatus(target = %d); err = %w", roundedIterationTarget, err)
		}
		k.currentUsers = roundedIterationTarget

		remaining -= usersPerSecond
		if remaining <= 0 {
			break
		}
	}

	// Since actions under the time.Tick may deviate the next tick, we resolve
	// the deviance by making a final user set.
	if err := k.SetUsers(target); err != nil {
		return fmt.Errorf("final ramp() encountered an error on client.SetStatus(target = %d); err = %w", target, err)
	}
	k.currentUsers = target

	log.Debugf("Ramp(target = %d, duration = %d) complete", target, durationSeconds)
	return nil
}
