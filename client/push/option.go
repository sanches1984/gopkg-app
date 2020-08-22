package push

import "github.com/prometheus/client_golang/prometheus"

type Option func(c *Client)

func WithMetricSuccess(metric prometheus.Counter) Option {
	return func(c *Client) {
		c.metricSuccess = &metric
	}
}

func WithMetricFailed(metric prometheus.Counter) Option {
	return func(c *Client) {
		c.metricFailed = &metric
	}
}

func WithInfoLog() Option {
	return func(c *Client) {
		c.showInfo = true
	}
}
