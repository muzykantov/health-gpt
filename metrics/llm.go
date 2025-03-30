package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestDuration measures LLM API request execution time
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_request_duration_seconds",
			Help:    "Duration of LLM API requests in seconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // Starts from 0.1s with 10 buckets
		},
		[]string{"provider", "model", "status"},
	)

	// RequestTotal counts total number of requests
	RequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_requests_total",
			Help: "Total number of LLM API requests",
		},
		[]string{"provider", "model", "status"},
	)

	// TokensProcessed measures tokens usage
	TokensProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_tokens_processed_total",
			Help: "Total number of tokens processed",
		},
		[]string{"provider", "model", "type"}, // type: prompt, completion
	)

	// ValidationScores measures validation reliability scores
	ValidationScores = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_validation_score",
			Help:    "Distribution of validation reliability scores",
			Buckets: prometheus.LinearBuckets(0, 0.1, 11), // 0.0 to 1.0 in 0.1 increments
		},
		[]string{"provider", "model"},
	)

	// ValidationRetries counts validation retries
	ValidationRetries = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_validation_retries",
			Help:    "Number of retries needed for validation",
			Buckets: []float64{0, 1, 2, 3, 4, 5},
		},
		[]string{"provider", "model"},
	)

	// ValidationDuration measures total validation process time including retries
	ValidationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_validation_duration_seconds",
			Help:    "Duration of entire validation process in seconds",
			Buckets: prometheus.ExponentialBuckets(0.5, 2, 8), // From 0.5s to ~64s
		},
		[]string{"provider", "model", "status"},
	)
)

// ObserveRequestDuration records the duration of a request with its result
func ObserveRequestDuration(provider, model, status string, duration time.Duration) {
	RequestDuration.WithLabelValues(provider, model, status).Observe(duration.Seconds())
	RequestTotal.WithLabelValues(provider, model, status).Inc()
}

// AddTokens adds token counts to metrics
func AddTokens(provider, model string, promptTokens, completionTokens int) {
	if promptTokens > 0 {
		TokensProcessed.WithLabelValues(provider, model, "prompt").Add(float64(promptTokens))
	}
	if completionTokens > 0 {
		TokensProcessed.WithLabelValues(provider, model, "completion").Add(float64(completionTokens))
	}
}

// ObserveValidationScore records a validation score
func ObserveValidationScore(provider, model string, score float64) {
	ValidationScores.WithLabelValues(provider, model).Observe(score)
}

// ObserveValidationRetries records number of retries for validation
func ObserveValidationRetries(provider, model string, retries int) {
	ValidationRetries.WithLabelValues(provider, model).Observe(float64(retries))
}

// ObserveValidationDuration records the duration of the entire validation process
func ObserveValidationDuration(provider, model, status string, duration time.Duration) {
	ValidationDuration.WithLabelValues(provider, model, status).Observe(duration.Seconds())
}
