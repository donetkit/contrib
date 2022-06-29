package internal

import (
	"github.com/go-logr/logr"
	"sync"
	"time"
)

const manifestTTL = 3600
const version = 1

// Manifest represents a full sampling ruleset and provides
// options for configuring Logger, Clock and xrayClient.
type Manifest struct {
	Rules                          []Rule
	SamplingTargetsPollingInterval time.Duration
	refreshedAt                    time.Time
	xrayClient                     *xrayClient
	clientID                       *string
	logger                         logr.Logger
	clock                          clock
	mu                             sync.RWMutex
}
