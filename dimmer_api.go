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

func (a *DimmerAPI) StartServer() {
	if _, err := a.client.Post("/start").Request(); err != nil {
		panic(fmt.Errorf("StartServer encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPI) StopServer() {
	if _, err := a.client.Post("/stop").Request(); err != nil {
		panic(fmt.Errorf("StopServer encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPI) ClearPathProbabilities() {
	if _, err := a.client.Delete("/probabilities").Request(); err != nil {
		panic(fmt.Errorf("ClearPathProbabilities encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPI) SetPathProbabilities(rules []PathProbabilityRule) {
	if _, err := a.client.Post("/probabilities").BodyJSON(rules).Request(); err != nil {
		panic(fmt.Errorf("SetPathProbabilities encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPI) StartResponseTimeCollector() {
	if _, err := a.client.Post("/collector").Request(); err != nil {
		panic(fmt.Errorf("StartResponseTimeCollector encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPI) StopResponseTimeCollector() {
	if _, err := a.client.Delete("/collector").Request(); err != nil {
		panic(fmt.Errorf("StopResponseTimeCollector encountered unexpected error: %w", err))
	}
}

func (a *DimmerAPI) GetResponseTimeCollectorStats() *ResponseTimes {
	responseTimes := new(ResponseTimes)
	if _, err := a.client.Get("/collector").ReceiveSuccess(responseTimes); err != nil {
		panic(fmt.Errorf("GetResponseTimeCollectorStats encountered unexpected error: %w", err))
	}
	return responseTimes
}
