package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFindGitRoot tests the findGitRoot function
func TestFindGitRoot(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create .git directory in temp root
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "a", "b", "c")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	tests := []struct {
		name        string
		startPath   string
		expectError bool
		expectedRoot string
	}{
		{
			name:         "Find from root",
			startPath:    tempDir,
			expectError:  false,
			expectedRoot: tempDir,
		},
		{
			name:         "Find from nested directory",
			startPath:    nestedDir,
			expectError:  false,
			expectedRoot: tempDir,
		},
		{
			name:        "No git repo",
			startPath:   "/",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := findGitRoot(tt.startPath)
			if (err != nil) != tt.expectError {
				t.Errorf("findGitRoot() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && root != tt.expectedRoot {
				t.Errorf("findGitRoot() = %v, expected %v", root, tt.expectedRoot)
			}
		})
	}
}

// newTempGitRepo creates a temporary git repository for testing
func newTempGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	return dir
}

// TestValidateFilePath tests the validateFilePath function for security
func TestValidateFilePath(t *testing.T) {
	tempDir := newTempGitRepo(t)

	// Create test files so symlink resolution works
	versionFile := filepath.Join(tempDir, "version.go")
	if err := os.WriteFile(versionFile, []byte("package main\nconst Version = \"1.0.0\""), 0o644); err != nil {
		t.Fatalf("failed to create version.go: %v", err)
	}

	pkgDir := filepath.Join(tempDir, "pkg", "version")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("failed to create pkg/version dir: %v", err)
	}
	nestedFile := filepath.Join(pkgDir, "version.go")
	if err := os.WriteFile(nestedFile, []byte("package version\nconst Version = \"1.0.0\""), 0o644); err != nil {
		t.Fatalf("failed to create nested version.go: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		repoPath    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid relative path",
			filePath:    "version.go",
			repoPath:    tempDir,
			expectError: false,
		},
		{
			name:        "Valid nested path",
			filePath:    "pkg/version/version.go",
			repoPath:    tempDir,
			expectError: false,
		},
		{
			name:        "Empty path",
			filePath:    "",
			repoPath:    tempDir,
			expectError: true,
			errorMsg:    "empty",
		},
		{
			name:        "Whitespace-only path",
			filePath:    "   ",
			repoPath:    tempDir,
			expectError: true,
			errorMsg:    "empty",
		},
		{
			name:        "Path traversal with ..",
			filePath:    "../../../etc/passwd",
			repoPath:    tempDir,
			expectError: true,
			errorMsg:    "invalid",
		},
		{
			name:        "Absolute path",
			filePath:    "/etc/passwd",
			repoPath:    tempDir,
			expectError: true,
			errorMsg:    "absolute",
		},
		{
			name:        "Null byte injection",
			filePath:    "version\x00.go",
			repoPath:    tempDir,
			expectError: true,
			errorMsg:    "invalid characters",
		},
		{
			name:        "Newline injection",
			filePath:    "version\n.go",
			repoPath:    tempDir,
			expectError: true,
			errorMsg:    "invalid characters",
		},
		{
			name:        "Carriage return injection",
			filePath:    "version\r.go",
			repoPath:    tempDir,
			expectError: true,
			errorMsg:    "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFilePath(tt.filePath, tt.repoPath)
			if (err != nil) != tt.expectError {
				t.Errorf("validateFilePath() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestValidateFilePathBoundaryChecks tests edge cases for boundary validation
func TestValidateFilePathBoundaryChecks(t *testing.T) {
	tempDir := newTempGitRepo(t)

	// Create a subdirectory to test boundary checks
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// Create test file
	testFile := filepath.Join(subDir, "file.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "Path within repo",
			filePath:    "subdir/file.go",
			expectError: false,
		},
		{
			name:        "Attempt to escape with ../",
			filePath:    "subdir/../../outside.go",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFilePath(tt.filePath, tempDir)
			if (err != nil) != tt.expectError {
				t.Errorf("validateFilePath() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestUpdateVersionFileValidation tests updateVersionFile input validation
func TestUpdateVersionFileValidation(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		nextTag     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Empty file path",
			filePath:    "",
			nextTag:     "v1.0.0",
			expectError: true,
			errorMsg:    "empty",
		},
		{
			name:        "Path traversal attempt",
			filePath:    "../../../etc/passwd",
			nextTag:     "v1.0.0",
			expectError: true,
			errorMsg:    "invalid",
		},
		{
			name:        "Invalid tag format",
			filePath:    "version.go",
			nextTag:     "invalid",
			expectError: true,
			errorMsg:    "parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := updateVersionFile(tt.filePath, tt.nextTag)
			if (err != nil) != tt.expectError {
				t.Errorf("updateVersionFile() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestBumpVersionNoRepo tests bumpVersion when not in a git repository
func TestBumpVersionNoRepo(t *testing.T) {
	// Create a temp directory WITHOUT .git
	tempDir := t.TempDir()

	// Change to temp directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	err = bumpVersion("patch", "", "", false, false)
	if err == nil {
		t.Error("bumpVersion should error when not in a git repository")
	}
}

// TestCreateCommandStructure tests that createCommand returns proper command structure
func TestCreateCommandStructure(t *testing.T) {
	cmd := createCommand("patch", "p", "Test usage")

	if cmd.Name != "patch" {
		t.Errorf("expected name 'patch', got '%s'", cmd.Name)
	}

	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "p" {
		t.Errorf("expected alias 'p', got %v", cmd.Aliases)
	}

	if cmd.Usage != "Test usage" {
		t.Errorf("expected usage 'Test usage', got '%s'", cmd.Usage)
	}

	// Check flags exist
	flagNames := []string{"suffix", "update-file", "push", "dry-run"}
	for _, flagName := range flagNames {
		found := false
		for _, flag := range cmd.Flags {
			if flag.Names()[0] == flagName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected flag '%s' not found", flagName)
		}
	}

	if cmd.Action == nil {
		t.Error("expected Action to be set")
	}
}
