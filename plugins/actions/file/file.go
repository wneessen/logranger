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
type File struct{}

// Process satisfies the plugins.Action interface for the File type
// It takes in the log message (lm), match groups (mg), and configuration map (cm).
func (f *File) Process(lm parsesyslog.LogMsg, mg []string, cm map[string]any) error {
	if cm["file"] == nil {
		return nil
	}
	c, ok := cm["file"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing configuration for file action")
	}
	ot, ok := c["output_template"].(string)
	if !ok || ot == "" {
		return fmt.Errorf("not output_template configured for file action")
	}

	fn, ok := c["output_filepath"].(string)
	if !ok || fn == "" {
		return fmt.Errorf("no output_filename configured for file action")
	}

	of := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	if ow, ok := c["overwrite"].(bool); ok && ow {
		of = os.O_WRONLY | os.O_CREATE
	}

	fh, err := os.OpenFile(fn, of, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open file for writing in file action: %w", err)
	}
	defer func() {
		_ = fh.Close()
	}()

	t, err := template.Compile(lm, mg, ot)
	if err != nil {
		return err
	}
	_, err = fh.WriteString(t)
	if err != nil {
		return fmt.Errorf("failed to write log message to file %q: %w", fn, err)
	}
	if err := fh.Sync(); err != nil {
		return fmt.Errorf("failed to sync memory to file %q: %w", fn, err)
	}

	return nil
}

// init registers the "file" action with the Actions map.
func init() {
	actions.Add("file", &File{})
}
