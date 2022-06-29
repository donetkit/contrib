package internal

import "net/http"

type xrayClient struct {
	// HTTP client for sending sampling requests to the collector.
	httpClient *http.Client

	// Resolved URL to call getSamplingRules API.
	samplingRulesURL string

	// Resolved URL to call getSamplingTargets API.
	samplingTargetsURL string
}

// ruleProperties is the base set of properties that define a sampling rule.
type ruleProperties struct {
	RuleName      string            `json:"RuleName"`
	ServiceType   string            `json:"ServiceType"`
	ResourceARN   string            `json:"ResourceARN"`
	Attributes    map[string]string `json:"Attributes"`
	ServiceName   string            `json:"ServiceName"`
	Host          string            `json:"Host"`
	HTTPMethod    string            `json:"HTTPMethod"`
	URLPath       string            `json:"URLPath"`
	ReservoirSize float64           `json:"ReservoirSize"`
	FixedRate     float64           `json:"FixedRate"`
	Priority      int64             `json:"Priority"`
	Version       int64             `json:"Version"`
}
