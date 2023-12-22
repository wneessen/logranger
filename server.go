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

	"github.com/wneessen/go-parsesyslog"
	_ "github.com/wneessen/go-parsesyslog/rfc3164"
	"github.com/wneessen/go-parsesyslog/rfc5424"
	_ "github.com/wneessen/go-parsesyslog/rfc5424"
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

// New creates a new instance of Server based on the provided Config
func New(c *Config) *Server {
	s := &Server{
		conf: c,
	}
	s.setLogLevel()
	return s
}

// Run starts the logranger Server by creating a new listener using the NewListener
// method and calling RunWithListener with the obtained listener.
func (s *Server) Run() error {
	l, err := NewListener(s.conf)
	if err != nil {
		return err
	}
	return s.RunWithListener(l)
}

// RunWithListener sets the listener for the server and performs some additional
// tasks for initializing the server. It creates a PID file, writes the process ID
// to the file, and listens for connections. It returns an error if any of the
// initialization steps fail.
func (s *Server) RunWithListener(l net.Listener) error {
	s.listener = l

	// Create PID file
	pf, err := os.Create(s.conf.Server.PIDFile)
	if err != nil {
		s.log.Error("failed to create PID file", LogErrKey, err)
		os.Exit(1)
	}
	pid := os.Getpid()
	s.log.Debug("creating PID file", slog.String("pid_file", pf.Name()),
		slog.Int("pid", pid))
	_, err = pf.WriteString(fmt.Sprintf("%d", pid))
	if err != nil {
		s.log.Error("failed to write PID to PID file", LogErrKey, err)
		_ = pf.Close()
	}
	if err = pf.Close(); err != nil {
		s.log.Error("failed to close PID file", LogErrKey, err)
	}

	// Listen for connections
	s.wg.Add(1)
	go s.Listen()

	return nil
}

func (s *Server) Listen() {
	defer s.wg.Done()
	s.log.Info("listening for new connections", slog.String("listen_addr", s.listener.Addr().String()))
	for {
		c, err := s.listener.Accept()
		if err != nil {
			s.log.Error("failed to accept new connection", LogErrKey, err)
			continue
		}
		s.log.Debug("accepted new connection", slog.String("remote_addr", c.RemoteAddr().String()))
		conn := NewConnection(c)
		s.wg.Add(1)
		go func(co *Connection) {
			s.HandleConnection(co)
			s.wg.Done()
		}(conn)
	}
}

func (s *Server) HandleConnection(c *Connection) {
	s.log.Debug("handling connection")
	defer c.conn.Close()
	pa, err := parsesyslog.New(rfc5424.Type)
	if err != nil {
		s.log.Error("failed to initialize logger", LogErrKey, err)
		return
	}
	lm, err := pa.ParseReader(c.rb)
	if err != nil {
		s.log.Error("failed to parse message", LogErrKey, err)
		return
	}
	s.log.Info("log message received", slog.String("message", lm.Message.String()))

}

// setLogLevel sets the log level based on the value of `s.conf.Log.Level`.
// It creates a new `slog.HandlerOptions` and assigns the corresponding `slog.Level`
// based on the value of `s.conf.Log.Level`. If the value is not one of the valid levels,
// `info` is used as the default level.
// It then creates a new `slog.JSONHandler` with `os.Stdout` and the handler options.
// Finally, it creates a new `slog.Logger` with the JSON handler and sets the `s.log` field
// of the `Server` struct to the logger, with a context value of "logranger".
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
