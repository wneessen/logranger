// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package template

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/wneessen/go-parsesyslog"
)

// FuncMap represents a mapping of function names to their corresponding
// functions.
// It is used to define custom functions that can be accessed in Go
// templates.
type FuncMap struct{}

// Compile compiles a template string using a given LogMsg, match group,
// and output template.
// It replaces special characters in the output template and creates a
// new template, named "template", with custom template functions from
// the FuncMap. It then populates a map with values from the LogMsg
// and current time and executes the template using the map as the
// data source. The compiled template result or an error is returned.
func Compile(lm parsesyslog.LogMsg, mg []string, ot string) (string, error) {
	pt := strings.Builder{}
	fm := NewTemplateFuncMap()

	ot = strings.ReplaceAll(ot, `\n`, "\n")
	ot = strings.ReplaceAll(ot, `\t`, "\t")
	ot = strings.ReplaceAll(ot, `\r`, "\r")
	tpl, err := template.New("template").Funcs(fm).Parse(ot)
	if err != nil {
		return pt.String(), fmt.Errorf("failed to create template: %w", err)
	}

	dm := make(map[string]any)
	dm["match"] = mg
	dm["hostname"] = lm.Hostname
	dm["timestamp"] = lm.Timestamp
	dm["now_rfc3339"] = time.Now().Format(time.RFC3339)
	dm["now_unix"] = time.Now().Unix()
	dm["severity"] = lm.Severity.String()
	dm["facility"] = lm.Facility.String()
	dm["appname"] = lm.AppName
	dm["original_message"] = lm.Message

	if err = tpl.Execute(&pt, dm); err != nil {
		return pt.String(), fmt.Errorf("failed to compile template: %w", err)
	}
	return pt.String(), nil
}

// NewTemplateFuncMap creates a new template function map by returning a
// template.FuncMap.
func NewTemplateFuncMap() template.FuncMap {
	fm := FuncMap{}
	return template.FuncMap{
		"_ToLower": fm.ToLower,
	}
}

// ToLower returns a given string as lower-case representation
func (*FuncMap) ToLower(s string) string {
	return strings.ToLower(s)
}

// ToUpper returns a given string as upper-case representation
func (*FuncMap) ToUpper(s string) string {
	return strings.ToUpper(s)
}
