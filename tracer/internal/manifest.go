package internal

import (
	"github.com/donetkit/contrib-log/glog"
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
