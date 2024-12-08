package retry

import (
	"context"
	"errors"
	"time"
)

var (
	PermanentError      = errors.New("permanent error, unable to retry again")
	MaxRetryTimeError   = errors.New("the maximum number of retries has been reached")
	MaxElapsedTimeError = errors.New("the maximum retry execution time was reached")
)

const (
	DefaultMaxRetryTime   = 3
	DefaultMaxElapsedTime = 60 * time.Second
)

var defaultRetryer = &retryer{maxRetryTime: DefaultMaxRetryTime, maxElapsedTime: DefaultMaxElapsedTime, backoff: DefaultExponentialBackoff}

type GoFunc func(ctx context.Context) error

func Do(ctx context.Context, f GoFunc) error {
	return defaultRetryer.Do(ctx, f)
}

func WithMaxRetryTime(n int) *retryer {
	return &retryer{maxRetryTime: n}
}

func WithMaxElapsedTime(n time.Duration) *retryer {
	return &retryer{maxElapsedTime: n}
}

func WithBackoff(backoff Backoff) *retryer {
	return &retryer{backoff: backoff}
}

type retryer struct {
	maxRetryTime   int           // 最大重试次数
	maxElapsedTime time.Duration // 最长重试耗时
	backoff        Backoff       // 重试退避策略
}

func (r *retryer) Do(ctx context.Context, fn GoFunc) (err error) {
	var (
		retryTime = 0
		startTime = time.Now()
	)

	if r.backoff == nil {
		return errors.New("no backoff provided")
	}

	for {
		if err = fn(ctx); err == nil {
			return nil
		}

		if errors.Is(err, PermanentError) {
			return err
		}
		if r.maxRetryTime > 0 && retryTime+1 > r.maxRetryTime {
			return MaxRetryTimeError
		}
		if time.Since(startTime) > r.maxElapsedTime {
			return MaxElapsedTimeError
		}

		ch := time.NewTimer(r.backoff.Next())
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch.C:
			retryTime++
		}
	}
}

func (r *retryer) WithMaxRetryTime(n int) *retryer {
	r.maxRetryTime = n
	return r
}

func (r *retryer) WithMaxElapsedTime(n time.Duration) *retryer {
	r.maxElapsedTime = n
	return r
}

func (r *retryer) WithBackoff(backoff Backoff) *retryer {
	r.backoff = backoff
	return r
}
