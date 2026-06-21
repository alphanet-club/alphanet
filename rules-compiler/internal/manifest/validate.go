package manifest

import (
	"fmt"

	"github.com/alphanet/rules-compiler/internal/air"
)

// ValidateManifest checks required fields and structural validity.
func ValidateManifest(m *air.Manifest) error {
	var errs []string

	if m.Name == "" {
		errs = append(errs, "manifest: name is required")
	}
	if m.SpecVersion == "" {
		errs = append(errs, "manifest: spec_version is required")
	}

	// Validate compiler mode if present
	if m.Compiler.Mode != "" {
		switch m.Compiler.Mode {
		case "none", "manual", "single", "ensemble":
			// valid
		default:
			errs = append(errs, fmt.Sprintf("manifest: unsupported compiler mode %q", m.Compiler.Mode))
		}
	}

	// Validate engines if present
	for i, e := range m.Compiler.Engines {
		if e.Name == "" {
			errs = append(errs, fmt.Sprintf("manifest: compiler.engines[%d]: name is required", i))
		}
		if e.Version == "" {
			errs = append(errs, fmt.Sprintf("manifest: compiler.engines[%d]: version is required", i))
		}
	}

	// Validate training window
	tw := m.Compiler.TrainingWindow
	if tw != nil {
		if tw.LookbackDays < 0 && tw.Start == "" && tw.End == "" {
			errs = append(errs, "manifest: training_window must specify lookback_days or start/end")
		}
		if tw.Start != "" && tw.End != "" && tw.Start > tw.End {
			errs = append(errs, "manifest: training_window.start must be before training_window.end")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", joinErrors(errs))
	}
	return nil
}

func joinErrors(errs []string) string {
	result := ""
	for i, e := range errs {
		if i > 0 {
			result += "; "
		}
		result += e
	}
	return result
}
