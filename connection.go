// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package logranger

import (
	"bufio"
	"net"
)

// Connection represents a connection to a network resource.
type Connection struct {
	conn net.Conn
	id   string
	rb   *bufio.Reader
	wb   *bufio.Writer
}

// NewConnection creates a new Connection object with the provided net.Conn.
// The Connection object holds a reference to the provided net.Conn, along with an ID string,
// bufio.Reader, and bufio.Writer. It returns a pointer to the created Connection object.
func NewConnection(nc net.Conn) *Connection {
	c := &Connection{
		conn: nc,
		id:   "foo",
		rb:   bufio.NewReader(nc),
		wb:   bufio.NewWriter(nc),
	}
	return c
}
