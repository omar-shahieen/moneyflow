package ports

import "time"

type Queue interface {
	EnqueueCritical(taskName string, payload []byte, opts ...TaskOption) error
	EnqueueDefault(taskName string, payload []byte, opts ...TaskOption) error
	EnqueueLow(taskName string, payload []byte, opts ...TaskOption) error
}

type TaskOption func(*TaskConfig)

type TaskConfig struct {
	MaxRetry  int
	Timeout   time.Duration
	QueueName string
}

func WithMaxRetry(n int) TaskOption {
	return func(c *TaskConfig) {
		c.MaxRetry = n
	}
}

func WithTimeout(d time.Duration) TaskOption {
	return func(c *TaskConfig) {
		c.Timeout = d
	}
}
