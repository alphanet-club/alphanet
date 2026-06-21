package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// SchemaValidator loads and caches JSON schemas for validating strategy files.
type SchemaValidator struct {
	specDir string
	schemas map[string]*jsonschema.Schema
}

// NewSchemaValidator creates a validator that loads schemas from the spec directory.
func NewSchemaValidator(specDir string) *SchemaValidator {
	return &SchemaValidator{
		specDir: specDir,
		schemas: make(map[string]*jsonschema.Schema),
	}
}

// ValidateManifest validates manifest.json bytes against manifest.schema.json.
func (sv *SchemaValidator) ValidateManifest(raw []byte) ([]string, error) {
	schema, err := sv.loadSchema("manifest.schema.json")
	if err != nil {
		return nil, fmt.Errorf("loading manifest schema: %w", err)
	}
	return sv.validateBytes(schema, raw)
}

// ValidateRules validates rules.json bytes against rule.schema.json.
func (sv *SchemaValidator) ValidateRules(raw []byte) ([]string, error) {
	schema, err := sv.loadSchema("rule.schema.json")
	if err != nil {
		return nil, fmt.Errorf("loading rule schema: %w", err)
	}
	return sv.validateBytes(schema, raw)
}

// ValidateAIR validates strategy.ir.json bytes against alphanet.schema.json.
func (sv *SchemaValidator) ValidateAIR(raw []byte) ([]string, error) {
	schema, err := sv.loadSchema("alphanet.schema.json")
	if err != nil {
		return nil, fmt.Errorf("loading alphanet schema: %w", err)
	}
	return sv.validateBytes(schema, raw)
}

// loadSchema loads, compiles, and caches a JSON schema from the spec directory.
func (sv *SchemaValidator) loadSchema(filename string) (*jsonschema.Schema, error) {
	if cached, ok := sv.schemas[filename]; ok {
		return cached, nil
	}

	path := filepath.Join(sv.specDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filename, err)
	}

	// Extract the $id from the schema for resource registration
	var meta struct {
		ID string `json:"$id"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("extracting $id from %s: %w", filename, err)
	}

	schemaURL := meta.ID
	if schemaURL == "" {
		schemaURL = filename
	}

	// Parse the schema JSON so the compiler receives decoded Go values
	var schemaDoc any
	if err := json.Unmarshal(data, &schemaDoc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filename, err)
	}

	// Compile the schema
	c := jsonschema.NewCompiler()

	if err := c.AddResource(schemaURL, schemaDoc); err != nil {
		return nil, fmt.Errorf("adding resource %s: %w", filename, err)
	}

	compiled, err := c.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("compiling %s: %w", filename, err)
	}

	sv.schemas[filename] = compiled
	return compiled, nil
}

// validateBytes validates raw JSON bytes against a compiled schema.
func (sv *SchemaValidator) validateBytes(schema *jsonschema.Schema, raw []byte) ([]string, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	err := schema.Validate(v)
	if err == nil {
		return nil, nil
	}

	var errs []string
	if ve, ok := err.(*jsonschema.ValidationError); ok {
		errs = collectValidationErrors(ve, "")
	} else {
		errs = []string{err.Error()}
	}

	return errs, nil
}

// collectValidationErrors recursively flattens ValidationError into string messages.
func collectValidationErrors(ve *jsonschema.ValidationError, prefix string) []string {
	var result []string
	msg := ve.Error()
	// Only add if there's meaningful content
	if msg != "" {
		if prefix != "" {
			msg = prefix + msg
		}
		result = append(result, msg)
	}
	for _, cause := range ve.Causes {
		result = append(result, collectValidationErrors(cause, prefix+"  ")...)
	}
	return result
}

// SchemaCheck represents a single schema validation result.
type SchemaCheck struct {
	SchemaFile string
	CheckName  string
	Errors     []string
}

// RunAllSchemaValidations runs schema validation for all available source file bytes.
func (sv *SchemaValidator) RunAllSchemaValidations(manifestRaw, rulesRaw, airRaw []byte) []SchemaCheck {
	var checks []SchemaCheck

	if manifestRaw != nil {
		errs, err := sv.ValidateManifest(manifestRaw)
		if err != nil {
			errs = []string{fmt.Sprintf("schema loader error: %v", err)}
		}
		checks = append(checks, SchemaCheck{
			SchemaFile: "manifest.schema.json",
			CheckName:  "manifest_schema",
			Errors:     errs,
		})
	}

	if rulesRaw != nil {
		errs, err := sv.ValidateRules(rulesRaw)
		if err != nil {
			errs = []string{fmt.Sprintf("schema loader error: %v", err)}
		}
		checks = append(checks, SchemaCheck{
			SchemaFile: "rule.schema.json",
			CheckName:  "rules_schema",
			Errors:     errs,
		})
	}

	if airRaw != nil {
		errs, err := sv.ValidateAIR(airRaw)
		if err != nil {
			errs = []string{fmt.Sprintf("schema loader error: %v", err)}
		}
		checks = append(checks, SchemaCheck{
			SchemaFile: "alphanet.schema.json",
			CheckName:  "air_schema",
			Errors:     errs,
		})
	}

	return checks
}
