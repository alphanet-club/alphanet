package air

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteOutputs writes all compiled output files to the given output directory.
func WriteOutputs(air *AIR, provenance *Provenance, reasoning string, validation *ValidationReport, outDir string, dryRun bool) error {
	// Create output directory
	if !dryRun {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	// 1. Write strategy.ir.json
	if err := writeJSON(filepath.Join(outDir, "strategy.ir.json"), air, dryRun); err != nil {
		return fmt.Errorf("writing strategy.ir.json: %w", err)
	}

	// 2. Write provenance.json
	if err := writeJSON(filepath.Join(outDir, "provenance.json"), provenance, dryRun); err != nil {
		return fmt.Errorf("writing provenance.json: %w", err)
	}

	// 3. Write reasoning.md
	if err := writeText(filepath.Join(outDir, "reasoning.md"), reasoning, dryRun); err != nil {
		return fmt.Errorf("writing reasoning.md: %w", err)
	}

	// 4. Write validation-report.json
	if err := writeJSON(filepath.Join(outDir, "validation-report.json"), validation, dryRun); err != nil {
		return fmt.Errorf("writing validation-report.json: %w", err)
	}

	return nil
}

// writeJSON serializes v to pretty-printed JSON without HTML escaping.
func writeJSON(path string, v any, dryRun bool) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	data := buf.Bytes()

	// json.Encoder adds a trailing newline; that's fine for files
	if dryRun {
		fmt.Printf("=== %s ===\n", filepath.Base(path))
		fmt.Print(string(data))
		if len(data) > 0 && data[len(data)-1] != '\n' {
			fmt.Println()
		}
		return nil
	}

	return os.WriteFile(path, data, 0644)
}

func writeText(path string, content string, dryRun bool) error {
	if dryRun {
		fmt.Printf("=== %s ===\n", filepath.Base(path))
		fmt.Println(content)
		fmt.Println()
		return nil
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// ValidationReport holds the results of schema and semantic validation.
type ValidationReport struct {
	Status      string              `json:"status"`
	SpecVersion string              `json:"spec_version,omitempty"`
	ValidatedAt string              `json:"validated_at,omitempty"`
	Schemas     []string            `json:"schemas"`
	Checks      []ValidationCheck   `json:"checks"`
	Warnings    []ValidationWarning `json:"warnings,omitempty"`
	Errors      []string            `json:"errors"`
}

// ValidationCheck is a single validation check result.
type ValidationCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// ValidationWarning is a structured warning with a machine-readable code.
type ValidationWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
