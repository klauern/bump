package bump

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	// "github.com/klauern/bump"
)

// newTempRepo creates a temporary repository structure for testing
func newTempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	// ensure a config exists for preference tests
	cfg := filepath.Join(dir, ".git", "config")
	if err := os.WriteFile(cfg, []byte(""), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return dir
}

func TestNewGitInfo(t *testing.T) {
	// TODO: Replace with actual path
	_, err := NewGitInfo(".")
	if err != nil {
		t.Errorf("NewGitInfo error = %v", err)
	}
}

func TestParseTagVersion(t *testing.T) {
	version, ok := ParseTagVersion("v1.2.3")
	if !ok {
		t.Errorf("Expected ok to be true, got false")
	}

	if version.Major != 1 {
		t.Errorf("Expected version.Major to be 1, got %d", version.Major)
	}

	if version.Minor != 2 {
		t.Errorf("Expected version.Minor to be 2, got %d", version.Minor)
	}

	if version.Patch != 3 {
		t.Errorf("Expected version.Patch to be 3, got %d", version.Patch)
	}
}

func TestSortVersions(t *testing.T) {
	versions := []*tagVersion{
		{Major: 1, Minor: 0, Patch: 0, Tag: "v1.0.0"},
		{Major: 2, Minor: 0, Patch: 0, Tag: "v2.0.0"},
		{Major: 1, Minor: 1, Patch: 0, Tag: "v1.1.0"},
	}
	sortVersions(versions)
	if versions[0].Tag != "v2.0.0" {
		t.Errorf("Expected versions[0].Tag to be 'v2.0.0', got '%s'", versions[0].Tag)
	}

	if versions[1].Tag != "v1.1.0" {
		t.Errorf("Expected versions[1].Tag to be 'v1.1.0', got '%s'", versions[1].Tag)
	}

	if versions[2].Tag != "v1.0.0" {
		t.Errorf("Expected versions[2].Tag to be 'v1.0.0', got '%s'", versions[2].Tag)
	}
}

func TestGetLatestTag(t *testing.T) {
	refs := []plumbing.Reference{
		*plumbing.NewReferenceFromStrings("refs/tags/v0.1.0", "a670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf43"),
		*plumbing.NewReferenceFromStrings("refs/tags/v0.2.0", "b670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf44"),
		*plumbing.NewReferenceFromStrings("refs/tags/v0.3.0", "c670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf45"),
		*plumbing.NewReferenceFromStrings("refs/tags/v1.0.0", "d670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf46"),
		// Add more references as needed
	}
	tagRefs := NewMockReferenceIter(refs)

	tag, err := GetLatestTag(tagRefs)
	if err != nil {
		t.Errorf("GetLatestTag error = %v", err)
	}
	// TODO: Replace with expected latest tag
	if tag != "v1.0.0" {
		t.Errorf("Expected tag to be 'v1.0.0', got '%s'", tag)
	}
}

func TestGetNextTag(t *testing.T) {
	tests := []struct {
		name        string
		currentTag  string
		bumpType    string
		suffix      string
		expectedTag string
		expectError bool
	}{
		{
			name:        "Bump minor version",
			currentTag:  "v1.2.3",
			bumpType:    "minor",
			suffix:      "",
			expectedTag: "v1.3.0",
			expectError: false,
		},
		{
			name:        "Bump major version",
			currentTag:  "v1.2.3",
			bumpType:    "major",
			suffix:      "",
			expectedTag: "v2.0.0",
			expectError: false,
		},
		{
			name:        "Bump patch version",
			currentTag:  "v1.2.3",
			bumpType:    "patch",
			suffix:      "",
			expectedTag: "v1.2.4",
			expectError: false,
		},
		{
			name:        "Invalid bump type",
			currentTag:  "v1.2.3",
			bumpType:    "invalid",
			suffix:      "",
			expectedTag: "",
			expectError: true,
		},
		{
			name:        "Invalid current tag - 1",
			currentTag:  "",
			bumpType:    "patch",
			suffix:      "",
			expectedTag: "",
			expectError: true,
		},
		{
			name:        "Invalid current tag - 2",
			currentTag:  "v",
			bumpType:    "minor",
			suffix:      "",
			expectedTag: "",
			expectError: true,
		},
		{
			name:        "suffixed version",
			currentTag:  "v0.1.0-prerelease",
			bumpType:    "minor",
			suffix:      "",
			expectedTag: "v0.2.0",
			expectError: false,
		},
		{
			name:        "patch with suffix",
			currentTag:  "v0.1.0",
			bumpType:    "patch",
			suffix:      "prerelease",
			expectedTag: "v0.1.1-prerelease",
			expectError: false,
		},
		{
			name:        "minor with suffix",
			currentTag:  "v0.1.1",
			bumpType:    "minor",
			suffix:      "prerelease",
			expectedTag: "v0.2.0-prerelease",
			expectError: false,
		},
		{
			name:        "major with suffix",
			currentTag:  "v0.1.1",
			bumpType:    "major",
			suffix:      "rc1",
			expectedTag: "v1.0.0-rc1",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextTag, err := GetNextTag(tt.currentTag, tt.bumpType, tt.suffix)
			if (err != nil) != tt.expectError {
				t.Errorf("GetNextTag() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if nextTag != tt.expectedTag {
				t.Errorf("Expected nextTag to be '%s', got '%s'", tt.expectedTag, nextTag)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	if result := parseInt("123"); result != 123 {
		t.Errorf("Expected ParseInt('123') to be 123, got %d", result)
	}
	if result := parseInt("abc"); result != 0 {
		t.Errorf("Expected ParseInt('abc') to be 0, got %d", result)
	}
}

func TestOpenGitRepoInvalidPath(t *testing.T) {
	// Test case to ensure openGitRepo returns an error for an invalid path
	repo, err := openGitRepo("/invalid/path")
	if err == nil {
		t.Errorf("Expected error for invalid path, got nil")
	}
	if repo != nil {
		t.Errorf("Expected repo to be nil for invalid path")
	}
}

func TestCreateTag(t *testing.T) {
	// Test case to ensure createTag returns an error for an invalid command
	err := createTag("")
	if err == nil {
		t.Errorf("Expected error for invalid tag command, got nil")
	}
}

func TestCompareVersionsEqual(t *testing.T) {
	// This test ensures compareVersions returns false for equal versions
	version1 := &tagVersion{Major: 1, Minor: 0, Patch: 0}
	version2 := &tagVersion{Major: 1, Minor: 0, Patch: 0}
	if compareVersions(version1, version2) {
		t.Errorf("Expected compareVersions to return false for equal versions")
	}
}

func TestNewGitInfoInvalidPath(t *testing.T) {
	// This test ensures NewGitInfo returns an error for an invalid path
	_, err := NewGitInfo("/invalid/path")
	if err == nil {
		t.Errorf("Expected error for invalid path, got nil")
	}
}

func TestCreateTagInvalid(t *testing.T) {
	// This test ensures CreateTag returns an error for an invalid tag
	err := CreateTag("")
	if err == nil {
		t.Errorf("Expected error for invalid tag, got nil")
	}
}

func TestPushTagInvalid(t *testing.T) {
	// Override execCommand to simulate a failure
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		// Return a command that always fails
		return exec.Command("false")
	}

	err := PushTag()
	if err == nil {
		t.Errorf("Expected error for push outside a git repo, got nil")
	}
}

func TestCompareVersionsHigherPatch(t *testing.T) {
	// This test ensures compareVersions correctly compares versions with different patch numbers
	version1 := &tagVersion{Major: 1, Minor: 0, Patch: 1}
	version2 := &tagVersion{Major: 1, Minor: 0, Patch: 2}
	if !compareVersions(version2, version1) {
		t.Errorf("Expected version2 to be greater than version1 by patch")
	}
	if compareVersions(version1, version2) {
		t.Errorf("Expected version1 to be less than version2 by patch")
	}
}

// TestCompareSuffixes tests the compareSuffixes function with various suffix combinations
func TestCompareSuffixes(t *testing.T) {
	tests := []struct {
		name     string
		suffix1  string
		suffix2  string
		expected bool
	}{
		{
			name:     "Empty suffix1, non-empty suffix2 (no suffix is greater)",
			suffix1:  "",
			suffix2:  "-alpha",
			expected: true,
		},
		{
			name:     "Non-empty suffix1, empty suffix2 (no suffix is greater)",
			suffix1:  "-alpha",
			suffix2:  "",
			expected: false,
		},
		{
			name:     "Both empty suffixes",
			suffix1:  "",
			suffix2:  "",
			expected: false,
		},
		{
			name:     "alpha < beta (beta should come first in descending sort)",
			suffix1:  "-alpha",
			suffix2:  "-beta",
			expected: false,
		},
		{
			name:     "beta > alpha (beta should come first in descending sort)",
			suffix1:  "-beta",
			suffix2:  "-alpha",
			expected: true,
		},
		{
			name:     "Equal suffixes",
			suffix1:  "-alpha",
			suffix2:  "-alpha",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareSuffixes(tt.suffix1, tt.suffix2)
			if result != tt.expected {
				t.Errorf("compareSuffixes(%q, %q) = %v, expected %v", tt.suffix1, tt.suffix2, result, tt.expected)
			}
		})
	}
}

// TestCompareSuffixesSemVer2 tests compareSuffixes according to SemVer 2.0 specification
func TestCompareSuffixesSemVer2(t *testing.T) {
	tests := []struct {
		name     string
		suffix1  string
		suffix2  string
		expected bool // true if suffix1 > suffix2 (for descending sort)
	}{
		// Stable vs pre-release
		{
			name:     "stable > pre-release",
			suffix1:  "",
			suffix2:  "-alpha",
			expected: true,
		},
		{
			name:     "pre-release < stable",
			suffix1:  "-alpha",
			suffix2:  "",
			expected: false,
		},
		// Numeric comparison within identifiers
		{
			name:     "beta.11 > beta.2 (numeric comparison)",
			suffix1:  "-beta.11",
			suffix2:  "-beta.2",
			expected: true,
		},
		{
			name:     "beta.2 < beta.11 (numeric comparison)",
			suffix1:  "-beta.2",
			suffix2:  "-beta.11",
			expected: false,
		},
		{
			name:     "alpha.1 < alpha.2",
			suffix1:  "-alpha.1",
			suffix2:  "-alpha.2",
			expected: false,
		},
		// Numeric vs alphanumeric: numeric has lower precedence
		{
			name:     "alpha.1 < alpha.beta (numeric < alphanumeric)",
			suffix1:  "-alpha.1",
			suffix2:  "-alpha.beta",
			expected: false,
		},
		{
			name:     "alpha.beta > alpha.1 (alphanumeric > numeric)",
			suffix1:  "-alpha.beta",
			suffix2:  "-alpha.1",
			expected: true,
		},
		{
			name:     "beta.2 < beta.11 < beta.rc",
			suffix1:  "-beta.11",
			suffix2:  "-beta.rc",
			expected: false,
		},
		// Longer list has higher precedence when all preceding are equal
		{
			name:     "alpha.1 > alpha (more identifiers)",
			suffix1:  "-alpha.1",
			suffix2:  "-alpha",
			expected: true,
		},
		{
			name:     "alpha < alpha.1 (fewer identifiers)",
			suffix1:  "-alpha",
			suffix2:  "-alpha.1",
			expected: false,
		},
		{
			name:     "alpha.beta.gamma > alpha.beta",
			suffix1:  "-alpha.beta.gamma",
			suffix2:  "-alpha.beta",
			expected: true,
		},
		// Lexical comparison for alphanumeric
		{
			name:     "alpha < beta (lexical)",
			suffix1:  "-alpha",
			suffix2:  "-beta",
			expected: false,
		},
		{
			name:     "beta > alpha (lexical)",
			suffix1:  "-beta",
			suffix2:  "-alpha",
			expected: true,
		},
		{
			name:     "rc > beta (lexical)",
			suffix1:  "-rc",
			suffix2:  "-beta",
			expected: true,
		},
		// SemVer 2.0 canonical example sequence:
		// 1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta < 1.0.0-beta.2 < 1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0
		{
			name:     "alpha < alpha.1",
			suffix1:  "-alpha",
			suffix2:  "-alpha.1",
			expected: false,
		},
		{
			name:     "alpha.1 < alpha.beta",
			suffix1:  "-alpha.1",
			suffix2:  "-alpha.beta",
			expected: false,
		},
		{
			name:     "alpha.beta < beta",
			suffix1:  "-alpha.beta",
			suffix2:  "-beta",
			expected: false,
		},
		{
			name:     "beta < beta.2",
			suffix1:  "-beta",
			suffix2:  "-beta.2",
			expected: false,
		},
		{
			name:     "beta.2 < beta.11",
			suffix1:  "-beta.2",
			suffix2:  "-beta.11",
			expected: false,
		},
		{
			name:     "beta.11 < rc.1",
			suffix1:  "-beta.11",
			suffix2:  "-rc.1",
			expected: false,
		},
		{
			name:     "rc.1 < stable",
			suffix1:  "-rc.1",
			suffix2:  "",
			expected: false,
		},
		// Equal identifiers
		{
			name:     "alpha.1 == alpha.1",
			suffix1:  "-alpha.1",
			suffix2:  "-alpha.1",
			expected: false,
		},
		{
			name:     "beta.11 == beta.11",
			suffix1:  "-beta.11",
			suffix2:  "-beta.11",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareSuffixes(tt.suffix1, tt.suffix2)
			if result != tt.expected {
				t.Errorf("compareSuffixes(%q, %q) = %v, expected %v", tt.suffix1, tt.suffix2, result, tt.expected)
			}
		})
	}
}

// TestParseNumericIdentifier tests the parseNumericIdentifier function
func TestParseNumericIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		identifier  string
		expectedNum int
		expectedOk  bool
	}{
		{
			name:        "Simple numeric",
			identifier:  "123",
			expectedNum: 123,
			expectedOk:  true,
		},
		{
			name:        "Zero",
			identifier:  "0",
			expectedNum: 0,
			expectedOk:  true,
		},
		{
			name:        "Large number",
			identifier:  "999999",
			expectedNum: 999999,
			expectedOk:  true,
		},
		{
			name:        "Alphanumeric",
			identifier:  "alpha",
			expectedNum: 0,
			expectedOk:  false,
		},
		{
			name:        "Mixed alphanumeric",
			identifier:  "beta1",
			expectedNum: 0,
			expectedOk:  false,
		},
		{
			name:        "With dash",
			identifier:  "1-2",
			expectedNum: 0,
			expectedOk:  false,
		},
		{
			name:        "Empty string",
			identifier:  "",
			expectedNum: 0,
			expectedOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num, ok := parseNumericIdentifier(tt.identifier)
			if ok != tt.expectedOk {
				t.Errorf("parseNumericIdentifier(%q) ok = %v, expected %v", tt.identifier, ok, tt.expectedOk)
			}
			if ok && num != tt.expectedNum {
				t.Errorf("parseNumericIdentifier(%q) num = %v, expected %v", tt.identifier, num, tt.expectedNum)
			}
		})
	}
}

// TestSortVersionsSemVer2 tests that version sorting follows SemVer 2.0 specification
func TestSortVersionsSemVer2(t *testing.T) {
	// Test the canonical SemVer 2.0 example sequence
	versions := []*tagVersion{
		{Major: 1, Minor: 0, Patch: 0, Suffix: "", Tag: "v1.0.0"},
		{Major: 1, Minor: 0, Patch: 0, Suffix: "-rc.1", Tag: "v1.0.0-rc.1"},
		{Major: 1, Minor: 0, Patch: 0, Suffix: "-beta.11", Tag: "v1.0.0-beta.11"},
		{Major: 1, Minor: 0, Patch: 0, Suffix: "-beta.2", Tag: "v1.0.0-beta.2"},
		{Major: 1, Minor: 0, Patch: 0, Suffix: "-beta", Tag: "v1.0.0-beta"},
		{Major: 1, Minor: 0, Patch: 0, Suffix: "-alpha.beta", Tag: "v1.0.0-alpha.beta"},
		{Major: 1, Minor: 0, Patch: 0, Suffix: "-alpha.1", Tag: "v1.0.0-alpha.1"},
		{Major: 1, Minor: 0, Patch: 0, Suffix: "-alpha", Tag: "v1.0.0-alpha"},
	}

	sortVersions(versions)

	// After sorting in descending order, the expected order is:
	expected := []string{
		"v1.0.0",           // stable version highest
		"v1.0.0-rc.1",      // rc > beta
		"v1.0.0-beta.11",   // beta.11 > beta.2 (numeric comparison)
		"v1.0.0-beta.2",    // beta.2 > beta (more identifiers)
		"v1.0.0-beta",      // beta > alpha.beta (lexical)
		"v1.0.0-alpha.beta", // alpha.beta > alpha.1 (alphanumeric > numeric)
		"v1.0.0-alpha.1",   // alpha.1 > alpha (more identifiers)
		"v1.0.0-alpha",     // alpha lowest
	}

	for i, v := range versions {
		if v.Tag != expected[i] {
			t.Errorf("Position %d: expected %s, got %s", i, expected[i], v.Tag)
		}
	}
}

// TestValidateRepositoryPath tests the validateRepositoryPath function
func TestValidateRepositoryPath(t *testing.T) {
	okRepo := newTempRepo(t)
	tests := []struct {
		name        string
		repoPath    string
		expectError bool
	}{
		{
			name:        "Empty path",
			repoPath:    "",
			expectError: true,
		},
		{
			name:        "Temp git repo",
			repoPath:    okRepo,
			expectError: false,
		},
		{
			name:        "Non-existent path",
			repoPath:    "/nonexistent/path",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRepositoryPath(tt.repoPath)
			if (err != nil) != tt.expectError {
				t.Errorf("validateRepositoryPath(%q) error = %v, expectError %v", tt.repoPath, err, tt.expectError)
			}
		})
	}
}

// TestFindGitRepoRoot tests the findGitRepoRoot function
func TestFindGitRepoRoot(t *testing.T) {
	repo := newTempRepo(t)
	nested := filepath.Join(repo, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	tests := []struct {
		name        string
		startPath   string
		expectError bool
	}{
		{
			name:        "Nested path (should find git root)",
			startPath:   nested,
			expectError: false,
		},
		{
			name:        "Root directory (should fail)",
			startPath:   "/",
			expectError: true,
		},
		{
			name:        "Temp directory (should fail)",
			startPath:   "/tmp",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := findGitRepoRoot(tt.startPath)
			if (err != nil) != tt.expectError {
				t.Errorf("findGitRepoRoot(%q) error = %v, expectError %v", tt.startPath, err, tt.expectError)
			}
		})
	}
}

// TestGetLatestTagEmpty tests GetLatestTag with no valid tags
func TestGetLatestTagEmpty(t *testing.T) {
	// Create empty reference iterator
	refs := []plumbing.Reference{}
	tagRefs := NewMockReferenceIter(refs)

	tag, err := GetLatestTag(tagRefs)
	if err != nil {
		t.Errorf("GetLatestTag with empty tags should not error, got: %v", err)
	}
	if tag != "" {
		t.Errorf("GetLatestTag with empty tags should return empty string, got: %s", tag)
	}
}

// TestGetLatestTagNonSemVer tests GetLatestTag with non-semantic version tags
func TestGetLatestTagNonSemVer(t *testing.T) {
	// Create references with non-semver tags
	refs := []plumbing.Reference{
		*plumbing.NewReferenceFromStrings("refs/tags/release", "a670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf43"),
		*plumbing.NewReferenceFromStrings("refs/tags/foo", "b670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf44"),
	}
	tagRefs := NewMockReferenceIter(refs)

	tag, err := GetLatestTag(tagRefs)
	if err != nil {
		t.Errorf("GetLatestTag should handle non-semver tags gracefully, got error: %v", err)
	}
	if tag != "" {
		t.Errorf("GetLatestTag with only non-semver tags should return empty string, got: %s", tag)
	}
}

// TestGetDefaultPushPreference tests the GetDefaultPushPreference function
func TestGetDefaultPushPreference(t *testing.T) {
	repo := newTempRepo(t)
	tests := []struct {
		name          string
		repoPath      string
		expectError   bool
		expectedValue bool
		expectedIsSet bool
	}{
		{
			name:        "Empty path should error",
			repoPath:    "",
			expectError: true,
		},
		{
			name:        "Non-existent path should error",
			repoPath:    "/nonexistent/path",
			expectError: true,
		},
		{
			name:          "Temp repo (may not have preference set)",
			repoPath:      repo,
			expectError:   false,
			expectedValue: false,
			expectedIsSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, isSet, err := GetDefaultPushPreference(tt.repoPath)
			if (err != nil) != tt.expectError {
				t.Errorf("GetDefaultPushPreference(%q) error = %v, expectError %v", tt.repoPath, err, tt.expectError)
				return
			}
			if !tt.expectError {
				// For the temp repo test, we just verify the function runs without error
				// The actual values depend on whether the preference is set
				_ = value
				_ = isSet
			}
		})
	}
}

// TestSetDefaultPushPreference tests the SetDefaultPushPreference function
func TestSetDefaultPushPreference(t *testing.T) {
	repo := newTempRepo(t)
	tests := []struct {
		name        string
		repoPath    string
		value       bool
		expectError bool
	}{
		{
			name:        "Empty path should error",
			repoPath:    "",
			value:       true,
			expectError: true,
		},
		{
			name:        "Non-existent path should error",
			repoPath:    "/nonexistent/path",
			value:       false,
			expectError: true,
		},
		{
			name:        "Temp repo - set to true",
			repoPath:    repo,
			value:       true,
			expectError: false,
		},
		{
			name:        "Temp repo - set to false",
			repoPath:    repo,
			value:       false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetDefaultPushPreference(tt.repoPath, tt.value)
			if (err != nil) != tt.expectError {
				t.Errorf("SetDefaultPushPreference(%q, %v) error = %v, expectError %v", tt.repoPath, tt.value, err, tt.expectError)
			}

			// If we successfully set a value, verify we can read it back
			if !tt.expectError && tt.repoPath == repo {
				value, isSet, err := GetDefaultPushPreference(tt.repoPath)
				if err != nil {
					t.Errorf("Failed to read back preference: %v", err)
				}
				if !isSet {
					t.Errorf("Expected preference to be set after SetDefaultPushPreference")
				}
				if value != tt.value {
					t.Errorf("Expected value %v, got %v", tt.value, value)
				}
			}
		})
	}
}

// TestMockReferenceIterNext tests the Next method of MockReferenceIter
func TestMockReferenceIterNext(t *testing.T) {
	refs := []plumbing.Reference{
		*plumbing.NewReferenceFromStrings("refs/tags/v1.0.0", "a670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf43"),
		*plumbing.NewReferenceFromStrings("refs/tags/v2.0.0", "b670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf44"),
	}
	iter := NewMockReferenceIter(refs)

	// Test first call to Next
	ref, err := iter.Next()
	if err != nil {
		t.Errorf("First Next() should not error, got: %v", err)
	}
	if ref.Name().String() != "refs/tags/v1.0.0" {
		t.Errorf("Expected first reference to be refs/tags/v1.0.0, got: %s", ref.Name().String())
	}

	// Test second call to Next
	ref, err = iter.Next()
	if err != nil {
		t.Errorf("Second Next() should not error, got: %v", err)
	}
	if ref.Name().String() != "refs/tags/v2.0.0" {
		t.Errorf("Expected second reference to be refs/tags/v2.0.0, got: %s", ref.Name().String())
	}

	// Test third call to Next (should return EOF)
	ref, err = iter.Next()
	if err == nil {
		t.Errorf("Third Next() should return EOF error")
	}
	if ref != nil {
		t.Errorf("Expected nil reference at end of iteration")
	}
}

// TestMockReferenceIterClose tests the Close method of MockReferenceIter
func TestMockReferenceIterClose(t *testing.T) {
	refs := []plumbing.Reference{
		*plumbing.NewReferenceFromStrings("refs/tags/v1.0.0", "a670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf43"),
	}
	iter := NewMockReferenceIter(refs)

	// Advance iterator
	_, err := iter.Next()
	if err != nil {
		t.Fatalf("Next() should not error, got: %v", err)
	}

	// Close should reset the iterator
	iter.Close()

	// After close, Next should start from beginning again
	ref, err := iter.Next()
	if err != nil {
		t.Errorf("Next() after Close() should not error, got: %v", err)
	}
	if ref.Name().String() != "refs/tags/v1.0.0" {
		t.Errorf("Expected first reference after Close(), got: %s", ref.Name().String())
	}
}

// TestMockReferenceIterForEachError tests ForEach with callback that returns error
func TestMockReferenceIterForEachError(t *testing.T) {
	refs := []plumbing.Reference{
		*plumbing.NewReferenceFromStrings("refs/tags/v1.0.0", "a670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf43"),
		*plumbing.NewReferenceFromStrings("refs/tags/v2.0.0", "b670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf44"),
	}
	iter := NewMockReferenceIter(refs)

	// Test ForEach with callback that returns an error
	testErr := fmt.Errorf("test error")
	err := iter.ForEach(func(ref *plumbing.Reference) error {
		return testErr
	})

	if err != testErr {
		t.Errorf("Expected ForEach to return test error, got: %v", err)
	}
}

// TestMockReferenceIterForEachSuccess tests ForEach with successful iteration
func TestMockReferenceIterForEachSuccess(t *testing.T) {
	refs := []plumbing.Reference{
		*plumbing.NewReferenceFromStrings("refs/tags/v1.0.0", "a670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf43"),
		*plumbing.NewReferenceFromStrings("refs/tags/v2.0.0", "b670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf44"),
	}
	iter := NewMockReferenceIter(refs)

	// Test ForEach with successful callback
	count := 0
	err := iter.ForEach(func(ref *plumbing.Reference) error {
		count++
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error from ForEach, got: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected ForEach to iterate 2 times, got: %d", count)
	}
}

// TestAcquireGitLockInvalidPath tests acquireGitLock with invalid paths
func TestAcquireGitLockInvalidPath(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
	}{
		{
			name:     "Empty path",
			repoPath: "",
		},
		{
			name:     "Non-existent path",
			repoPath: "/nonexistent/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lock, err := acquireGitLock(tt.repoPath)
			if err == nil {
				t.Errorf("acquireGitLock(%q) should error for invalid path", tt.repoPath)
				if lock != nil {
					_ = lock.Release()
				}
			}
			if lock != nil {
				t.Errorf("acquireGitLock(%q) should return nil lock for invalid path", tt.repoPath)
			}
		})
	}
}

// TestGitLockReleaseNotAcquired tests Release on a lock that wasn't acquired
func TestGitLockReleaseNotAcquired(t *testing.T) {
	lock := &GitLock{
		lockFile: "",
		acquired: false,
		mutex:    nil,
	}

	err := lock.Release()
	if err != nil {
		t.Errorf("Release() on non-acquired lock should not error, got: %v", err)
	}
}

// TestGetVersionsWithValidReferences tests getVersions successfully processes valid references
func TestGetVersionsWithValidReferences(t *testing.T) {
	refs := []plumbing.Reference{
		*plumbing.NewReferenceFromStrings("refs/tags/v1.0.0", "a670469b3e8a6e2e6d53635b3f3e6b1b8f6bcf43"),
	}
	iter := NewMockReferenceIter(refs)

	versions := getVersions(iter)
	if versions == nil {
		t.Errorf("getVersions should not return nil for valid references")
	}
	if len(versions) == 0 {
		t.Errorf("getVersions should return at least one version")
	}
}

// TestCreateTagError tests CreateTag with empty tag
func TestCreateTagError(t *testing.T) {
	_ = newTempRepo(t) // Create temp repo for isolation even if not directly used

	// Mock execCommand to avoid actual git calls
	orig := execCommand
	defer func() { execCommand = orig }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("false")
	}

	// Test with empty string
	err := CreateTag("")
	if err == nil {
		t.Errorf("CreateTag with empty string should return error")
	}
}

// TestPushTagError tests PushTag error scenarios
func TestPushTagError(t *testing.T) {
	_ = newTempRepo(t) // Create temp repo for isolation even if not directly used

	// Mock the execCommand to simulate failure
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("false")
	}

	err := PushTag()
	if err == nil {
		t.Errorf("PushTag should return error when git push fails")
	}
}

// TestParseTagVersionEdgeCases tests ParseTagVersion with edge cases
func TestParseTagVersionEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		expectOk  bool
	}{
		{
			name:     "Valid version with pre-release",
			tag:      "v1.2.3-alpha",
			expectOk: true,
		},
		{
			name:     "Valid version with build metadata",
			tag:      "v1.2.3-beta.1",
			expectOk: true,
		},
		{
			name:     "Invalid - no v prefix",
			tag:      "1.2.3",
			expectOk: false,
		},
		{
			name:     "Invalid - missing patch",
			tag:      "v1.2",
			expectOk: false,
		},
		{
			name:     "Invalid - non-numeric",
			tag:      "vabc",
			expectOk: false,
		},
		{
			name:     "Empty string",
			tag:      "",
			expectOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := ParseTagVersion(tt.tag)
			if ok != tt.expectOk {
				t.Errorf("ParseTagVersion(%q) ok = %v, expected %v", tt.tag, ok, tt.expectOk)
			}
		})
	}
}

// TestCompareVersionsWithSuffixes tests compareVersions with pre-release suffixes
func TestCompareVersionsWithSuffixes(t *testing.T) {
	tests := []struct {
		name     string
		v1       *tagVersion
		v2       *tagVersion
		expected bool
	}{
		{
			name:     "Same version, v1 has suffix, v2 has no suffix (v2 should be greater)",
			v1:       &tagVersion{Major: 1, Minor: 0, Patch: 0, Suffix: "-alpha"},
			v2:       &tagVersion{Major: 1, Minor: 0, Patch: 0, Suffix: ""},
			expected: false,
		},
		{
			name:     "Same version, v1 has no suffix, v2 has suffix (v1 should be greater)",
			v1:       &tagVersion{Major: 1, Minor: 0, Patch: 0, Suffix: ""},
			v2:       &tagVersion{Major: 1, Minor: 0, Patch: 0, Suffix: "-alpha"},
			expected: true,
		},
		{
			name:     "Same version, beta > alpha per SemVer (beta should come first)",
			v1:       &tagVersion{Major: 1, Minor: 0, Patch: 0, Suffix: "-alpha"},
			v2:       &tagVersion{Major: 1, Minor: 0, Patch: 0, Suffix: "-beta"},
			expected: false,
		},
		{
			name:     "Same version, beta > alpha per SemVer (beta should come first)",
			v1:       &tagVersion{Major: 1, Minor: 0, Patch: 0, Suffix: "-beta"},
			v2:       &tagVersion{Major: 1, Minor: 0, Patch: 0, Suffix: "-alpha"},
			expected: true,
		},
		{
			name:     "Different major versions",
			v1:       &tagVersion{Major: 2, Minor: 0, Patch: 0, Suffix: "-alpha"},
			v2:       &tagVersion{Major: 1, Minor: 0, Patch: 0, Suffix: ""},
			expected: true,
		},
		{
			name:     "Different minor versions",
			v1:       &tagVersion{Major: 1, Minor: 2, Patch: 0, Suffix: ""},
			v2:       &tagVersion{Major: 1, Minor: 1, Patch: 0, Suffix: ""},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("compareVersions() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
