package tracer

import (
	"github.com/go-logr/logr"
	"time"
)

// remoteSampler is a sampler for AWS X-Ray which polls sampling rules and sampling targets
// to make a sampling decision based on rules set by users on AWS X-Ray console.
type remoteSampler struct {
	// manifest is the list of known centralized sampling rules.
	manifest *internal.Manifest

	// pollerStarted, if true represents rule and target pollers are started.
	pollerStarted bool

	// samplingRulesPollingInterval, default is 300 seconds.
	samplingRulesPollingInterval time.Duration

	serviceName string

	cloudPlatform string

	fallbackSampler *FallbackSampler

	// logger for logging.
	logger logr.Logger
}
