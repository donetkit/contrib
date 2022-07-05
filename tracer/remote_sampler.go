package tracer

import (
	"context"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/tracer/internal"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"time"
)

// remoteSampler is a sampler for AWS X-Ray which polls sampling rules and sampling targets
// to make a sampling decision based on rules set by users on AWS X-Ray console.
type remoteSampler struct {
	// manifest is the list of known centralized sampling rules.
	manifest *internal.Manifest

	sample bool

	// pollerStarted, if true represents rule and target pollers are started.
	pollerStarted bool

	serviceName string

	fallbackSampler *FallbackSampler

	// logger for logging.
	logger glog.ILoggerEntry
}

// Compile time assertion that remoteSampler implements the Sampler interface.
var _ sdktrace.Sampler = (*remoteSampler)(nil)

// NewRemoteSampler returns a sampler which decides to sample a given request or not
// based on the sampling rules set by users on AWS X-Ray console. Sampler also periodically polls
// sampling rules and sampling targets.
func NewRemoteSampler(ctx context.Context, serviceName string, fraction float64, opts ...RemoteSamplerOption) (sdktrace.Sampler, error) {
	// Create new config based on options or set to default values.
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}
	remoteSampler := &remoteSampler{
		fallbackSampler: NewFallbackSampler(fraction),
		serviceName:     serviceName,
	}

	if cfg.logger != nil {
		remoteSampler.logger = cfg.logger.WithField("RemoteSampler", "RemoteSampler")
	}

	// create manifest with config
	m, err := internal.NewManifest(cfg.endpoint, cfg.samplingRulesPollingInterval, &remoteSampler.serviceName)
	if err != nil {
		return nil, err
	}

	remoteSampler.manifest = m

	remoteSampler.start(ctx)

	return remoteSampler, nil
}

// ShouldSample matches span attributes with retrieved sampling rules and returns a sampling result.
// If the sampling parameters do not match or the manifest is expired then the fallback sampler is used.
func (rs *remoteSampler) ShouldSample(parameters sdktrace.SamplingParameters) sdktrace.SamplingResult {
	if rs.take(time.Now(), quotaBalance) || rs.sample {
		return sdktrace.SamplingResult{
			Tracestate: trace.SpanContextFromContext(parameters.ParentContext).TraceState(),
			Decision:   sdktrace.RecordAndSample,
		}
	}
	// traceIDRatioBasedSampler to sample 5% of additional requests every second
	return rs.fallbackSampler.ShouldSample(parameters)

}

// take consumes quota from reservoir, if any remains, then returns true. False otherwise.
func (rs *remoteSampler) take(now time.Time, itemCost float64) bool {
	rs.fallbackSampler.mu.Lock()
	defer rs.fallbackSampler.mu.Unlock()
	if rs.fallbackSampler.lastTick.IsZero() {
		return false
		//fs.lastTick = now
	}
	if rs.fallbackSampler.quotaBalance >= itemCost {
		rs.fallbackSampler.quotaBalance -= itemCost
		return true
	}

	// update quota balance based on elapsed time
	rs.refreshQuotaBalanceLocked(now)

	if rs.fallbackSampler.quotaBalance >= itemCost {
		rs.fallbackSampler.quotaBalance -= itemCost
		return true
	}

	return false
}

// refreshQuotaBalanceLocked refreshes the quotaBalance considering elapsedTime.
// It is assumed the lock is held when calling this.
func (rs *remoteSampler) refreshQuotaBalanceLocked(now time.Time) {
	elapsedTime := now.Sub(rs.fallbackSampler.lastTick)
	rs.fallbackSampler.lastTick = now

	// when elapsedTime is higher than 1 even then we need to keep quotaBalance
	// near to 1 so making elapsedTime to 1 for only borrowing 1 per second case
	if elapsedTime.Seconds() > quotaBalance {
		rs.fallbackSampler.quotaBalance += quotaBalance
	} else {
		// calculate how much credit have we accumulated since the last tick
		rs.fallbackSampler.quotaBalance += elapsedTime.Seconds()
	}
}

// Description returns description of the sampler being used.
func (rs *remoteSampler) Description() string {
	return "XRayRemoteSampler{remote sampling with X-Ray}"
}

func (rs *remoteSampler) start(ctx context.Context) {
	if !rs.pollerStarted {
		rs.pollerStarted = true
		go rs.startPoller(ctx)
	}
}

// startPoller starts the rule and target poller in a single go routine which runs periodically
// to refresh the manifest and targets.
func (rs *remoteSampler) startPoller(ctx context.Context) {
	// jitter = 100ms, default duration 10 seconds.
	targetTicker := newTicker(rs.manifest.SamplingTargetsPollingInterval, 100*time.Millisecond)
	defer targetTicker.tick.Stop()
	for {
		select {
		case _, more := <-targetTicker.c():
			if !more {
				return
			}
			target, err := rs.refreshTargets(ctx)
			if err == nil {
				if target.FixedRate > 0.0 {
					rs.sample = true
				} else {
					rs.sample = false
				}
			}
			continue
		case <-ctx.Done():
			return
		}
	}
}

// refreshTarget refreshes the sampling targets in manifest retrieved via getSamplingTargets API.
func (rs *remoteSampler) refreshTargets(ctx context.Context) (target *internal.SamplingTargetsOutput, err error) {
	target, err = rs.manifest.RefreshManifestTargets(ctx)
	if err != nil {
		if rs.logger != nil {
			rs.logger.Error(err, "error occurred while refreshing sampling targets")
		}
	}
	return
}
