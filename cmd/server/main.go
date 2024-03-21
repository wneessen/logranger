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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(slog.String("context", "logranger"))
	confPath := "logranger.toml"
	confPathEnv := os.Getenv("LOGRANGER_CONFIG")
	if confPathEnv != "" {
		confPath = confPathEnv
	}

	path := filepath.Dir(confPath)
	file := filepath.Base(confPath)
	config, err := logranger.NewConfig(path, file)
	if err != nil {
		logger.Error("failed to read/parse config", LogErrKey, err)
		os.Exit(1)
	}

	server, err := logranger.New(config)
	if err != nil {
		logger.Error("failed to create new server", LogErrKey, err)
		os.Exit(1)
	}

	go func() {
		if err = server.Run(); err != nil {
			logger.Error("failed to start logranger", LogErrKey, err)
			os.Exit(1)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan)
	for recvSig := range signalChan {
		if recvSig == syscall.SIGKILL || recvSig == syscall.SIGABRT || recvSig == syscall.SIGINT || recvSig == syscall.SIGTERM {
			logger.Warn("received signal. shutting down server", slog.String("signal", recvSig.String()))
			// server.Stop()
			logger.Info("server gracefully shut down")
			os.Exit(0)
		}
		if recvSig == syscall.SIGHUP {
			logger.Info(`received signal`,
				slog.String("signal", "SIGHUP"),
				slog.String("action", "reloading config/ruleset"))
			if err = server.ReloadConfig(path, file); err != nil {
				logger.Error("failed to reload config", LogErrKey, err)
			}
		}
	}
}
