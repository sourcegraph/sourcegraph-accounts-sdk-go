package httpretry

import (
	"net/http"
	"testing"

	"github.com/PuerkitoBio/rehttp"
	"github.com/stretchr/testify/assert"
)

func TestRetryStrategy(t *testing.T) {
	strategy := newRetryStrategy()
	for _, tc := range []struct {
		name      string
		attempt   rehttp.Attempt
		wantRetry bool
	}{
		{
			name: "more than 3 attempts",
			attempt: rehttp.Attempt{
				Index:   4,
				Request: &http.Request{},
			},
			wantRetry: false,
		},
		{
			name: "more than 3 attempts",
			attempt: rehttp.Attempt{
				Index:   4,
				Request: &http.Request{},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			shouldRetry := strategy(tc.attempt)
			assert.Equal(t, tc.wantRetry, shouldRetry)
		})
	}
}
