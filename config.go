// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package logranger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kkyr/fig"
	"github.com/wneessen/go-parsesyslog"
	"github.com/wneessen/go-parsesyslog/rfc3164"
	"github.com/wneessen/go-parsesyslog/rfc5424"
)

// Config holds all the global configuration settings that are parsed by fig
type Config struct {
	// Server holds server specific configuration values
	Server struct {
		PIDFile  string `fig:"pid_file" default:"/var/run/logranger.pid"`
		RuleFile string `fig:"rule_file" default:"etc/logranger.rules.toml"`
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
		Level    string `fig:"level" default:"info"`
		Extended bool   `fig:"extended"`
	} `fig:"log"`
	Parser struct {
		Type    string        `fig:"type" validate:"required"`
		Timeout time.Duration `fig:"timeout" default:"500ms"`
	} `fig:"parser"`
	internal struct {
		ParserType parsesyslog.ParserType
	}
}

// NewConfig creates a new instance of the Config object by reading and loading
// configuration values. It takes in the file path and file name of the configuration
// file as parameters. It returns a pointer to the Config object and an error if
// there was a problem reading or loading the configuration.
func NewConfig(p, f string) (*Config, error) {
	co := Config{}
	_, err := os.Stat(fmt.Sprintf("%s/%s", p, f))
	if err != nil {
		return &co, fmt.Errorf("failed to read config: %w", err)
	}

	if err := fig.Load(&co, fig.Dirs(p), fig.File(f), fig.UseEnv("logranger")); err != nil {
		return &co, fmt.Errorf("failed to load config: %w", err)
	}

	switch {
	case strings.EqualFold(co.Parser.Type, "rfc3164"):
		co.internal.ParserType = rfc3164.Type
	case strings.EqualFold(co.Parser.Type, "rfc5424"):
		co.internal.ParserType = rfc5424.Type
	default:
		return nil, fmt.Errorf("unknown parser type: %s", co.Parser.Type)
	}

	return &co, nil
}
