package service

import "time"

// DeleteConfig controls batching, buffering and the maximum number of fan-in
// writers. Zero-valued fields keep their defaults; negative values are invalid.
type DeleteConfig struct {
	BatchSize     int
	FlushInterval time.Duration
	MaxWorkers    int
	StreamBuffer  int
	QueryTimeout  time.Duration
}

func DefaultDeleteConfig() DeleteConfig {
	return DeleteConfig{
		BatchSize:     100,
		FlushInterval: 500 * time.Millisecond,
		MaxWorkers:    8,
		StreamBuffer:  64,
		QueryTimeout:  5 * time.Second,
	}
}

type Option func(*ShortLinkService)

func WithDeleteConfig(config DeleteConfig) Option {
	return func(service *ShortLinkService) {
		if config.BatchSize < 0 ||
			config.FlushInterval < 0 ||
			config.MaxWorkers < 0 ||
			config.StreamBuffer < 0 ||
			config.QueryTimeout < 0 {
			panic("service: deletion settings must not be negative")
		}

		if config.BatchSize != 0 {
			service.deleteConfig.BatchSize = config.BatchSize
		}

		if config.FlushInterval != 0 {
			service.deleteConfig.FlushInterval = config.FlushInterval
		}

		if config.MaxWorkers != 0 {
			service.deleteConfig.MaxWorkers = config.MaxWorkers
		}

		if config.StreamBuffer != 0 {
			service.deleteConfig.StreamBuffer = config.StreamBuffer
		}

		if config.QueryTimeout != 0 {
			service.deleteConfig.QueryTimeout = config.QueryTimeout
		}
	}
}
