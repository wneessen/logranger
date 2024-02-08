// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package actions

import (
	"src.neessen.cloud/wneessen/logranger/plugins"
)

// Actions is a variable that represents a map of string keys to Action values. The keys are used to identify different actions, and the corresponding values are the functions that define
var Actions = map[string]plugins.Action{}

// Add adds an action with the given name to the Actions map. The action function must implement the Action interface.
func Add(name string, action plugins.Action) {
	Actions[name] = action
}
