package bump

import (
	"os/exec"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	// "github.com/klauern/bump"
)

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
