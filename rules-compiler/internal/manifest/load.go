package manifest

import (
	"fmt"
	"os"

	"github.com/alphanet/rules-compiler/internal/air"
)

// LoadManifest reads and unmarshals manifest.json from the given directory.
func LoadManifest(dir string) (*air.Manifest, error) {
	data, err := os.ReadFile(dir + "/manifest.json")
	if err != nil {
		return nil, fmt.Errorf("reading manifest.json: %w", err)
	}

	m := &air.Manifest{}
	if err := m.UnmarshalJSON(data); err != nil {
		return nil, fmt.Errorf("parsing manifest.json: %w", err)
	}

	if err := ValidateManifest(m); err != nil {
		return nil, fmt.Errorf("validating manifest.json: %w", err)
	}

	return m, nil
}
