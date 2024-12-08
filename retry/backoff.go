package retry

import (
	"math/rand"
	"time"
)

var (
	ZeroBackoff               = &zeroBackoff{}
	StopBackoff               = &stopBackoff{}
	DefaultConstantBackoff    = NewConstantBackoff(defaultConstantInterval)
	DefaultExponentialBackoff = NewExponentialBackoff()
)

type Backoff interface {
	Next() time.Duration
}

type zeroBackoff struct{}

func (b zeroBackoff) Next() time.Duration {
	return time.Duration(0)
}

type stopBackoff struct{}

var stop time.Duration = -1

func (b stopBackoff) Next() time.Duration {
	return stop
}

type constantBackoff struct {
	interval time.Duration
}

var defaultConstantInterval = 500 * time.Millisecond

func NewConstantBackoff(d time.Duration) *constantBackoff {
	if d < 0 {
		d = defaultConstantInterval
	}
	return &constantBackoff{interval: d}
}

func (b constantBackoff) Next() time.Duration {
	return b.interval
}

type exponentialBackoff struct {
	baseInterval        time.Duration
	maxInterval         time.Duration
	randomizationFactor float64
	multiplier          float64

	currentInterval time.Duration
}

type exponentialBackoffOption func(*exponentialBackoff)

var (
	defaultExBaseInterval        = 500 * time.Millisecond
	defaultExMaxInterval         = 60 * time.Second
	defaultExRandomizationFactor = 0.5
	defaultExMultiplier          = 2.0

	WithBaseInterval = func(interval time.Duration) exponentialBackoffOption {
		return func(backoff *exponentialBackoff) {
			if interval < 0 {
				interval = defaultExBaseInterval
			}
			backoff.baseInterval = interval
		}
	}
	WithMaxInterval = func(interval time.Duration) exponentialBackoffOption {
		return func(backoff *exponentialBackoff) {
			if interval < 0 {
				interval = 0
			}
			backoff.maxInterval = interval
		}
	}
	WithRandomizationFactor = func(randomizationFactor float64) exponentialBackoffOption {
		return func(backoff *exponentialBackoff) {
			if randomizationFactor < 0 {
				randomizationFactor = 0
			}
			if randomizationFactor >= 1.0 {
				randomizationFactor = randomizationFactor - float64(int(randomizationFactor))
			}

			backoff.randomizationFactor = randomizationFactor
		}
	}
	WithMultiplier = func(multiplier float64) exponentialBackoffOption {
		return func(backoff *exponentialBackoff) {
			if multiplier <= 0 {
				multiplier = defaultExMultiplier
			}
			backoff.multiplier = multiplier
		}
	}
)

func NewExponentialBackoff(opts ...exponentialBackoffOption) *exponentialBackoff {
	v := exponentialBackoff{
		baseInterval:        defaultExBaseInterval,
		maxInterval:         defaultExMaxInterval,
		randomizationFactor: defaultExRandomizationFactor,
		multiplier:          defaultExMultiplier,
	}

	for _, opt := range opts {
		opt(&v)
	}

	v.currentInterval = v.baseInterval

	return &v
}

func (b *exponentialBackoff) Next() time.Duration {
	defer func() {
		b.incrementCurrentInterval()
	}()

	return b.getRandomizationInterval()
}

func (b *exponentialBackoff) getRandomizationInterval() time.Duration {
	delta := time.Duration(float64(b.currentInterval) * b.randomizationFactor)
	minRandomizationInterval, maxRandomizationInterval := b.currentInterval-delta, b.currentInterval+delta
	return time.Duration(float64(minRandomizationInterval) + rand.Float64()*float64(maxRandomizationInterval-minRandomizationInterval+1))
}

func (b *exponentialBackoff) incrementCurrentInterval() {
	if interval := float64(b.currentInterval) * b.multiplier; interval > float64(b.maxInterval) {
		b.currentInterval = b.maxInterval
	} else {
		b.currentInterval = time.Duration(interval)
	}
}
