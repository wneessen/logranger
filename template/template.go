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
func Compile(logMessage parsesyslog.LogMsg, matchGroup []string, outputTpl string) (string, error) {
	procText := strings.Builder{}
	funcMap := NewTemplateFuncMap()

	outputTpl = strings.ReplaceAll(outputTpl, `\n`, "\n")
	outputTpl = strings.ReplaceAll(outputTpl, `\t`, "\t")
	outputTpl = strings.ReplaceAll(outputTpl, `\r`, "\r")
	tpl, err := template.New("template").Funcs(funcMap).Parse(outputTpl)
	if err != nil {
		return procText.String(), fmt.Errorf("failed to create template: %w", err)
	}

	dataMap := make(map[string]any)
	dataMap["match"] = matchGroup
	dataMap["hostname"] = logMessage.Hostname
	dataMap["timestamp"] = logMessage.Timestamp
	dataMap["now_rfc3339"] = time.Now().Format(time.RFC3339)
	dataMap["now_unix"] = time.Now().Unix()
	dataMap["severity"] = logMessage.Severity.String()
	dataMap["facility"] = logMessage.Facility.String()
	dataMap["appname"] = logMessage.AppName
	dataMap["original_message"] = logMessage.Message

	if err = tpl.Execute(&procText, dataMap); err != nil {
		return procText.String(), fmt.Errorf("failed to compile template: %w", err)
	}
	return procText.String(), nil
}

// NewTemplateFuncMap creates a new template function map by returning a
// template.FuncMap.
func NewTemplateFuncMap() template.FuncMap {
	funcMap := FuncMap{}
	return template.FuncMap{
		"_ToLower":  funcMap.ToLower,
		"_ToUpper":  funcMap.ToUpper,
		"_ToBase64": funcMap.ToBase64,
		"_ToSHA1":   funcMap.ToSHA1,
		"_ToSHA256": funcMap.ToSHA256,
		"_ToSHA512": funcMap.ToSHA512,
	}
}

// ToLower returns a given string as lower-case representation
func (*FuncMap) ToLower(value string) string {
	return strings.ToLower(value)
}

// ToUpper returns a given string as upper-case representation
func (*FuncMap) ToUpper(value string) string {
	return strings.ToUpper(value)
}

// ToBase64 returns the base64 encoding of a given string.
func (*FuncMap) ToBase64(value string) string {
	return base64.RawStdEncoding.EncodeToString([]byte(value))
}

// ToSHA1 returns the SHA-1 hash of the given string
func (*FuncMap) ToSHA1(value string) string {
	return toSHA(value, SHA1)
}

// ToSHA256 returns the SHA-256 hash of the given string
func (*FuncMap) ToSHA256(value string) string {
	return toSHA(value, SHA256)
}

// ToSHA512 returns the SHA-512 hash of the given string
func (*FuncMap) ToSHA512(value string) string {
	return toSHA(value, SHA512)
}

// toSHA is a function that converts a string to a SHA hash.
//
// The function takes two parameters: a string 's' and a 'sa' of
// type SHAAlgo which defines the SHA algorithm to be used.
func toSHA(value string, algo SHAAlgo) string {
	var dataHash hash.Hash
	switch algo {
	case SHA1:
		dataHash = sha1.New()
	case SHA256:
		dataHash = sha256.New()
	case SHA512:
		dataHash = sha512.New()
	default:
		return ""
	}

	_, err := io.WriteString(dataHash, value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", dataHash.Sum(nil))
}
