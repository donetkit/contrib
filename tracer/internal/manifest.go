package internal

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Manifest represents a full sampling ruleset and provides
// options for configuring Logger, Clock and xrayClient.
type Manifest struct {
	name                           *string
	SamplingTargetsPollingInterval time.Duration
	xrayClient                     *xrayClient
}

// NewManifest return manifest object configured the passed with logging and an xrayClient
// configured to address addr.
func NewManifest(addr url.URL, samplingTargetsPollingInterval time.Duration, name *string) (*Manifest, error) {
	// Generate client for getSamplingRules and getSamplingTargets API call.
	client, err := newClient(addr)
	if err != nil {
		return nil, err
	}
	return &Manifest{
		xrayClient:                     client,
		SamplingTargetsPollingInterval: samplingTargetsPollingInterval,
		name:                           name,
	}, nil
}

// RefreshManifestTargets updates sampling targets (statistics) for each rule.
func (m *Manifest) RefreshManifestTargets(ctx context.Context) (target *SamplingTargetsOutput, err error) {
	target, err = m.xrayClient.getSamplingTargets(ctx, m.name)
	if err != nil {
		return nil, fmt.Errorf("refreshTargets: error occurred while getting sampling targets: %w", err)
	}
	return
}
