package telegram // import "github.com/robertgzr/joe-telegram-adapter"

import "go.uber.org/zap"

type Option func(*Config) error

func WithLogger(logger *zap.Logger) Option {
	return func(conf *Config) error {
		conf.Logger = logger
		return nil
	}
}

func WithUpdateTimeout(secs int) Option {
	return func(conf *Config) error {
		conf.UpdateTimeoutSec = secs
		return nil
	}
}
