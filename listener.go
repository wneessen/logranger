// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package logranger

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
)

// ListenerType is an enumeration wrapper for the different listener types
type ListenerType uint

const (
	// ListenerUnix is a constant of type ListenerType that represents a UNIX listener.
	ListenerUnix ListenerType = iota
	// ListenerTCP is a constant representing the type of listener that uses TCP protocol.
	ListenerTCP
	// ListenerTLS is a constant of type ListenerType that represents a TLS listener.
	ListenerTLS
)

// NewListener initializes and returns a net.Listener based on the provided
// configuration. It takes a pointer to a Config struct as a parameter.
// Returns the net.Listener and an error if any occurred during initialization.
func NewListener(config *Config) (net.Listener, error) {
	var listener net.Listener
	var listenerErr error
	switch config.Listener.Type {
	case ListenerUnix:
		resolveUnixAddr, err := net.ResolveUnixAddr("unix", config.Listener.ListenerUnix.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve UNIX listener socket: %w", err)
		}
		listener, listenerErr = net.Listen("unix", resolveUnixAddr.String())
	case ListenerTCP:
		listenAddr := net.JoinHostPort(config.Listener.ListenerTCP.Addr,
			fmt.Sprintf("%d", config.Listener.ListenerTCP.Port))
		listener, listenerErr = net.Listen("tcp", listenAddr)
	case ListenerTLS:
		if config.Listener.ListenerTLS.CertPath == "" || config.Listener.ListenerTLS.KeyPath == "" {
			return nil, ErrCertConfigEmpty
		}
		cert, err := tls.LoadX509KeyPair(config.Listener.ListenerTLS.CertPath, config.Listener.ListenerTLS.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load X509 certificate: %w", err)
		}
		listenAddr := net.JoinHostPort(config.Listener.ListenerTLS.Addr, fmt.Sprintf("%d", config.Listener.ListenerTLS.Port))
		listenConf := &tls.Config{Certificates: []tls.Certificate{cert}}
		listener, listenerErr = tls.Listen("tcp", listenAddr, listenConf)
	default:
		return nil, fmt.Errorf("failed to initialize listener: unknown listener type in config")
	}
	if listenerErr != nil {
		return nil, fmt.Errorf("failed to initialize listener: %w", listenerErr)
	}
	return listener, nil
}

// UnmarshalString satisfies the fig.StringUnmarshaler interface for the ListenerType type
func (l *ListenerType) UnmarshalString(value string) error {
	switch strings.ToLower(value) {
	case "unix":
		*l = ListenerUnix
	case "tcp":
		*l = ListenerTCP
	case "tls":
		*l = ListenerTLS
	default:
		return fmt.Errorf("unknown listener type: %s", value)
	}
	return nil
}

// String satisfies the fmt.Stringer interface for the ListenerType type
func (l ListenerType) String() string {
	switch l {
	case ListenerUnix:
		return "UNIX listener"
	case ListenerTCP:
		return "TCP listener"
	case ListenerTLS:
		return "TLS listener"
	default:
		return "Unknown listener type"
	}
}
