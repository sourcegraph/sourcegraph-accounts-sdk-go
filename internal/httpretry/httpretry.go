package httpretry

import (
	"errors"
	"io"
	"math/rand"
	"net/http"
	"syscall"
	"time"

	"github.com/PuerkitoBio/rehttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.GetMeterProvider().Meter("sourcegraph-accounts-sdk-go/httpretry")

var roundTripperRetryCounter = func() metric.Int64Counter {
	c, err := meter.Int64Counter("httpretry.roundtripper_retry.count")
	if err != nil {
		panic(err)
	}
	return c
}()

// NewRoundTripper wraps base in a retry strategy tailored for retrying common
// transient failures.
//
// This is intentionally standalone from the Sourcegraph 'internal/httpcli'
// package for a more minimal feature set.
func NewRoundTripper(base http.RoundTripper) http.RoundTripper {
	strategy := newRetryStrategy()
	randGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))

	return rehttp.NewTransport(
		base,
		func(attempt rehttp.Attempt) bool {
			shouldRetry := strategy(attempt)

			// Record metrics about retries before returning.
			if shouldRetry {
				attrs := []attribute.KeyValue{
					attribute.String("dst_host", attempt.Request.Host),
				}
				if attempt.Request.URL != nil {
					attrs = append(attrs,
						attribute.String("dst_scheme", attempt.Request.URL.Scheme))
				}
				roundTripperRetryCounter.Add(attempt.Request.Context(), 1,
					metric.WithAttributeSet(attribute.NewSet(attrs...)))
			}

			return shouldRetry
		},
		// Aggressive delays to make sure that we don't retry for too long,
		// using our own random generator to avoid package-level var dependency.
		rehttp.ExpJitterDelayWithRand(
			10*time.Millisecond,
			100*time.Millisecond,
			randGenerator.Int63n))
}

func newRetryStrategy() rehttp.RetryFn {
	return rehttp.RetryAll(
		// Retry up to 3 times only
		rehttp.RetryMaxRetries(3),
		// Retry any of our likely-to-be-transient scenarios
		rehttp.RetryAny(
			rehttp.RetryIsErr(func(err error) bool {
				// Specific errors to retry on.
				switch {
				case errors.Is(err, syscall.ECONNRESET):
					return true
				case errors.Is(err, syscall.ECONNABORTED):
					return true
				case errors.Is(err, io.ErrUnexpectedEOF):
					return true
				}
				// For all else, return false, unless another policy matches.
				return false
			}),
			rehttp.RetryStatuses(
				// When service is unreachable, try again. In particular,
				// 502/503 are often returned by various networking layers
				// when they cannot connect to the underlying service, which
				// can flake for reasons outside our control.
				http.StatusBadGateway,
				http.StatusServiceUnavailable,
			),
		),
	)
}
