package strategy

import (
	"fmt"
	"os"
)

// LoadStrategyMD reads strategy.md from the given directory.
// Returns an empty string if the file does not exist (not an error).
func LoadStrategyMD(dir string) (string, error) {
	data, err := os.ReadFile(dir + "/strategy.md")
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("reading strategy.md: %w", err)
	}
	return string(data), nil
}
