package main

import (
	"fmt"
	"github.com/dghubble/sling"
	"net/http"

	"time"
)

type DimmerAPI struct {
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

func NewDimmerAPI(baseURL string) *DimmerAPI {
	// The HTTP client does not automatically set a timeout, hence we
	// arbitrarily choose a timeout of ten seconds.
	// See: https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	return &DimmerAPI{
		client: sling.New().Client(client).Base(baseURL),
	}
}

func (a *DimmerAPI) ClearPathProbabilities() {
	if _, err := a.client.Delete("/probabilities").ReceiveSuccess(nil); err != nil {
		panic(fmt.Errorf("ClearPathProbabilities encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPI) SetPathProbabilities(rules []PathProbabilityRule) {
	if _, err := a.client.Post("/probabilities").BodyJSON(rules).ReceiveSuccess(nil); err != nil {
		panic(fmt.Errorf("SetPathProbabilities encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPI) StartTrainingMode() {
	if _, err := a.client.Post("/training").ReceiveSuccess(nil); err != nil {
		panic(fmt.Errorf("StartTrainingMode encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPI) StopTrainingMode() {
	if _, err := a.client.Delete("/training").ReceiveSuccess(nil); err != nil {
		panic(fmt.Errorf("StopTrainingMode encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPI) GetTrainingModeStats() *ResponseTimes {
	responseTimes := new(ResponseTimes)
	if _, err := a.client.Get("/training/stats").ReceiveSuccess(responseTimes); err != nil {
		panic(fmt.Errorf("GetTrainingModeStats encountered unexpected error: %w", err))
	}
	return responseTimes
}
