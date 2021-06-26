package telegram // import "github.com/robertgzr/joe-telegram-adapter"

import "go.uber.org/zap"

type Option func(*Config) error

// WithLogger allows configuring a custom zap.Logger for the adapter to use
func WithLogger(logger *zap.Logger) Option {
	return func(conf *Config) error {
		conf.Logger = logger
		return nil
	}
}

// WithUpdateTimeout allows configuring the update loop timeout (in seconds)
func WithUpdateTimeout(secs int) Option {
	return func(conf *Config) error {
		conf.UpdateTimeoutSec = secs
		return nil
	}
}

// WithUpdateResumeFrom allows setting the starting Update ID from which to
// process updates froms.
func WithUpdateResumeFrom(id int) Option {
	return func(conf *Config) error {
		conf.UpdateResumeFrom = id
		return nil
	}
}
