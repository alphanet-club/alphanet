package compiler

// manual.go — Manual mode compilation is handled in compiler.go's compileManual function.
// This file exists as a placeholder for future manual-mode-specific logic.
//
// In the current implementation, compileManual() in compiler.go handles:
//   - Normalizing rules
//   - Normalizing portfolio
//   - Building AIR with default decision hierarchy and execution config
//   - Building provenance with "manual" mode
//   - Computing IR hash
//   - Building reasoning narrative
//   - Building validation report
