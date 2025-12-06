package main

import (
	"fmt"
	"strings"
	"testing"
)

// TestCalculateNextVersion tests the pure function for calculating next version
func TestCalculateNextVersion(t *testing.T) {
	tests := []struct {
		name        string
		latestTag   string
		bumpType    string
		suffix      string
		expected    string
		expectError bool
	}{
		{
			name:        "Empty tag starts at v0.1.0",
			latestTag:   "",
			bumpType:    "patch",
			suffix:      "",
			expected:    "v0.1.0",
			expectError: false,
		},
		{
			name:        "Patch bump",
			latestTag:   "v1.2.3",
			bumpType:    "patch",
			suffix:      "",
			expected:    "v1.2.4",
			expectError: false,
		},
		{
			name:        "Minor bump",
			latestTag:   "v1.2.3",
			bumpType:    "minor",
			suffix:      "",
			expected:    "v1.3.0",
			expectError: false,
		},
		{
			name:        "Major bump",
			latestTag:   "v1.2.3",
			bumpType:    "major",
			suffix:      "",
			expected:    "v2.0.0",
			expectError: false,
		},
		{
			name:        "Patch bump with suffix",
			latestTag:   "v1.2.3",
			bumpType:    "patch",
			suffix:      "beta",
			expected:    "v1.2.4-beta",
			expectError: false,
		},
		{
			name:        "Minor bump with suffix",
			latestTag:   "v0.1.0",
			bumpType:    "minor",
			suffix:      "alpha",
			expected:    "v0.2.0-alpha",
			expectError: false,
		},
		{
			name:        "Major bump with suffix",
			latestTag:   "v2.5.9",
			bumpType:    "major",
			suffix:      "rc1",
			expected:    "v3.0.0-rc1",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculateNextVersion(tt.latestTag, tt.bumpType, tt.suffix)
			if (err != nil) != tt.expectError {
				t.Errorf("calculateNextVersion() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("calculateNextVersion() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateDevVersion tests the pure function for calculating dev version
func TestCalculateDevVersion(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		expected    string
		expectError bool
	}{
		{
			name:        "Basic version",
			tag:         "v1.0.0",
			expected:    "1.0.1-dev",
			expectError: false,
		},
		{
			name:        "Version with patch",
			tag:         "v0.1.0",
			expected:    "0.1.1-dev",
			expectError: false,
		},
		{
			name:        "Higher version numbers",
			tag:         "v2.5.9",
			expected:    "2.5.10-dev",
			expectError: false,
		},
		{
			name:        "Version with suffix (should still work)",
			tag:         "v1.2.3-beta",
			expected:    "1.2.4-dev",
			expectError: false,
		},
		{
			name:        "Invalid tag format",
			tag:         "invalid",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Empty tag",
			tag:         "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Missing v prefix",
			tag:         "1.0.0",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculateDevVersion(tt.tag)
			if (err != nil) != tt.expectError {
				t.Errorf("calculateDevVersion() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("calculateDevVersion() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestFormatBumpMessage tests the pure function for formatting success messages
func TestFormatBumpMessage(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		pushed   bool
		expected string
	}{
		{
			name:     "Tag created and pushed",
			tag:      "v1.0.0",
			pushed:   true,
			expected: "Successfully created and pushed tag v1.0.0",
		},
		{
			name:     "Tag created but not pushed",
			tag:      "v1.0.0",
			pushed:   false,
			expected: "Successfully created tag v1.0.0. To push, run: git push --tags",
		},
		{
			name:     "Tag with suffix pushed",
			tag:      "v2.5.3-beta",
			pushed:   true,
			expected: "Successfully created and pushed tag v2.5.3-beta",
		},
		{
			name:     "Tag with suffix not pushed",
			tag:      "v0.1.0-alpha",
			pushed:   false,
			expected: "Successfully created tag v0.1.0-alpha. To push, run: git push --tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBumpMessage(tt.tag, tt.pushed)
			if result != tt.expected {
				t.Errorf("formatBumpMessage() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestFormatDryRunMessage tests the pure function for formatting dry-run messages
func TestFormatDryRunMessage(t *testing.T) {
	tests := []struct {
		name           string
		tag            string
		wouldPush      bool
		updateFile     string
		expectedOutput []string // Expected substrings in output
	}{
		{
			name:       "Basic dry run",
			tag:        "v1.0.0",
			wouldPush:  false,
			updateFile: "",
			expectedOutput: []string{
				"Would create tag: v1.0.0",
			},
		},
		{
			name:       "Dry run with push",
			tag:        "v1.2.3",
			wouldPush:  true,
			updateFile: "",
			expectedOutput: []string{
				"Would create tag: v1.2.3",
				"Would push tag to remote",
			},
		},
		{
			name:       "Dry run with file update",
			tag:        "v2.0.0",
			wouldPush:  false,
			updateFile: "version.go",
			expectedOutput: []string{
				"Would create tag: v2.0.0",
				"Would update file: version.go",
			},
		},
		{
			name:       "Dry run with push and file update",
			tag:        "v0.5.0-beta",
			wouldPush:  true,
			updateFile: "pkg/version/version.go",
			expectedOutput: []string{
				"Would create tag: v0.5.0-beta",
				"Would push tag to remote",
				"Would update file: pkg/version/version.go",
			},
		},
		{
			name:       "Dry run no optional flags",
			tag:        "v3.1.4",
			wouldPush:  false,
			updateFile: "",
			expectedOutput: []string{
				"Would create tag: v3.1.4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDryRunMessage(tt.tag, tt.wouldPush, tt.updateFile)

			// Check that all expected substrings are present
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(result, expected) {
					t.Errorf("formatDryRunMessage() output missing expected substring:\nGot: %v\nExpected to contain: %v", result, expected)
				}
			}

			// Verify wouldPush message appears only when expected
			if tt.wouldPush && !strings.Contains(result, "Would push tag to remote") {
				t.Errorf("formatDryRunMessage() missing push message when wouldPush=true")
			}
			if !tt.wouldPush && strings.Contains(result, "Would push tag to remote") {
				t.Errorf("formatDryRunMessage() includes push message when wouldPush=false")
			}

			// Verify updateFile message appears only when expected
			if tt.updateFile != "" && !strings.Contains(result, fmt.Sprintf("Would update file: %s", tt.updateFile)) {
				t.Errorf("formatDryRunMessage() missing file update message")
			}
			if tt.updateFile == "" && strings.Contains(result, "Would update file") {
				t.Errorf("formatDryRunMessage() includes file update message when updateFile is empty")
			}
		})
	}
}
