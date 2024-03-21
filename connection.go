// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package logranger

import (
	"bufio"
	"fmt"
	"math/rand"
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
func NewConnection(netConn net.Conn) *Connection {
	connection := &Connection{
		conn: netConn,
		id:   NewConnectionID(),
		rb:   bufio.NewReader(netConn),
		wb:   bufio.NewWriter(netConn),
	}
	return connection
}

// NewConnectionID generates a new unique message ID using a random number generator
// and returns it as a hexadecimal string.
func NewConnectionID() string {
	return fmt.Sprintf("%x", rand.Int63())
}
