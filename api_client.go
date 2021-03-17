package main

import (
	"fmt"
	"github.com/dghubble/sling"
	"net/http"

	"time"
)

type DimmerAPIClient struct {
	client *sling.Sling
}

type PathProbabilityRule struct {
	Path        string
	Probability float64
}

type ResponseTimes struct {
	P50 float64 `json:"P50"`
	P75 float64 `json:"P75"`
	P95 float64 `json:"P95"`
}

func NewDimmerAPIClient(baseURL string) *DimmerAPIClient {
	// The HTTP client does not automatically set a timeout, hence we
	// arbitrarily choose a timeout of ten seconds.
	// See: https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	return &DimmerAPIClient{
		client: sling.New().Client(client).Base(baseURL),
	}
}

func (a *DimmerAPIClient) ClearPathProbabilities() {
	if _, err := a.client.Delete("/probabilities").ReceiveSuccess(nil); err != nil {
		panic(fmt.Errorf("ClearPathProbabilities encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPIClient) SetPathProbabilities(rules []PathProbabilityRule) {
	if _, err := a.client.Post("/probabilities").BodyJSON(rules).ReceiveSuccess(nil); err != nil {
		panic(fmt.Errorf("SetPathProbabilities encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPIClient) StartTrainingMode() {
	if _, err := a.client.Post("/training/offline").ReceiveSuccess(nil); err != nil {
		panic(fmt.Errorf("StartTrainingMode encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPIClient) StopTrainingMode() {
	if _, err := a.client.Delete("/training/offline").ReceiveSuccess(nil); err != nil {
		panic(fmt.Errorf("StopTrainingMode encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPIClient) GetTrainingModeStats() *ResponseTimes {
	responseTimes := new(ResponseTimes)
	if _, err := a.client.Get("/training/offline/stats").ReceiveSuccess(responseTimes); err != nil {
		panic(fmt.Errorf("GetTrainingModeStats encountered unexpected error: %w", err))
	}
	return responseTimes
}
