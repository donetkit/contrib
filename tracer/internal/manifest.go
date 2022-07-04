package internal

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"net/url"
	"strings"
	"sync"
	"time"
)

const manifestTTL = 3600
const version = 1

// Manifest represents a full sampling ruleset and provides
// options for configuring Logger, Clock and xrayClient.
type Manifest struct {
	SamplingTargetsPollingInterval time.Duration
	refreshedAt                    time.Time
	xrayClient                     *xrayClient
	logger                         glog.ILogger
	clock                          clock
	mu                             sync.RWMutex
}

// NewManifest return manifest object configured the passed with logging and an xrayClient
// configured to address addr.
func NewManifest(addr url.URL, logger glog.ILogger) (*Manifest, error) {
	// Generate client for getSamplingRules and getSamplingTargets API call.
	client, err := newClient(addr)
	if err != nil {
		return nil, err
	}
	return &Manifest{
		xrayClient:                     client,
		clock:                          &defaultClock{},
		logger:                         logger,
		SamplingTargetsPollingInterval: 30 * time.Second,
	}, nil
}

// Expired returns true if the manifest has not been successfully refreshed in
// manifestTTL seconds.
func (m *Manifest) Expired() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	manifestLiveTime := m.refreshedAt.Add(time.Second * manifestTTL)
	return m.clock.now().After(manifestLiveTime)
}

// RefreshManifestTargets updates sampling targets (statistics) for each rule.
func (m *Manifest) RefreshManifestTargets(ctx context.Context) (refresh bool, err error) {
	// Deep copy manifest object.
	manifest := m.deepCopy()

	// Generate sampling statistics based on the data in temporary manifest.
	statistics, err := manifest.snapshots()
	if err != nil {
		return false, err
	}

	// Return if no statistics to report.
	if len(statistics) == 0 {
		m.logger.WithField("Manifest", "Manifest").Debug("no statistics to report and not refreshing sampling targets")
		return false, nil
	}

	// Get sampling targets (statistics) for every expired rule from AWS X-Ray.
	targets, err := m.xrayClient.getSamplingTargets(ctx, statistics)
	if err != nil {
		return false, fmt.Errorf("refreshTargets: error occurred while getting sampling targets: %w", err)
	}

	m.logger.WithField("Manifest", "Manifest").Debug("successfully fetched sampling targets")

	// Update temporary manifest with retrieved targets (statistics) for each rule.
	refresh, err = manifest.updateTargets(targets)
	if err != nil {
		return refresh, err
	}
	return
}

// snapshots takes a snapshot of sampling statistics from all rules, resetting
// statistics counters in the process.
func (m *Manifest) snapshots() ([]*samplingStatisticsDocument, error) {
	statistics := make([]*samplingStatisticsDocument, 0, 0)
	return statistics, nil
}

func (m *Manifest) updateTargets(targets *getSamplingTargetsOutput) (refresh bool, err error) {
	// Update sampling targets for each rule.
	for _, t := range targets.SamplingTargetDocuments {
		if err := m.updateReservoir(t); err != nil {
			return false, err
		}
	}

	// Consume unprocessed statistics messages.
	for _, s := range targets.UnprocessedStatistics {
		m.logger.WithField("UpdateTargets", "UpdateTargets").Debug(
			"error occurred updating sampling target for rule, code and message", "RuleName", s.RuleName, "ErrorCode",
			s.ErrorCode,
			"Message", s.Message,
		)

		// Do not set any flags if error is unknown.
		if s.ErrorCode == nil || s.RuleName == nil {
			continue
		}

		// Set batch failure if any sampling statistics returned 5xx.
		if strings.HasPrefix(*s.ErrorCode, "5") {
			return false, fmt.Errorf("sampling statistics returned 5xx")
		}

		// Set refresh flag if any sampling statistics returned 4xx.
		if strings.HasPrefix(*s.ErrorCode, "4") {
			refresh = true
		}
	}
	return
}

func (m *Manifest) updateReservoir(t *samplingTargetDocument) (err error) {
	if t.FixedRate == nil {
		return fmt.Errorf("invalid sampling target for fixed rate %f: missing fixed rate", *t.FixedRate)
	}
	return
}

// deepCopy copies the m to another manifest object.
func (m *Manifest) deepCopy() *Manifest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	manifest := Manifest{}
	// Copying other manifest fields.
	manifest.SamplingTargetsPollingInterval = m.SamplingTargetsPollingInterval
	manifest.refreshedAt = m.refreshedAt
	manifest.xrayClient = m.xrayClient
	manifest.logger = m.logger
	manifest.clock = m.clock

	return &manifest
}
