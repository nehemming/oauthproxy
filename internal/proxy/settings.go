/*
Copyright © 2018-2021 Neil Hemming
*/
package proxy

import (
	"errors"
	"fmt"
	"strings"
	"time"

	multierror "github.com/hashicorp/go-multierror"
)

const (
	CacheTTLMinValue            = 10 * time.Minute
	RequestTimeoutMinValue      = 10 * time.Second
	ShutdownGracePeriodMinValue = 5 * time.Second
)

type (
	// LoggerFunc logging function
	LoggerFunc func(bool, string, ...interface{})

	// Settings contains the proxy services settings
	Settings struct {
		// CacheTTL how long a item remains valid in the cache
		CacheTTL time.Duration

		// RequestTimeout timeout period for a down sstream request
		RequestTimeout time.Duration

		// ShutdownGracePeriod how long to wait for shutdown
		ShutdownGracePeriod time.Duration

		// HttpListenAddr address and port to listen on
		HttpListenAddr string

		// Downstream endpoint
		Endpoint string

		// Logger recices bogging messages from the service
		Logger LoggerFunc

		// PoolSize is the number of go routines servicing downstream requests
		PoolSize int
	}
)

func DefaultSettings() Settings {
	return Settings{
		CacheTTL:            20 * time.Minute,
		RequestTimeout:      30 * time.Second,
		ShutdownGracePeriod: ShutdownGracePeriodMinValue,
		HttpListenAddr:      "127.0.0.1:8090",
		PoolSize:            2,
	}
}

func (settings Settings) validateSettings() error {

	var result error

	if settings.CacheTTL < CacheTTLMinValue {
		result = multierror.Append(result, fmt.Errorf("cache TTL must be longer than %d minutes", CacheTTLMinValue/time.Minute))
	}

	if settings.CacheTTL < RequestTimeoutMinValue {
		result = multierror.Append(result, fmt.Errorf("request time out must be longer than %d seconds", CacheTTLMinValue/time.Second))
	}

	if settings.ShutdownGracePeriod < ShutdownGracePeriodMinValue {
		result = multierror.Append(result, fmt.Errorf("seervice shutdown grace period must be longer than %d seconds", ShutdownGracePeriodMinValue/time.Second))
	}

	if settings.HttpListenAddr == "" {
		result = multierror.Append(result, errors.New("no listen address provided"))
	}

	if settings.Endpoint == "" {
		result = multierror.Append(result, errors.New("endpoint cannot be blank"))
	}

	if settings.PoolSize < 1 {
		result = multierror.Append(result, fmt.Errorf("pool size must be bigger than %d", 1))
	}

	return result
}

// WithEndpoint sets the down stream oauth services endpoint, the request path is appended to this setting
func (settings Settings) WithEndpoint(endpoint string) Settings {
	if endpoint != "" {
		settings.Endpoint = endpoint
	}

	return settings
}

func (settings Settings) WithHttpPort(port uint) Settings {
	if port != 0 {
		parts := strings.Split(settings.HttpListenAddr, ":")
		if len(parts) == 2 {
			settings.HttpListenAddr = fmt.Sprintf("%s:%d", parts[0], port)
		} else {
			settings.HttpListenAddr = fmt.Sprintf("127.0.0.1:%d", port)
		}
	}

	return settings
}

func (settings Settings) WithLogger(logger LoggerFunc) Settings {
	if logger != nil {
		settings.Logger = logger
	}

	return settings
}
