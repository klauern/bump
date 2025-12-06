package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewVersionFileUpdater tests the constructor
func TestNewVersionFileUpdater(t *testing.T) {
	updater := NewVersionFileUpdater()
	if updater == nil {
		t.Error("NewVersionFileUpdater() should not return nil")
	}
}

// TestParseGoFile tests parsing valid and invalid Go files
func TestParseGoFile(t *testing.T) {
	updater := NewVersionFileUpdater()

	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name: "Valid Go file with Version constant",
			content: `package main

const Version = "1.0.0"
`,
			expectError: false,
		},
		{
			name: "Valid Go file with multiple constants",
			content: `package main

const (
	AppName = "test"
	Version = "1.0.0"
	BuildDate = "2023-01-01"
)
`,
			expectError: false,
		},
		{
			name: "Valid Go file without Version constant",
			content: `package main

func main() {
	println("Hello")
}
`,
			expectError: false,
		},
		{
			name:        "Invalid Go syntax",
			content:     `package main\n\nconst Version = `,
			expectError: true,
		},
		{
			name:        "Empty file",
			content:     ``,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile := filepath.Join(t.TempDir(), "test.go")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			node, fset, err := updater.ParseGoFile(tmpFile)

			if (err != nil) != tt.expectError {
				t.Errorf("ParseGoFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				if node == nil {
					t.Error("ParseGoFile() returned nil node for valid file")
				}
				if fset == nil {
					t.Error("ParseGoFile() returned nil FileSet for valid file")
				}
			}
		})
	}
}

// TestParseGoFile_NonexistentFile tests error handling for missing files
func TestParseGoFile_NonexistentFile(t *testing.T) {
	updater := NewVersionFileUpdater()

	_, _, err := updater.ParseGoFile("/nonexistent/file/path.go")
	if err == nil {
		t.Error("ParseGoFile() should error on nonexistent file")
	}
}

// TestUpdateVersionConstant tests AST manipulation
func TestUpdateVersionConstant(t *testing.T) {
	updater := NewVersionFileUpdater()

	tests := []struct {
		name         string
		content      string
		newVersion   string
		expectError  bool
		expectedCode string
	}{
		{
			name: "Update simple Version constant",
			content: `package main

const Version = "1.0.0"
`,
			newVersion:  "2.0.0",
			expectError: false,
			expectedCode: `package main

const Version = "2.0.0"
`,
		},
		{
			name: "Update Version in const block",
			content: `package main

const (
	AppName = "test"
	Version = "1.0.0"
	BuildDate = "2023-01-01"
)
`,
			newVersion:  "3.5.2-beta",
			expectError: false,
			expectedCode: `package main

const (
	AppName   = "test"
	Version   = "3.5.2-beta"
	BuildDate = "2023-01-01"
)
`,
		},
		{
			name: "No Version constant found",
			content: `package main

const AppName = "test"
`,
			newVersion:  "1.0.0",
			expectError: true,
		},
		{
			name: "Empty file (no declarations)",
			content: `package main
`,
			newVersion:  "1.0.0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the source code
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test.go", tt.content, parser.ParseComments)
			if err != nil {
				t.Fatalf("failed to parse test fixture: %v", err)
			}

			// Update the version constant
			err = updater.UpdateVersionConstant(node, tt.newVersion)

			if (err != nil) != tt.expectError {
				t.Errorf("UpdateVersionConstant() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// If successful, verify the update
			if !tt.expectError {
				// Find the Version constant and check its value
				found := false
				var actualValue string

				ast.Inspect(node, func(n ast.Node) bool {
					if gen, ok := n.(*ast.GenDecl); ok && gen.Tok == token.CONST {
						for _, spec := range gen.Specs {
							if value, ok := spec.(*ast.ValueSpec); ok {
								for i, ident := range value.Names {
									if ident.Name == "Version" {
										if lit, ok := value.Values[i].(*ast.BasicLit); ok {
											actualValue = strings.Trim(lit.Value, `"`)
											found = true
										}
									}
								}
							}
						}
					}
					return true
				})

				if !found {
					t.Error("Version constant not found after update")
				}

				if actualValue != tt.newVersion {
					t.Errorf("Version = %v, expected %v", actualValue, tt.newVersion)
				}
			}
		})
	}
}

// TestWriteFormattedFile tests writing AST back to file
func TestWriteFormattedFile(t *testing.T) {
	updater := NewVersionFileUpdater()

	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name: "Write valid AST",
			content: `package main

const Version = "1.0.0"
`,
			expectError: false,
		},
		{
			name: "Write complex AST",
			content: `package main

import "fmt"

const (
	Version = "2.0.0"
	AppName = "test"
)

func main() {
	fmt.Println(Version)
}
`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the content
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test.go", tt.content, parser.ParseComments)
			if err != nil {
				t.Fatalf("failed to parse test fixture: %v", err)
			}

			// Write to temp file
			tmpFile := filepath.Join(t.TempDir(), "output.go")
			err = updater.WriteFormattedFile(tmpFile, fset, node)

			if (err != nil) != tt.expectError {
				t.Errorf("WriteFormattedFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Read back and verify it's valid Go code
				written, err := os.ReadFile(tmpFile)
				if err != nil {
					t.Fatalf("failed to read written file: %v", err)
				}

				// Parse the written file to ensure it's valid Go
				_, err = parser.ParseFile(token.NewFileSet(), tmpFile, written, 0)
				if err != nil {
					t.Errorf("written file is not valid Go code: %v", err)
				}
			}
		})
	}
}

// TestWriteFormattedFile_InvalidPath tests error handling for bad paths
func TestWriteFormattedFile_InvalidPath(t *testing.T) {
	updater := NewVersionFileUpdater()

	// Parse some valid content
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", "package main\nconst Version = \"1.0.0\"", 0)
	if err != nil {
		t.Fatalf("failed to parse test fixture: %v", err)
	}

	// Try to write to an invalid path
	err = updater.WriteFormattedFile("/invalid/path/that/doesnt/exist/file.go", fset, node)
	if err == nil {
		t.Error("WriteFormattedFile() should error on invalid path")
	}
}

// TestUpdateVersionInFile tests the convenience method end-to-end
func TestUpdateVersionInFile(t *testing.T) {
	updater := NewVersionFileUpdater()

	tests := []struct {
		name        string
		content     string
		newVersion  string
		expectError bool
	}{
		{
			name: "Successful update",
			content: `package main

const Version = "1.0.0"
`,
			newVersion:  "2.5.3",
			expectError: false,
		},
		{
			name: "Update in const block",
			content: `package main

const (
	AppName = "myapp"
	Version = "0.1.0"
)
`,
			newVersion:  "1.0.0-dev",
			expectError: false,
		},
		{
			name: "No Version constant",
			content: `package main

const AppName = "test"
`,
			newVersion:  "1.0.0",
			expectError: true,
		},
		{
			name:        "Invalid Go syntax",
			content:     `package main\nconst Version = `,
			newVersion:  "1.0.0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile := filepath.Join(t.TempDir(), "version.go")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			// Update version
			err := updater.UpdateVersionInFile(tmpFile, tt.newVersion)

			if (err != nil) != tt.expectError {
				t.Errorf("UpdateVersionInFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Read back the file
				updated, err := os.ReadFile(tmpFile)
				if err != nil {
					t.Fatalf("failed to read updated file: %v", err)
				}

				// Parse and verify the version was updated
				fset := token.NewFileSet()
				node, err := parser.ParseFile(fset, tmpFile, updated, 0)
				if err != nil {
					t.Fatalf("updated file is not valid Go: %v", err)
				}

				// Find and check the Version constant
				found := false
				var actualValue string

				ast.Inspect(node, func(n ast.Node) bool {
					if gen, ok := n.(*ast.GenDecl); ok && gen.Tok == token.CONST {
						for _, spec := range gen.Specs {
							if value, ok := spec.(*ast.ValueSpec); ok {
								for i, ident := range value.Names {
									if ident.Name == "Version" {
										if lit, ok := value.Values[i].(*ast.BasicLit); ok {
											actualValue = strings.Trim(lit.Value, `"`)
											found = true
										}
									}
								}
							}
						}
					}
					return true
				})

				if !found {
					t.Error("Version constant not found in updated file")
				}

				if actualValue != tt.newVersion {
					t.Errorf("Version = %v, expected %v", actualValue, tt.newVersion)
				}
			}
		})
	}
}

// TestUpdateVersionInFile_NonexistentFile tests error handling
func TestUpdateVersionInFile_NonexistentFile(t *testing.T) {
	updater := NewVersionFileUpdater()

	err := updater.UpdateVersionInFile("/nonexistent/file.go", "1.0.0")
	if err == nil {
		t.Error("UpdateVersionInFile() should error on nonexistent file")
	}
}
