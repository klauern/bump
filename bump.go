package bump

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// NewGitInfo returns the git version and commit hash of a given path
func NewGitInfo(path string) ([]string, error) {
	var versions []string

	r, err := git.PlainOpen(path)
	if err != nil {
		log.Error("Error opening git repository: ", "err", err)
		return nil, err
	}

	tagRefs, err := r.Tags()
	if err != nil {
		log.Error("Error getting tags: ", "err", err)
		return nil, err
	}
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		if tagRef.Name().IsTag() {
			if strings.HasPrefix("v", tagRef.Name().String()) {
				log.Debug("adding tag", "tag", tagRef.Name().String())
				versions = append(versions, tagRef.Name().String())
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return versions, nil
}

// semanticVersionRegex is a regular expression for parsing semantic version strings.
var semanticVersionRegex = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(-[0-9A-Za-z-.]+)?$`)

// tagVersion represents a semantic version and its corresponding git tag.
type tagVersion struct {
	Major, Minor, Patch int
	Suffix              string
	Tag                 string
}

// ParseTagVersion parses a semantic version string into a tagVersion struct.
func ParseTagVersion(tag string) (*tagVersion, bool) {
	matches := semanticVersionRegex.FindStringSubmatch(tag)
	if matches == nil {
		return nil, false
	}
	return &tagVersion{
		Major:  parseInt(matches[1]),
		Minor:  parseInt(matches[2]),
		Patch:  parseInt(matches[3]),
		Suffix: matches[4], // Capture the optional suffix.
		Tag:    tag,
	}, true
}

// Adjusted sorting logic to account for suffixes. Versions without suffixes are considered newer than those with them.
func sortVersions(versions []*tagVersion) {
	sort.Slice(versions, func(i, j int) bool {
		if versions[i].Major != versions[j].Major {
			return versions[i].Major > versions[j].Major
		}
		if versions[i].Minor != versions[j].Minor {
			return versions[i].Minor > versions[j].Minor
		}
		if versions[i].Patch != versions[j].Patch {
			return versions[i].Patch > versions[j].Patch
		}
		// If versions are equal, compare suffixes. No suffix is considered newer.
		if versions[i].Suffix == "" && versions[j].Suffix != "" {
			return true
		}
		if versions[i].Suffix != "" && versions[j].Suffix == "" {
			return false
		}
		// If both have suffixes, use string comparison.
		return versions[i].Suffix < versions[j].Suffix
	})
}

// GetLatestTag returns the latest tag in a reference iterator.
func GetLatestTag(tagRefs storer.ReferenceIter) (string, error) {
	var versions []*tagVersion

	err := tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tag := ref.Name().Short()
		if version, ok := ParseTagVersion(tag); ok {
			versions = append(versions, version)
		}
		return nil
	})
	if err != nil {
		log.Error("error getting latest tag", err)
		return "", err
	}

	sortVersions(versions)

	if len(versions) > 0 {
		return versions[0].Tag, nil
	}

	log.Debug("No semantic version tags found")
	return "", nil
}

// GetNextTag takes the current latest tag, the bump type (major, minor, patch),
// and an optional suffix, returning the next version tag.
func GetNextTag(currentTag, bumpType, suffix string) (string, error) {
	version, ok := ParseTagVersion(currentTag)
	if !ok {
		log.Error("invalid current tag", "currentTag", currentTag)
		return "", fmt.Errorf("invalid current tag format: %s", currentTag)
	}

	switch bumpType {
	case "major":
		version.Major++
		version.Minor = 0
		version.Patch = 0
	case "minor":
		version.Minor++
		version.Patch = 0
	case "patch":
		version.Patch++
	default:
		log.Error("unknown bump type", "bumpType", bumpType)
		return "", fmt.Errorf("unknown bump type: %s", bumpType)
	}

	// If a non-empty suffix is passed, use it to update the version's suffix.
	// This allows for flexible handling of prerelease or build metadata.
	if suffix != "" {
		version.Suffix = "-" + suffix // Ensures correct formatting with the leading dash.
	} else {
		version.Suffix = "" // Ensures no suffix is added if empty.
	}

	// Construct the next version tag string.
	nextTag := fmt.Sprintf("v%d.%d.%d%s", version.Major, version.Minor, version.Patch, version.Suffix)
	return nextTag, nil
}

// ParseInt safely converts a string to an int. Returns 0 on error.
func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// TagAndPush tags and pushes a new version of the repository at repoPath with the given tag.
func TagAndPush(repoPath, tag string) error {
	// Create the new tag
	cmdTag := exec.Command("git", "tag", tag)
	if err := cmdTag.Run(); err != nil {
		log.Error("failed to create tag", "err", err)
		return fmt.Errorf("failed to create tag: %w", err)
	}

	// Push the new tag to the default remote
	cmdPush := exec.Command("git", "push", "--tags")
	if err := cmdPush.Run(); err != nil {
		log.Error("failed to push tag", "err", err)
		return fmt.Errorf("failed to push tag: %w", err)
	}

	return nil
}
