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
func NewListener(c *Config) (net.Listener, error) {
	var l net.Listener
	var lerr error
	switch c.Listener.Type {
	case ListenerUnix:
		rua, err := net.ResolveUnixAddr("unix", c.Listener.ListenerUnix.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve UNIX listener socket: %w", err)
		}
		l, lerr = net.Listen("unix", rua.String())
	case ListenerTCP:
		la := net.JoinHostPort(c.Listener.ListenerTCP.Addr, fmt.Sprintf("%d", c.Listener.ListenerTCP.Port))
		l, lerr = net.Listen("tcp", la)
	case ListenerTLS:
		if c.Listener.ListenerTLS.CertPath == "" || c.Listener.ListenerTLS.KeyPath == "" {
			return nil, ErrCertConfigEmpty
		}
		ce, err := tls.LoadX509KeyPair(c.Listener.ListenerTLS.CertPath, c.Listener.ListenerTLS.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load X509 certificate: %w", err)
		}
		la := net.JoinHostPort(c.Listener.ListenerTLS.Addr, fmt.Sprintf("%d", c.Listener.ListenerTLS.Port))
		lc := &tls.Config{Certificates: []tls.Certificate{ce}}
		l, lerr = tls.Listen("tcp", la, lc)
	default:
		return nil, fmt.Errorf("failed to initialize listener: unknown listener type in config")
	}
	if lerr != nil {
		return nil, fmt.Errorf("failed to initalize listener: %w", lerr)
	}
	return l, nil
}

// UnmarshalString satisfies the fig.StringUnmarshaler interface for the ListenerType type
func (l *ListenerType) UnmarshalString(v string) error {
	switch strings.ToLower(v) {
	case "unix":
		*l = ListenerUnix
	case "tcp":
		*l = ListenerTCP
	case "tls":
		*l = ListenerTLS
	default:
		return fmt.Errorf("unknown listener type: %s", v)
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
