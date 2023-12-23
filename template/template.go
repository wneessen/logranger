// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package template

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/wneessen/go-parsesyslog"
)

// SHAAlgo is a enum-like type wrapper representing a SHA algorithm
type SHAAlgo uint

const (
	// SHA1 is a constant of type SHAAlgo, representing the SHA-1 algorithm.
	SHA1 SHAAlgo = iota
	// SHA256 is a constant of type SHAAlgo, representing the SHA-256 algorithm.
	SHA256
	// SHA512 is a constant of type SHAAlgo, representing the SHA-512 algorithm.
	SHA512
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
		"_ToLower":  fm.ToLower,
		"_ToUpper":  fm.ToUpper,
		"_ToBase64": fm.ToBase64,
		"_ToSHA1":   fm.ToSHA1,
		"_ToSHA256": fm.ToSHA256,
		"_ToSHA512": fm.ToSHA512,
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

// ToBase64 returns the base64 encoding of a given string.
func (*FuncMap) ToBase64(s string) string {
	return base64.RawStdEncoding.EncodeToString([]byte(s))
}

// ToSHA1 returns the SHA-1 hash of the given string
func (*FuncMap) ToSHA1(s string) string {
	return toSHA(s, SHA1)
}

// ToSHA256 returns the SHA-256 hash of the given string
func (*FuncMap) ToSHA256(s string) string {
	return toSHA(s, SHA256)
}

// ToSHA512 returns the SHA-512 hash of the given string
func (*FuncMap) ToSHA512(s string) string {
	return toSHA(s, SHA512)
}

// toSHA is a function that converts a string to a SHA hash.
//
// The function takes two parameters: a string 's' and a 'sa' of
// type SHAAlgo which defines the SHA algorithm to be used.
func toSHA(s string, sa SHAAlgo) string {
	var h hash.Hash
	switch sa {
	case SHA1:
		h = sha1.New()
	case SHA256:
		h = sha256.New()
	case SHA512:
		h = sha512.New()
	default:
		return ""
	}

	_, err := io.WriteString(h, s)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
