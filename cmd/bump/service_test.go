package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewBumpService tests the service constructor
func TestNewBumpService(t *testing.T) {
	tests := []struct {
		name    string
		repo    GitRepository
		updater *VersionFileUpdater
		output  *bytes.Buffer
	}{
		{
			name:    "All dependencies provided",
			repo:    NewMockRepoWithTags([]string{}),
			updater: NewVersionFileUpdater(),
			output:  &bytes.Buffer{},
		},
		{
			name:    "Nil output (uses stdout)",
			repo:    NewMockRepoWithTags([]string{}),
			updater: NewVersionFileUpdater(),
			output:  nil,
		},
		{
			name:    "Nil updater (creates new)",
			repo:    NewMockRepoWithTags([]string{}),
			updater: nil,
			output:  &bytes.Buffer{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewBumpService(tt.repo, tt.updater, tt.output)
			if svc == nil {
				t.Error("NewBumpService() should not return nil")
				return
			}
			if svc.repo == nil {
				t.Error("Service repo should not be nil")
			}
			if svc.updater == nil {
				t.Error("Service updater should not be nil")
			}
			if svc.output == nil {
				t.Error("Service output should not be nil")
			}
		})
	}
}

// TestBump_Success tests successful bump operations
func TestBump_Success(t *testing.T) {
	tests := []struct {
		name         string
		existingTags []string
		opts         BumpOptions
		expectedTag  string
		expectedPush bool
	}{
		{
			name:         "Patch bump from v1.0.0",
			existingTags: []string{"v1.0.0"},
			opts: BumpOptions{
				BumpType: "patch",
				Push:     false,
			},
			expectedTag:  "v1.0.1",
			expectedPush: false,
		},
		{
			name:         "Minor bump from v1.0.0",
			existingTags: []string{"v1.0.0"},
			opts: BumpOptions{
				BumpType: "minor",
				Push:     false,
			},
			expectedTag:  "v1.1.0",
			expectedPush: false,
		},
		{
			name:         "Major bump from v1.2.3",
			existingTags: []string{"v1.2.3"},
			opts: BumpOptions{
				BumpType: "major",
				Push:     false,
			},
			expectedTag:  "v2.0.0",
			expectedPush: false,
		},
		{
			name:         "Bump with push",
			existingTags: []string{"v0.1.0"},
			opts: BumpOptions{
				BumpType: "patch",
				Push:     true,
			},
			expectedTag:  "v0.1.1",
			expectedPush: true,
		},
		{
			name:         "Bump with suffix",
			existingTags: []string{"v2.0.0"},
			opts: BumpOptions{
				BumpType: "patch",
				Suffix:   "beta",
				Push:     false,
			},
			expectedTag:  "v2.0.1-beta",
			expectedPush: false,
		},
		{
			name:         "First tag (no existing tags)",
			existingTags: []string{},
			opts: BumpOptions{
				BumpType: "patch",
				Push:     false,
			},
			expectedTag:  "v0.1.0",
			expectedPush: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			output := &bytes.Buffer{}
			repo := NewMockRepoWithTags(tt.existingTags)

			svc := NewBumpService(repo, nil, output)

			// Execute
			result, err := svc.Bump(tt.opts)

			// Verify
			if err != nil {
				t.Errorf("Bump() unexpected error = %v", err)
				return
			}

			if result.NextTag != tt.expectedTag {
				t.Errorf("NextTag = %v, expected %v", result.NextTag, tt.expectedTag)
			}

			if result.Pushed != tt.expectedPush {
				t.Errorf("Pushed = %v, expected %v", result.Pushed, tt.expectedPush)
			}

			// Verify output contains success message
			outputStr := output.String()
			if !strings.Contains(outputStr, tt.expectedTag) {
				t.Errorf("Output should contain tag %v, got: %v", tt.expectedTag, outputStr)
			}
		})
	}
}

// TestBump_DryRun tests dry-run mode
func TestBump_DryRun(t *testing.T) {
	tests := []struct {
		name         string
		existingTags []string
		opts         BumpOptions
		expectedTag  string
		expectOutput []string
	}{
		{
			name:         "Dry run patch bump",
			existingTags: []string{"v1.0.0"},
			opts: BumpOptions{
				BumpType: "patch",
				DryRun:   true,
			},
			expectedTag: "v1.0.1",
			expectOutput: []string{
				"Would create tag: v1.0.1",
			},
		},
		{
			name:         "Dry run with push",
			existingTags: []string{"v1.0.0"},
			opts: BumpOptions{
				BumpType: "patch",
				Push:     true,
				DryRun:   true,
			},
			expectedTag: "v1.0.1",
			expectOutput: []string{
				"Would create tag: v1.0.1",
				"Would push tag to remote",
			},
		},
		{
			name:         "Dry run with file update",
			existingTags: []string{"v1.0.0"},
			opts: BumpOptions{
				BumpType:   "minor",
				UpdateFile: "version.go",
				DryRun:     true,
			},
			expectedTag: "v1.1.0",
			expectOutput: []string{
				"Would create tag: v1.1.0",
				"Would update file: version.go",
			},
		},
		{
			name:         "Dry run first tag",
			existingTags: []string{},
			opts: BumpOptions{
				BumpType: "patch",
				DryRun:   true,
			},
			expectedTag: "v0.1.0",
			expectOutput: []string{
				"No tags found, would start at v0.1.0",
				"Would create tag: v0.1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			repo := NewMockRepoWithTags(tt.existingTags)
			svc := NewBumpService(repo, nil, output)

			result, err := svc.Bump(tt.opts)

			if err != nil {
				t.Errorf("Bump() unexpected error = %v", err)
				return
			}

			if result.NextTag != tt.expectedTag {
				t.Errorf("NextTag = %v, expected %v", result.NextTag, tt.expectedTag)
			}

			if result.WouldPush != tt.opts.Push {
				t.Errorf("WouldPush = %v, expected %v", result.WouldPush, tt.opts.Push)
			}

			if result.WouldUpdate != (tt.opts.UpdateFile != "") {
				t.Errorf("WouldUpdate = %v, expected %v", result.WouldUpdate, tt.opts.UpdateFile != "")
			}

			// Verify dry-run doesn't actually create tags
			if result.Pushed {
				t.Error("Dry-run should not actually push tags")
			}
			if result.FileUpdated {
				t.Error("Dry-run should not actually update files")
			}

			// Verify output
			outputStr := output.String()
			for _, expected := range tt.expectOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Output missing expected string: %v\nGot: %v", expected, outputStr)
				}
			}
		})
	}
}

// TestBump_Errors tests error handling
func TestBump_Errors(t *testing.T) {
	tests := []struct {
		name        string
		repo        GitRepository
		opts        BumpOptions
		expectError string
	}{
		{
			name:        "Tags() error",
			repo:        NewMockRepoWithError(fmt.Errorf("tags failed"), nil, nil),
			opts:        BumpOptions{BumpType: "patch"},
			expectError: "failed to fetch tags",
		},
		{
			name:        "CreateTag() error",
			repo:        NewMockRepoWithError(nil, fmt.Errorf("create failed"), nil),
			opts:        BumpOptions{BumpType: "patch"},
			expectError: "failed to create tag",
		},
		{
			name:        "PushTags() error",
			repo:        NewMockRepoWithError(nil, nil, fmt.Errorf("push failed")),
			opts:        BumpOptions{BumpType: "patch", Push: true},
			expectError: "failed to push tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			svc := NewBumpService(tt.repo, nil, output)

			_, err := svc.Bump(tt.opts)

			if err == nil {
				t.Error("Bump() should return error")
				return
			}

			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Error should contain %q, got: %v", tt.expectError, err)
			}
		})
	}
}

// TestUpdateVersionFile_Success tests successful file updates
func TestUpdateVersionFile_Success(t *testing.T) {
	// Create temp directory (this will be the repo root)
	tmpDir := t.TempDir()

	// Create a simple version file in the repo root
	versionFile := filepath.Join(tmpDir, "version.go")
	initialContent := `package main

const Version = "1.0.0"
`
	if err := os.WriteFile(versionFile, []byte(initialContent), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Setup mock repo
	repo := &MockGitRepository{
		PathFunc: func() string { return tmpDir },
		WorktreeFunc: func() (GitWorktree, error) {
			return &MockGitWorktree{}, nil
		},
	}

	// Create service
	svc := NewBumpService(repo, nil, &bytes.Buffer{})

	// Execute with relative path (validateFilePath requires relative paths)
	err := svc.UpdateVersionFile("version.go", "v1.0.1")
	if err != nil {
		t.Errorf("UpdateVersionFile() unexpected error = %v", err)
		return
	}

	// Verify file was updated
	updated, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	updatedStr := string(updated)
	expectedVersion := "1.0.2-dev" // Next dev version after 1.0.1
	if !strings.Contains(updatedStr, expectedVersion) {
		t.Errorf("Updated file should contain %q, got: %v", expectedVersion, updatedStr)
	}
}

// TestUpdateVersionFile_Errors tests error handling
func TestUpdateVersionFile_Errors(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (filePath, nextTag string, repo GitRepository)
		expectError string
	}{
		{
			name: "Empty file path",
			setup: func(t *testing.T) (string, string, GitRepository) {
				return "", "v1.0.0", &MockGitRepository{
					PathFunc: func() string { return "/tmp" },
				}
			},
			expectError: "invalid file path",
		},
		{
			name: "Path traversal attempt",
			setup: func(t *testing.T) (string, string, GitRepository) {
				return "../../../etc/passwd", "v1.0.0", &MockGitRepository{
					PathFunc: func() string { return "/tmp" },
				}
			},
			expectError: "invalid file path",
		},
		{
			name: "Invalid tag format",
			setup: func(t *testing.T) (string, string, GitRepository) {
				// Create a valid file for this test
				tmpDir := t.TempDir()
				versionFile := filepath.Join(tmpDir, "version.go")
				if err := os.WriteFile(versionFile, []byte("package main\nconst Version = \"1.0.0\""), 0o644); err != nil {
					t.Fatalf("failed to write version file: %v", err)
				}

				return "version.go", "invalid", &MockGitRepository{
					PathFunc: func() string { return tmpDir },
				}
			},
			expectError: "failed to calculate dev version",
		},
		{
			name: "Worktree error",
			setup: func(t *testing.T) (string, string, GitRepository) {
				// Create a valid file for this test
				tmpDir := t.TempDir()
				versionFile := filepath.Join(tmpDir, "version.go")
				if err := os.WriteFile(versionFile, []byte("package main\nconst Version = \"1.0.0\""), 0o644); err != nil {
					t.Fatalf("failed to write version file: %v", err)
				}

				return "version.go", "v1.0.0", &MockGitRepository{
					PathFunc: func() string { return tmpDir },
					WorktreeFunc: func() (GitWorktree, error) {
						return nil, fmt.Errorf("worktree failed")
					},
				}
			},
			expectError: "failed to get working tree",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, nextTag, repo := tt.setup(t)
			svc := NewBumpService(repo, nil, &bytes.Buffer{})

			err := svc.UpdateVersionFile(filePath, nextTag)

			if err == nil {
				t.Error("UpdateVersionFile() should return error")
				return
			}

			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Error should contain %q, got: %v", tt.expectError, err)
			}
		})
	}
}
