// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package logranger

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wneessen/go-parsesyslog"
	_ "github.com/wneessen/go-parsesyslog/rfc3164"
	_ "github.com/wneessen/go-parsesyslog/rfc5424"

	"github.com/wneessen/logranger/plugins/actions"
	_ "github.com/wneessen/logranger/plugins/actions/all"
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
	// parser is a parsesyslog.Parser
	parser parsesyslog.Parser
	// ruleset is a pointer to the ruleset
	ruleset *Ruleset
	// wg is a sync.WaitGroup
	wg sync.WaitGroup
}

// New creates a new instance of Server based on the provided Config
func New(c *Config) (*Server, error) {
	s := &Server{
		conf: c,
	}

	s.setLogLevel()

	if err := s.setRules(); err != nil {
		return s, err
	}

	p, err := parsesyslog.New(s.conf.internal.ParserType)
	if err != nil {
		return s, fmt.Errorf("failed to initialize syslog parser: %w", err)
	}
	s.parser = p

	if len(actions.Actions) <= 0 {
		return s, fmt.Errorf("no action plugins found/configured")
	}

	return s, nil
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

// Listen handles incoming connections and processes log messages.
func (s *Server) Listen() {
	defer s.wg.Done()
	s.log.Info("listening for new connections", slog.String("listen_addr", s.listener.Addr().String()))
	for {
		c, err := s.listener.Accept()
		if err != nil {
			s.log.Error("failed to accept new connection", LogErrKey, err)
			continue
		}
		s.log.Debug("accepted new connection",
			slog.String("remote_addr", c.RemoteAddr().String()))
		conn := NewConnection(c)
		s.wg.Add(1)
		go func(co *Connection) {
			s.HandleConnection(co)
			s.wg.Done()
		}(conn)
	}
}

// HandleConnection handles a single connection by parsing and processing log messages.
// It logs debug information about the connection and measures the processing time.
// It closes the connection when done, and logs any error encountered during the process.
func (s *Server) HandleConnection(c *Connection) {
	defer func() {
		if err := c.conn.Close(); err != nil {
			s.log.Error("failed to close connection", LogErrKey, err)
		}
	}()

ReadLoop:
	for {
		if err := c.conn.SetDeadline(time.Now().Add(s.conf.Parser.Timeout)); err != nil {
			s.log.Error("failed to set processing deadline", LogErrKey, err,
				slog.Duration("timeout", s.conf.Parser.Timeout))
			return
		}
		lm, err := s.parser.ParseReader(c.rb)
		if err != nil {
			var ne *net.OpError
			switch {
			case errors.As(err, &ne):
				if s.conf.Log.Extended {
					s.log.Error("network error while processing message", LogErrKey,
						ne.Error())
				}
				return
			case errors.Is(err, io.EOF):
				if s.conf.Log.Extended {
					s.log.Error("message could not be processed", LogErrKey,
						"EOF received")
				}
				return
			default:
				s.log.Error("failed to parse message", LogErrKey, err,
					slog.String("parser_type", s.conf.Parser.Type))
				continue ReadLoop
			}
		}
		if err = s.processMessage(lm); err != nil {
			s.log.Error("failed to process actions on log message", LogErrKey, err)
		}
	}
}

func (s *Server) processMessage(lm parsesyslog.LogMsg) error {
	if s.ruleset != nil {
		for _, r := range s.ruleset.Rule {
			if !r.Regexp.MatchString(lm.Message.String()) {
				continue
			}
			if r.HostMatch != nil && !r.HostMatch.MatchString(lm.Hostname) {
				continue
			}
			mg := r.Regexp.FindStringSubmatch(lm.Message.String())
			for n, a := range actions.Actions {
				s.log.Debug("trying to execute action", slog.String("action_name", n))
				if err := a.Process(lm, mg, r.Actions); err != nil {
					s.log.Error("failed to process action", LogErrKey, err,
						slog.String("action", n), slog.String("rule_id", r.ID))
				}
			}
		}
	}
	return nil
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

// setRules initializes/updates the ruleset for the logranger Server by
// calling NewRuleset with the config and assigns the returned ruleset
// to the Server's ruleset field.
// It returns an error if there is a failure in reading or loading the ruleset.
func (s *Server) setRules() error {
	rs, err := NewRuleset(s.conf)
	if err != nil {
		return fmt.Errorf("failed to read ruleset: %w", err)
	}
	s.ruleset = rs
	return nil
}
