package topmind

import "time"

type Option func(c *client)

func WithLogSuccess() Option {
	return func(c *client) {
		c.LogSuccess = true
	}
}

func WithRetry(maxAttempt int, interval time.Duration) Option {
	return func(c *client) {
		c.RetryMaxAttempt = maxAttempt
		c.RetryInterval = interval
	}
}