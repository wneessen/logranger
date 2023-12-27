// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package main

import (
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/wneessen/logranger"
)

const (
	// LogErrKey is the keyword used in slog for error messages
	LogErrKey = "error"
)

func main() {
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(slog.String("context", "logranger"))
	cp := "logranger.toml"
	cpe := os.Getenv("LOGRANGER_CONFIG")
	if cpe != "" {
		cp = cpe
	}

	p := filepath.Dir(cp)
	f := filepath.Base(cp)
	c, err := logranger.NewConfig(p, f)
	if err != nil {
		l.Error("failed to read/parse config", LogErrKey, err)
		os.Exit(1)
	}

	s, err := logranger.New(c)
	if err != nil {
		l.Error("failed to create new server", LogErrKey, err)
		os.Exit(1)
	}

	go func() {
		if err = s.Run(); err != nil {
			l.Error("failed to start logranger", LogErrKey, err)
			os.Exit(1)
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc)
	for rc := range sc {
		if rc == syscall.SIGKILL || rc == syscall.SIGABRT || rc == syscall.SIGINT || rc == syscall.SIGTERM {
			l.Warn("received signal. shutting down server", slog.String("signal", rc.String()))
			// s.Stop()
			l.Info("server gracefully shut down")
			os.Exit(0)
		}
		if rc == syscall.SIGHUP {
			l.Info(`received signal`,
				slog.String("signal", "SIGHUP"),
				slog.String("action", "reloading config/ruleset"))
			if err = s.ReloadConfig(p, f); err != nil {
				l.Error("failed to reload config", LogErrKey, err)
			}
		}
	}
}
