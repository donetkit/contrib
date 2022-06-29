package internal

import (
	"sync"
	"time"
)

type reservoir struct {
	// Quota expiration timestamp.
	expiresAt time.Time

	// Quota assigned to client to consume per second.
	quota float64

	// Current balance of quota.
	quotaBalance float64

	// Total size of reservoir consumption per second.
	capacity float64

	// Quota refresh timestamp.
	refreshedAt time.Time

	// Polling interval for quota.
	interval time.Duration

	// Stores reservoir ticks.
	lastTick time.Time

	mu sync.RWMutex
}
