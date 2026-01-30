// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package natsexporter

import (
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/nats/confignats"
)

// Config defines configuration for the NATS exporter.
type Config struct {
	exporterhelper.TimeoutConfig `mapstructure:",squash"`
	configretry.BackOffConfig    `mapstructure:"retry_on_failure"`
	confignats.ClientConfig      `mapstructure:",squash"`

	// Traces configuration.
	Traces SignalConfig `mapstructure:"traces"`

	// Metrics configuration.
	Metrics SignalConfig `mapstructure:"metrics"`

	// Logs configuration.
	Logs SignalConfig `mapstructure:"logs"`
}

// SignalConfig holds signal-specific configuration.
type SignalConfig struct {
	// Subject is the NATS subject to publish to.
	Subject string `mapstructure:"subject"`

	// Encoding for message serialization (default: otlp_proto).
	// Currently only otlp_proto is supported.
	Encoding string `mapstructure:"encoding"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if err := confignats.ValidateURL(c.ClientConfig.URL); err != nil {
		return err
	}

	if err := c.ClientConfig.Auth.Validate(); err != nil {
		return err
	}

	if c.Traces.Subject == "" && c.Metrics.Subject == "" && c.Logs.Subject == "" {
		return errors.New("at least one signal subject must be configured")
	}

	// Validate each signal configuration
	signals := map[string]SignalConfig{
		"traces":  c.Traces,
		"metrics": c.Metrics,
		"logs":    c.Logs,
	}

	for name, cfg := range signals {
		// Validate subject format if configured (no wildcards allowed for publishing)
		if cfg.Subject != "" {
			if err := confignats.ValidatePublishSubject(cfg.Subject); err != nil {
				return errors.New(name + ".subject: " + err.Error())
			}
		}

		// Validate encoding if specified
		if cfg.Encoding != "" && cfg.Encoding != defaultEncoding {
			return errors.New("only otlp_proto encoding is currently supported")
		}
	}

	return nil
}
