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
	ListenerUnix ListenerType = iota
	ListenerTCP
	ListenerTLS
)

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
		la := net.JoinHostPort(c.Listener.ListenerTCP.Addr, fmt.Sprintf("%d", c.Listener.ListenerTCP.Port))
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
