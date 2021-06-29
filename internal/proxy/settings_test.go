/*
Copyright © 2018-2021 Neil Hemming
*/

package proxy

import (
	"testing"
	"time"
)

func TestShutdownGracePeriodMinValue(t *testing.T) {
	if ShutdownGracePeriodMinValue != 5*time.Second {
		t.Errorf("ShutdownGracePeriodMinValue expected %d got %d", 5*time.Second, ShutdownGracePeriodMinValue)
	}
}

func TestDefaultSettings(t *testing.T) {

	settings := DefaultSettings()

	if settings.CacheTTL != 20*time.Minute {
		t.Errorf("CacheTTL expected %d got %d", 20*time.Minute, settings.CacheTTL)
	}
	if settings.RequestTimeout != 30*time.Second {
		t.Errorf("RequestTimeout expected %d got %d", 20*time.Minute, settings.RequestTimeout)
	}
	if settings.ShutdownGracePeriod != ShutdownGracePeriodMinValue {
		t.Errorf("ShutdownGracePeriod expected %d got %d", ShutdownGracePeriodMinValue, settings.ShutdownGracePeriod)
	}
	if settings.HTTPListenAddr != "127.0.0.1:8090" {
		t.Errorf("HTTPListenAddr expected %s got %s", "127.0.0.1:8090", settings.HTTPListenAddr)
	}
	if settings.PoolSize != 2 {
		t.Errorf("PoolSize expected %d got %d", 2, settings.PoolSize)
	}
}

func TestWithEndpoint(t *testing.T) {

	settings := DefaultSettings()

	settingsWithEndpoint := settings.WithEndpoint("test:99")

	if settings.Endpoint != "" {
		t.Errorf("Default Endpoint expected empty got %s", settings.Endpoint)
	}

	if settingsWithEndpoint.Endpoint != "test:99" {
		t.Errorf("Endpoint expected %s got %s", "test:99", settingsWithEndpoint.Endpoint)
	}
}

func TestWithHTTPPort(t *testing.T) {

	settings := DefaultSettings()

	settingsWithPort := settings.WithHTTPPort(9990)

	if settings.HTTPListenAddr != "127.0.0.1:8090" {
		t.Errorf("Default HTTPListenAddr expected %s got %s", "127.0.0.1:8090", settings.HTTPListenAddr)
	}

	if settingsWithPort.HTTPListenAddr != "127.0.0.1:9990" {
		t.Errorf("HTTPListenAddr expected %s got %s", "127.0.0.1:9990", settingsWithPort.HTTPListenAddr)
	}
}

func TestWithHTTPPortWithSinglePart(t *testing.T) {

	settings := DefaultSettings()

	settings.HTTPListenAddr = "8080"

	settingsWithPort := settings.WithHTTPPort(9990)

	if settingsWithPort.HTTPListenAddr != "127.0.0.1:9990" {
		t.Errorf("HTTPListenAddr expected %s got %s", "127.0.0.1:9990", settingsWithPort.HTTPListenAddr)
	}
}

func TestWWithLogger(t *testing.T) {

	settings := DefaultSettings()

	if settings.Logger != nil {
		t.Error("Default logger not nil")
	}

	fn := func(bool, string, ...interface{}) {}

	settingsWithLogger := settings.WithLogger(fn)

	if settingsWithLogger.Logger == nil {
		t.Error("Logger is nil")
	}
}

func TestDefaultSettingsValidateSettingsFails(t *testing.T) {

	settings := DefaultSettings()

	err := settings.validateSettings()

	if err == nil {
		t.Error("DefaultSettings validated, expected an error as no endpoint")
	}
}

func TestValidateSettingsSucceeds(t *testing.T) {

	settings := DefaultSettings().WithEndpoint("test")

	err := settings.validateSettings()

	if err != nil {
		t.Error("DefaultSettings WithEndpoint has error", err)
	}
}

func TestValidateSettingsBadCacheTTLFails(t *testing.T) {

	settings := DefaultSettings().WithEndpoint("test")

	settings.CacheTTL = CacheTTLMinValue - 1

	err := settings.validateSettings()

	if err == nil {
		t.Error("Bad CacheTTL not caught")
	}
}

func TestValidateSettingsBadRequestTimeoutFails(t *testing.T) {

	settings := DefaultSettings().WithEndpoint("test")

	settings.RequestTimeout = RequestTimeoutMinValue - 1

	err := settings.validateSettings()

	if err == nil {
		t.Error("Bad RequestTimeout not caught")
	}
}

func TestValidateSettingsBadPoolSizeFails(t *testing.T) {

	settings := DefaultSettings().WithEndpoint("test")

	settings.PoolSize = 0

	err := settings.validateSettings()

	if err == nil {
		t.Error("Bad RequestTimeout not caught")
	}
}

func TestValidateSettingsBadShutdownGracePeriodFails(t *testing.T) {

	settings := DefaultSettings().WithEndpoint("test")

	settings.ShutdownGracePeriod = ShutdownGracePeriodMinValue - 1

	err := settings.validateSettings()

	if err == nil {
		t.Error("Bad ShutdownGracePeriod not caught")
	}
}

func TestValidateSettingsBadHTTPListenAddrFails(t *testing.T) {

	settings := DefaultSettings().WithEndpoint("test")

	settings.HTTPListenAddr = ""

	err := settings.validateSettings()

	if err == nil {
		t.Error("Bad HTTPListenAddr not caught")
	}
}
