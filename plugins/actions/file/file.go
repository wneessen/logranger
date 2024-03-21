// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package file

import (
	"fmt"
	"os"

	"github.com/wneessen/go-parsesyslog"

	"github.com/wneessen/logranger/plugins/actions"
	"github.com/wneessen/logranger/template"
)

// File represents a file action that can be performed on a log message.
type File struct {
	Enabled        bool
	FilePath       string
	OutputTemplate string
	Overwrite      bool
}

// Config satisfies the plugins.Action interface for the File type
// It updates the configuration of the File action based on the provided
// configuration map.
//
// It expects the configuration map to have a key "file" which contains a submap
// with the following keys:
//   - "output_filepath" (string): Specifies the file path where the output will be written.
//   - "output_template" (string): Specifies the template to use for formatting the output.
//   - "overwrite" (bool, optional): If true, the file will be overwritten instead of appended to.
//
// If any of the required configuration parameters are missing or invalid, an error
// is returned.
func (f *File) Config(configMap map[string]any) error {
	if configMap["file"] == nil {
		return nil
	}
	config, ok := configMap["file"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing configuration for file action")
	}
	f.Enabled = true

	filePath, ok := config["output_filepath"].(string)
	if !ok || filePath == "" {
		return fmt.Errorf("no output_filename configured for file action")
	}
	f.FilePath = filePath

	outputTpl, ok := config["output_template"].(string)
	if !ok || outputTpl == "" {
		return fmt.Errorf("not output_template configured for file action")
	}
	f.OutputTemplate = outputTpl

	if hasOverwrite, ok := config["overwrite"].(bool); ok && hasOverwrite {
		f.Overwrite = true
	}

	return nil
}

// Process satisfies the plugins.Action interface for the File type
// It takes in the log message (lm), match groups (mg), and configuration map (cm).
func (f *File) Process(logMessage parsesyslog.LogMsg, matchGroup []string) error {
	if !f.Enabled {
		return nil
	}

	openFlags := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	if f.Overwrite {
		openFlags = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	}

	fileHandle, err := os.OpenFile(f.FilePath, openFlags, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open file for writing in file action: %w", err)
	}
	defer func() {
		_ = fileHandle.Close()
	}()

	tpl, err := template.Compile(logMessage, matchGroup, f.OutputTemplate)
	if err != nil {
		return err
	}
	_, err = fileHandle.WriteString(tpl)
	if err != nil {
		return fmt.Errorf("failed to write log message to file %q: %w",
			f.FilePath, err)
	}
	if err = fileHandle.Sync(); err != nil {
		return fmt.Errorf("failed to sync memory to file %q: %w",
			f.FilePath, err)
	}

	return nil
}

// init registers the "file" action with the Actions map.
func init() {
	actions.Add("file", &File{})
}
