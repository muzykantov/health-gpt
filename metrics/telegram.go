package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TelegramMessagesTotal counts incoming messages
	TelegramMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_messages_total",
			Help: "Total number of messages received from Telegram",
		},
		[]string{"type"}, // text, command, callback
	)

	// TelegramMessagesError counts errors during message processing
	TelegramMessagesError = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_messages_error_total",
			Help: "Total number of errors while processing Telegram messages",
		},
		[]string{"error_type"}, // send_error, completion_error, etc.
	)

	// TelegramResponseTime measures response time
	TelegramResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "telegram_response_time_seconds",
			Help:    "Response time for Telegram messages",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // From 0.1s to ~100s
		},
		[]string{"message_type"}, // text, command, callback
	)

	// TelegramActiveUsers counts unique active users
	TelegramActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "telegram_active_users",
			Help: "Number of unique users interacting with the bot",
		},
	)

	// TelegramUserSessions counts user sessions
	TelegramUserSessions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_user_sessions_total",
			Help: "Total number of user sessions",
		},
		[]string{"user_type"}, // new, returning
	)
)

// RecordTelegramMessage records an incoming message metric
func RecordTelegramMessage(msgType string) {
	TelegramMessagesTotal.WithLabelValues(msgType).Inc()
}

// RecordTelegramError records an error metric
func RecordTelegramError(errorType string) {
	TelegramMessagesError.WithLabelValues(errorType).Inc()
}

// ObserveTelegramResponseTime records response time
func ObserveTelegramResponseTime(msgType string, seconds float64) {
	TelegramResponseTime.WithLabelValues(msgType).Observe(seconds)
}

// RecordUserSession records a user session
func RecordUserSession(isNewUser bool) {
	userType := "returning"
	if isNewUser {
		userType = "new"
	}
	TelegramUserSessions.WithLabelValues(userType).Inc()
}

// UpdateActiveUsers updates the active users counter
func UpdateActiveUsers(count int) {
	TelegramActiveUsers.Set(float64(count))
}
