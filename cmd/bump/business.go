package main

import (
	"fmt"

	"github.com/klauern/bump"
)

// calculateNextVersion determines the next semantic version tag based on the latest tag,
// bump type (patch/minor/major), and optional suffix.
// This is a pure function with no I/O dependencies.
func calculateNextVersion(latestTag, bumpType, suffix string) (string, error) {
	if latestTag == "" {
		return "v0.1.0", nil
	}
	return bump.GetNextTag(latestTag, bumpType, suffix)
}

// calculateDevVersion generates a development version string from a tag.
// It parses the tag and increments the patch version with a "-dev" suffix.
// This is a pure function with no I/O dependencies.
func calculateDevVersion(tag string) (string, error) {
	version, ok := bump.ParseTagVersion(tag)
	if !ok {
		return "", fmt.Errorf("failed to parse tag: %s", tag)
	}
	return fmt.Sprintf("%d.%d.%d-dev", version.Major, version.Minor, version.Patch+1), nil
}

// formatBumpMessage returns the success message after creating a tag.
// The message varies based on whether the tag was pushed to remote.
// This is a pure function with no I/O dependencies.
func formatBumpMessage(tag string, pushed bool) string {
	if pushed {
		return fmt.Sprintf("Successfully created and pushed tag %s", tag)
	}
	return fmt.Sprintf("Successfully created tag %s. To push, run: git push --tags", tag)
}

// formatDryRunMessage returns a preview message for dry-run mode.
// It shows what would be created without making actual changes.
// This is a pure function with no I/O dependencies.
func formatDryRunMessage(tag string, wouldPush bool, updateFile string) string {
	var msg string
	msg = fmt.Sprintf("Would create tag: %s\n", tag)
	if wouldPush {
		msg += "Would push tag to remote\n"
	}
	if updateFile != "" {
		msg += fmt.Sprintf("Would update file: %s\n", updateFile)
	}
	return msg
}
