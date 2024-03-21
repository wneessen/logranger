// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package logranger

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kkyr/fig"
)

// Ruleset represents a collection of rules.
type Ruleset struct {
	Rule []Rule `fig:"rule"`
}

// Rule represents a rule with its properties.
type Rule struct {
	ID        string         `fig:"id" validate:"required"`
	Regexp    *regexp.Regexp `fig:"regexp" validate:"required"`
	HostMatch *regexp.Regexp `fig:"host_match"`
	Actions   map[string]any `fig:"actions"`
}

// NewRuleset initializes a new Ruleset based on the provided Config.
// It reads the rule file specified in the Config, validates the file's
// existence, and loads the Ruleset using the fig library.
// It checks for duplicate rules and returns an error if any duplicates are found.
// If all operations are successful, it returns the created Ruleset and no error.
func NewRuleset(config *Config) (*Ruleset, error) {
	ruleset := &Ruleset{}
	path := filepath.Dir(config.Server.RuleFile)
	file := filepath.Base(config.Server.RuleFile)
	_, err := os.Stat(fmt.Sprintf("%s/%s", path, file))
	if err != nil {
		return ruleset, fmt.Errorf("failed to read config: %w", err)
	}

	if err = fig.Load(ruleset, fig.Dirs(path), fig.File(file), fig.UseStrict()); err != nil {
		return ruleset, fmt.Errorf("failed to load ruleset: %w", err)
	}

	rules := make([]string, 0)
	for _, rule := range ruleset.Rule {
		for _, rulename := range rules {
			if strings.EqualFold(rule.ID, rulename) {
				return nil, fmt.Errorf("duplicate rule found: %s", rule.ID)
			}
		}
		rules = append(rules, rule.ID)
	}

	return ruleset, nil
}
