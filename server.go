// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package logranger

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	// LogErrKey is the keyword used in slog for error messages
	LogErrKey = "error"
)

// Server is the main server struct
type Server struct {
	// conf is a pointer to the config.Config
	conf *Config
	// listener is a listener that satisfies the net.Listener interface
	listener net.Listener
	// log is a pointer to the slog.Logger
	log *slog.Logger

	// wg is a sync.WaitGroup
	wg sync.WaitGroup
}

// New returns a Server struct
func New(c *Config) *Server {
	s := &Server{
		conf: c,
	}
	s.setLogLevel()
	return s
}

// Run starts the logranger Server with a new Listener based on the config settings
func (s *Server) Run() error {
	l, err := NewListener(s.conf)
	if err != nil {
		return err
	}
	return s.RunWithListener(l)
}

// RunWithListener starts the logranger Server using a provided net.Listener
func (s *Server) RunWithListener(l net.Listener) error {
	s.listener = l

	// Create PID file
	pf, err := os.Create(s.conf.Server.PIDFile)
	if err != nil {
		s.log.Error("failed to create PID file", LogErrKey, err)
		os.Exit(1)
	}
	_, err = pf.WriteString(fmt.Sprintf("%d", os.Getpid()))
	if err != nil {
		s.log.Error("failed to write PID to PID file", LogErrKey, err)
		_ = pf.Close()
	}
	if err = pf.Close(); err != nil {
		s.log.Error("failed to close PID file", LogErrKey, err)
	}

	// Listen for connections
	s.wg.Add(1)

	return nil
}

// setLogLevel assigns a new slog.Logger instance to the Server based on the configured log level
func (s *Server) setLogLevel() {
	lo := slog.HandlerOptions{}
	switch strings.ToLower(s.conf.Log.Level) {
	case "debug":
		lo.Level = slog.LevelDebug
	case "info":
		lo.Level = slog.LevelInfo
	case "warn":
		lo.Level = slog.LevelWarn
	case "error":
		lo.Level = slog.LevelError
	default:
		lo.Level = slog.LevelInfo
	}
	lh := slog.NewJSONHandler(os.Stdout, &lo)
	s.log = slog.New(lh).With(slog.String("context", "logranger"))
}
