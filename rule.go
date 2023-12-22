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

type Ruleset struct {
	Rule []struct {
		ID     string         `fig:"id" validate:"required"`
		Regexp *regexp.Regexp `fig:"regexp" validate:"required"`
	} `fig:"rule"`
}

func NewRuleset(c *Config) (*Ruleset, error) {
	rs := &Ruleset{}
	p := filepath.Dir(c.Server.RuleFile)
	f := filepath.Base(c.Server.RuleFile)
	_, err := os.Stat(fmt.Sprintf("%s/%s", p, f))
	if err != nil {
		return rs, fmt.Errorf("failed to read config: %w", err)
	}

	if err = fig.Load(rs, fig.Dirs(p), fig.File(f), fig.UseStrict()); err != nil {
		return rs, fmt.Errorf("failed to load ruleset: %w", err)
	}

	rna := make([]string, 0)
	for _, r := range rs.Rule {
		for _, rn := range rna {
			if strings.EqualFold(r.ID, rn) {
				return nil, fmt.Errorf("duplicate rule found: %s", r.ID)
			}
		}
		rna = append(rna, r.ID)
	}

	return rs, nil
}
