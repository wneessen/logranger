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
func New(config *Config) (*Server, error) {
	server := &Server{
		conf: config,
	}

	server.setLogLevel()

	if err := server.setRules(); err != nil {
		return server, err
	}

	parser, err := parsesyslog.New(server.conf.internal.ParserType)
	if err != nil {
		return server, fmt.Errorf("failed to initialize syslog parser: %w", err)
	}
	server.parser = parser

	if len(actions.Actions) <= 0 {
		return server, fmt.Errorf("no action plugins found/configured")
	}

	return server, nil
}

// Run starts the logranger Server by creating a new listener using the NewListener
// method and calling RunWithListener with the obtained listener.
func (s *Server) Run() error {
	listener, err := NewListener(s.conf)
	if err != nil {
		return err
	}
	return s.RunWithListener(listener)
}

// RunWithListener sets the listener for the server and performs some additional
// tasks for initializing the server. It creates a PID file, writes the process ID
// to the file, and listens for connections. It returns an error if any of the
// initialization steps fail.
func (s *Server) RunWithListener(listener net.Listener) error {
	s.listener = listener

	// Create PID file
	pidFile, err := os.Create(s.conf.Server.PIDFile)
	if err != nil {
		s.log.Error("failed to create PID file", LogErrKey, err)
		os.Exit(1)
	}
	pid := os.Getpid()
	s.log.Debug("creating PID file", slog.String("pid_file", pidFile.Name()),
		slog.Int("pid", pid))
	_, err = pidFile.WriteString(fmt.Sprintf("%d", pid))
	if err != nil {
		s.log.Error("failed to write PID to PID file", LogErrKey, err)
		_ = pidFile.Close()
	}
	if err = pidFile.Close(); err != nil {
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
		acceptConn, err := s.listener.Accept()
		if err != nil {
			s.log.Error("failed to accept new connection", LogErrKey, err)
			continue
		}
		s.log.Debug("accepted new connection",
			slog.String("remote_addr", acceptConn.RemoteAddr().String()))
		connection := NewConnection(acceptConn)
		s.wg.Add(1)
		go func(co *Connection) {
			s.HandleConnection(co)
			s.wg.Done()
		}(connection)
	}
}

// HandleConnection handles a single connection by parsing and processing log messages.
// It logs debug information about the connection and measures the processing time.
// It closes the connection when done, and logs any error encountered during the process.
func (s *Server) HandleConnection(connection *Connection) {
	defer func() {
		if err := connection.conn.Close(); err != nil {
			s.log.Error("failed to close connection", LogErrKey, err)
		}
	}()

ReadLoop:
	for {
		if err := connection.conn.SetDeadline(time.Now().Add(s.conf.Parser.Timeout)); err != nil {
			s.log.Error("failed to set processing deadline", LogErrKey, err,
				slog.Duration("timeout", s.conf.Parser.Timeout))
			return
		}
		logMessage, err := s.parser.ParseReader(connection.rb)
		if err != nil {
			var netErr *net.OpError
			switch {
			case errors.As(err, &netErr):
				if s.conf.Log.Extended {
					s.log.Error("network error while processing message", LogErrKey,
						netErr.Error())
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
		s.wg.Add(1)
		go s.processMessage(logMessage)
	}
}

// processMessage processes a log message by matching it against the ruleset and executing
// the corresponding actions if a match is found. It takes a parsesyslog.LogMsg as input
// and returns an error if there was an error while processing the actions.
// The method first checks if the ruleset is not nil. If it is nil, no actions will be
// executed. For each rule in the ruleset, it checks if the log message matches the
// rule's regular expression.
func (s *Server) processMessage(logMessage parsesyslog.LogMsg) {
	defer s.wg.Done()
	if s.ruleset != nil {
		for _, rule := range s.ruleset.Rule {
			if !rule.Regexp.MatchString(logMessage.Message.String()) {
				continue
			}
			if rule.HostMatch != nil && !rule.HostMatch.MatchString(logMessage.Hostname) {
				continue
			}
			matchGroup := rule.Regexp.FindStringSubmatch(logMessage.Message.String())
			for name, action := range actions.Actions {
				startTime := time.Now()
				if err := action.Config(rule.Actions); err != nil {
					s.log.Error("failed to config action", LogErrKey, err,
						slog.String("action", name), slog.String("rule_id", rule.ID))
					continue
				}
				s.log.Debug("log message matches rule, executing action",
					slog.String("action", name), slog.String("rule_id", rule.ID))
				if err := action.Process(logMessage, matchGroup); err != nil {
					s.log.Error("failed to process action", LogErrKey, err,
						slog.String("action", name), slog.String("rule_id", rule.ID))
				}
				if s.conf.Log.Extended {
					procTime := time.Since(startTime)
					s.log.Debug("action processing benchmark",
						slog.Duration("processing_time", procTime),
						slog.String("processing_time_human", procTime.String()),
						slog.String("action", name), slog.String("rule_id", rule.ID))
				}
			}
		}
	}
}

// setLogLevel sets the log level based on the value of `s.conf.Log.Level`.
// It creates a new `slog.HandlerOptions` and assigns the corresponding `slog.Level`
// based on the value of `s.conf.Log.Level`. If the value is not one of the valid levels,
// `info` is used as the default level.
// It then creates a new `slog.JSONHandler` with `os.Stdout` and the handler options.
// Finally, it creates a new `slog.Logger` with the JSON handler and sets the `s.log` field
// of the `Server` struct to the logger, with a context value of "logranger".
func (s *Server) setLogLevel() {
	logOpts := slog.HandlerOptions{}
	switch strings.ToLower(s.conf.Log.Level) {
	case "debug":
		logOpts.Level = slog.LevelDebug
	case "info":
		logOpts.Level = slog.LevelInfo
	case "warn":
		logOpts.Level = slog.LevelWarn
	case "error":
		logOpts.Level = slog.LevelError
	default:
		logOpts.Level = slog.LevelInfo
	}
	logHandler := slog.NewJSONHandler(os.Stdout, &logOpts)
	s.log = slog.New(logHandler).With(slog.String("context", "logranger"))
}

// setRules initializes/updates the ruleset for the logranger Server by
// calling NewRuleset with the config and assigns the returned ruleset
// to the Server's ruleset field.
// It returns an error if there is a failure in reading or loading the ruleset.
func (s *Server) setRules() error {
	ruleset, err := NewRuleset(s.conf)
	if err != nil {
		return fmt.Errorf("failed to read ruleset: %w", err)
	}
	s.ruleset = ruleset
	return nil
}

// ReloadConfig reloads the configuration of the Server with the specified
// path and filename.
// It creates a new Config using the NewConfig method and updates the Server's
// conf field. It also reloads the configured Ruleset.
// If an error occurs while reloading the configuration, an error is returned.
func (s *Server) ReloadConfig(path, file string) error {
	config, err := NewConfig(path, file)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}
	s.conf = config

	if err := s.setRules(); err != nil {
		return fmt.Errorf("failed to reload ruleset: %w", err)
	}

	return nil
}
