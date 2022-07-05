package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type xrayClient struct {
	// HTTP client for sending sampling requests to the collector.
	httpClient *http.Client

	// Resolved URL to call getSamplingTargets API.
	samplingTargetsURL string
}

type getSamplingTargetsInput struct {
	Name *string
}

// SamplingTargetsOutput is used to store parsed json sampling targets.
type SamplingTargetsOutput struct {
	// The percentage of matching requests to instrument, after the reservoir is
	// exhausted.
	FixedRate float64 `json:"FixedRate,omitempty"`

	// The number of seconds for the service to wait before getting sampling targets
	// again.
	Interval int64 `json:"Interval,omitempty"`

	// The number of requests per second that X-Ray allocated this service.
	ReservoirQuota float64 `json:"ReservoirQuota,omitempty"`
}

// samplingTargetDocument contains updated targeted information retrieved from X-Ray service.
type samplingTargetDocument struct {
}

// newClient returns an HTTP client with proxy endpoint.
func newClient(endpoint url.URL) (client *xrayClient, err error) {
	// Construct resolved URLs for getSamplingTargets API calls.
	endpoint.Path = "/tracer/sampling/target"
	samplingTargetsURL := endpoint

	return &xrayClient{
		httpClient:         &http.Client{},
		samplingTargetsURL: samplingTargetsURL.String(),
	}, nil
}

// getSamplingTargets calls the collector(aws proxy enabled) for sampling targets.
func (c *xrayClient) getSamplingTargets(ctx context.Context, name *string) (*SamplingTargetsOutput, error) {
	statistics := getSamplingTargetsInput{
		Name: name,
	}
	statisticsByte, err := json.Marshal(statistics)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(statisticsByte)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.samplingTargetsURL, body)
	if err != nil {
		return nil, fmt.Errorf("xray client: failed to create http request: %w", err)
	}

	output, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("xray client: unable to retrieve sampling settings: %w", err)
	}
	defer output.Body.Close()

	var samplingTargetsOutput *SamplingTargetsOutput
	if err := json.NewDecoder(output.Body).Decode(&samplingTargetsOutput); err != nil {
		return nil, fmt.Errorf("xray client: unable to unmarshal the response body: %w", err)
	}

	return samplingTargetsOutput, nil
}
