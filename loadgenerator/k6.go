package loadgenerator

import (
	"context"
	"fmt"
	v1 "github.com/loadimpact/k6/api/v1"
	"github.com/loadimpact/k6/api/v1/client"
	"gopkg.in/guregu/null.v3"
)

type K6Generator struct {
	client *client.Client
}

func NewK6Generator(addr string) (*K6Generator, error) {
	c, err := client.New(addr)
	if err != nil {
		return nil, fmt.Errorf("NewK6Generator(addr = %s) encountered error on client.New(addr); err = %w", addr, err)
	}

	return &K6Generator{
		client: c,
	}, err
}

func (k *K6Generator) Start() error {
	_, err := k.client.SetStatus(context.Background(), v1.Status{Paused: null.BoolFrom(false)})
	if err != nil {
		return fmt.Errorf("start() encountered error on client.SetStatus; err = %w", err)
	}
	return nil
}

func (k *K6Generator) Stop() error {
	_, err := k.client.SetStatus(context.Background(), v1.Status{Paused: null.BoolFrom(true)})
	if err != nil {
		return fmt.Errorf("stop() encountered error on client.SetStatus; err = %w", err)
	}
	return nil
}

func (k *K6Generator) SetUsers(users int) error {
	_, err := k.client.SetStatus(context.Background(), v1.Status{VUs: null.IntFrom(int64(users))})
	if err != nil {
		return fmt.Errorf("setUsers(users = %d) encountered error on client.SetStatus; err = %w", users, err)
	}
	return nil
}
