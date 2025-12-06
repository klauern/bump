package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// VersionFileUpdater handles parsing, updating, and writing Go files
// that contain version constants. This struct isolates file operations
// from git operations for better testability.
type VersionFileUpdater struct{}

// NewVersionFileUpdater creates a new VersionFileUpdater instance.
func NewVersionFileUpdater() *VersionFileUpdater {
	return &VersionFileUpdater{}
}

// ParseGoFile parses a Go source file and returns its AST representation.
// This function is pure file I/O - no git operations.
func (u *VersionFileUpdater) ParseGoFile(filePath string) (*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse file: %w", err)
	}
	return node, fset, nil
}

// UpdateVersionConstant finds and updates the "Version" constant in an AST.
// It searches for a const declaration with a "Version" identifier and updates
// its value to the provided newVersion string.
// Returns an error if the Version constant is not found.
func (u *VersionFileUpdater) UpdateVersionConstant(node *ast.File, newVersion string) error {
	updated := false

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for const declarations
		if gen, ok := n.(*ast.GenDecl); ok && gen.Tok == token.CONST {
			for _, spec := range gen.Specs {
				if value, ok := spec.(*ast.ValueSpec); ok {
					// Check each identifier in the const declaration
					for i, ident := range value.Names {
						if ident.Name == "Version" {
							// Update the value with the new version
							value.Values[i] = &ast.BasicLit{
								Kind:  token.STRING,
								Value: fmt.Sprintf(`"%s"`, newVersion),
							}
							updated = true
							return false // Stop searching
						}
					}
				}
			}
		}
		return true // Continue searching
	})

	if !updated {
		return fmt.Errorf("version constant not found in file")
	}

	return nil
}

// WriteFormattedFile formats an AST and writes it back to a file.
// The file is written with standard Go formatting applied.
func (u *VersionFileUpdater) WriteFormattedFile(filePath string, fset *token.FileSet, node *ast.File) error {
	var buf strings.Builder
	if err := format.Node(&buf, fset, node); err != nil {
		return fmt.Errorf("failed to format AST: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(buf.String()), 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// UpdateVersionInFile is a convenience method that combines ParseGoFile,
// UpdateVersionConstant, and WriteFormattedFile into a single operation.
// This is useful for simple use cases where you just want to update a version.
func (u *VersionFileUpdater) UpdateVersionInFile(filePath, newVersion string) error {
	node, fset, err := u.ParseGoFile(filePath)
	if err != nil {
		return err
	}

	if err := u.UpdateVersionConstant(node, newVersion); err != nil {
		return err
	}

	if err := u.WriteFormattedFile(filePath, fset, node); err != nil {
		return err
	}

	return nil
}
