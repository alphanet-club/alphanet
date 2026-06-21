package rules

import (
	"fmt"
	"os"

	"github.com/alphanet/rules-compiler/internal/air"
)

// LoadRules reads and unmarshals rules.json from the given directory.
// Returns an empty slice if the file does not exist.
func LoadRules(dir string) ([]air.Rule, []byte, error) {
	data, err := os.ReadFile(dir + "/rules.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("reading rules.json: %w", err)
	}

	rf := &air.RulesFile{}
	if err := rf.UnmarshalJSON(data); err != nil {
		return nil, nil, fmt.Errorf("parsing rules.json: %w", err)
	}

	if rf.Rules == nil {
		rf.Rules = []air.Rule{}
	}

	return rf.Rules, data, nil
}
