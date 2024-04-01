package bump

import (
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
		expectedTag string
		expectError bool
	}{
		{
			name:        "Bump minor version",
			currentTag:  "v1.2.3",
			bumpType:    "minor",
			expectedTag: "v1.3.0",
			expectError: false,
		},
		{
			name:        "Bump major version",
			currentTag:  "v1.2.3",
			bumpType:    "major",
			expectedTag: "v2.0.0",
			expectError: false,
		},
		{
			name:        "Bump patch version",
			currentTag:  "v1.2.3",
			bumpType:    "patch",
			expectedTag: "v1.2.4",
			expectError: false,
		},
		{
			name:        "Invalid bump type",
			currentTag:  "v1.2.3",
			bumpType:    "invalid",
			expectedTag: "",
			expectError: true,
		},
		{
			name:        "Invalid current tag - 1",
			currentTag:  "",
			bumpType:    "patch",
			expectedTag: "",
			expectError: true,
		},
		{
			name:        "Invalid current tag - 2",
			currentTag:  "v",
			bumpType:    "minor",
			expectedTag: "",
			expectError: true,
		},
		{
			name:        "suffixed version",
			currentTag:  "v0.1.0-prerelease",
			bumpType:    "minor",
			expectedTag: "v0.2.0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextTag, err := GetNextTag(tt.currentTag, tt.bumpType)
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
