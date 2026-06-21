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
	if !dryRun {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	if air != nil && len(air.AgentReportContents) > 0 {
		if err := writeAgentReportArtifacts(air.AgentReportContents, outDir, dryRun); err != nil {
			return err
		}
	}

	if err := writeJSON(filepath.Join(outDir, "strategy.ir.json"), air, dryRun); err != nil {
		return fmt.Errorf("writing strategy.ir.json: %w", err)
	}
	if err := writeJSON(filepath.Join(outDir, "provenance.json"), provenance, dryRun); err != nil {
		return fmt.Errorf("writing provenance.json: %w", err)
	}
	if err := writeText(filepath.Join(outDir, "reasoning.md"), reasoning, dryRun); err != nil {
		return fmt.Errorf("writing reasoning.md: %w", err)
	}
	if err := writeJSON(filepath.Join(outDir, "validation-report.json"), validation, dryRun); err != nil {
		return fmt.Errorf("writing validation-report.json: %w", err)
	}
	return nil
}

func writeAgentReportArtifacts(reports []AgentReportArtifact, outDir string, dryRun bool) error {
	for _, report := range reports {
		if report.Ref.Path == "" || report.Content == "" {
			continue
		}
		path := filepath.Join(outDir, filepath.FromSlash(report.Ref.Path))
		if dryRun {
			fmt.Printf("=== %s ===\n", report.Ref.Path)
			fmt.Print(report.Content)
			if len(report.Content) == 0 || report.Content[len(report.Content)-1] != '\n' {
				fmt.Println()
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("creating agent report directory: %w", err)
		}
		if err := os.WriteFile(path, []byte(report.Content), 0644); err != nil {
			return fmt.Errorf("writing agent report %s: %w", report.Ref.Path, err)
		}
	}
	return nil
}

// writeJSON serializes v to pretty-printed JSON without HTML escaping.
func writeJSON(path string, v any, dryRun bool) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	data := buf.Bytes()

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
