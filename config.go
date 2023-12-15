// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package logranger

import (
	"fmt"
	"os"

	"github.com/kkyr/fig"
)

// Config holds all the global configuration settings that are parsed by fig
type Config struct {
	// Server holds server specific configuration values
	Server struct {
		PIDFile string `fig:"pid_file" default:"/var/run/logranger.pid"`
	}
	Listener struct {
		ListenerUnix struct {
			Path string `fig:"path" default:"/var/tmp/logranger.sock"`
		} `fig:"unix"`
		ListenerTCP struct {
			Addr string `fig:"addr" default:"0.0.0.0"`
			Port uint   `fig:"port" default:"9099"`
		} `fig:"tcp"`
		ListenerTLS struct {
			Addr     string `fig:"addr" default:"0.0.0.0"`
			Port     uint   `fig:"port" default:"9099"`
			CertPath string `fig:"cert_path"`
			KeyPath  string `fig:"key_path"`
		} `fig:"tls"`
		Type ListenerType `fig:"type" default:"unix"`
	} `fig:"listener"`
	Log struct {
		Level string `fig:"level" default:"info"`
	} `fig:"log"`
}

// NewConfig returns a new Config object
func NewConfig(p, f string) (*Config, error) {
	co := Config{}
	_, err := os.Stat(fmt.Sprintf("%s/%s", p, f))
	if err != nil {
		return &co, fmt.Errorf("failed to read config: %w", err)
	}

	if err := fig.Load(&co, fig.Dirs(p), fig.File(f), fig.UseEnv("logranger")); err != nil {
		return &co, fmt.Errorf("failed to load config: %w", err)
	}

	return &co, nil
}
