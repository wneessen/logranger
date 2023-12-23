// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package plugins

import (
	"github.com/wneessen/go-parsesyslog"
)

// Action is an interface that defines the behavior of an action to be performed
// on a log message.
//
// The Process method takes a log message, a slice of match groups, and a
// configuration map, and returns an error if any occurs during processing.
type Action interface {
	Process(logmessage parsesyslog.LogMsg, matchgroup []string, confmap map[string]any) error
}
