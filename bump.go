// Package bump provides functionality for semantic versioning and git tagging.
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

// semanticVersionRegex is a regular expression for semantic versioning.
var semanticVersionRegex = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(-[0-9A-Za-z-.]+)?$`)

// tagVersion represents a semantic version of a git tag.
type tagVersion struct {
	Major, Minor, Patch int    // Major, Minor, and Patch represent the major, minor, and patch versions respectively.
	Suffix              string // Suffix represents the optional suffix in a semantic version.
	Tag                 string // Tag represents the original tag string.
}

// NewGitInfo returns the semantic versions of all git tags in the repository at the given path.
func NewGitInfo(path string) ([]string, error) {
	r, err := openGitRepo(path)
	if err != nil {
		return nil, err
	}

	tagRefs, err := getTags(r)
	if err != nil {
		return nil, err
	}

	return getVersions(tagRefs), nil
}

// openGitRepo opens a git repository at the given path.
func openGitRepo(path string) (*git.Repository, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		log.Error("Error opening git repository: ", "err", err)
	}
	return r, err
}

// getTags returns all tags in the given git repository.
func getTags(r *git.Repository) (storer.ReferenceIter, error) {
	tagRefs, err := r.Tags()
	if err != nil {
		log.Error("Error getting tags: ", "err", err)
	}
	return tagRefs, err
}

// getVersions returns the semantic versions of the given git tags.
func getVersions(tagRefs storer.ReferenceIter) []string {
	var versions []string
	tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		if tagRef.Name().IsTag() && strings.HasPrefix("v", tagRef.Name().String()) {
			log.Debug("adding tag", "tag", tagRef.Name().String())
			versions = append(versions, tagRef.Name().String())
		}
		return nil
	})
	return versions
}

// ParseTagVersion parses a git tag into a semantic version.
func ParseTagVersion(tag string) (*tagVersion, bool) {
	matches := semanticVersionRegex.FindStringSubmatch(tag)
	if matches == nil {
		return nil, false
	}
	return &tagVersion{
		Major:  parseInt(matches[1]),
		Minor:  parseInt(matches[2]),
		Patch:  parseInt(matches[3]),
		Suffix: matches[4],
		Tag:    tag,
	}, true
}

// sortVersions sorts a slice of semantic versions in descending order.
func sortVersions(versions []*tagVersion) {
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j])
	})
}

// compareVersions compares two semantic versions.
func compareVersions(version1, version2 *tagVersion) bool {
	if version1.Major != version2.Major {
		return version1.Major > version2.Major
	}
	if version1.Minor != version2.Minor {
		return version1.Minor > version2.Minor
	}
	if version1.Patch != version2.Patch {
		return version1.Patch > version2.Patch
	}
	return compareSuffixes(version1.Suffix, version2.Suffix)
}

// compareSuffixes compares two suffixes in semantic versions.
func compareSuffixes(suffix1, suffix2 string) bool {
	if suffix1 == "" && suffix2 != "" {
		return true
	}
	if suffix1 != "" && suffix2 == "" {
		return false
	}
	return suffix1 < suffix2
}

// GetLatestTag returns the latest semantic version tag in the given git tags.
func GetLatestTag(tagRefs storer.ReferenceIter) (string, error) {
	versions, err := getTagVersions(tagRefs)
	if err != nil {
		return "", err
	}

	sortVersions(versions)

	if len(versions) > 0 {
		return versions[0].Tag, nil
	}

	log.Debug("No semantic version tags found")
	return "", nil
}

// getTagVersions returns the semantic versions of the given git tags.
func getTagVersions(tagRefs storer.ReferenceIter) ([]*tagVersion, error) {
	var versions []*tagVersion
	err := tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tag := ref.Name().Short()
		if version, ok := ParseTagVersion(tag); ok {
			versions = append(versions, version)
		}
		return nil
	})
	return versions, err
}

// GetNextTag returns the next semantic version tag based on the given current tag and bump type.
func GetNextTag(currentTag, bumpType, suffix string) (string, error) {
	version, ok := ParseTagVersion(currentTag)
	if !ok {
		log.Error("invalid current tag", "currentTag", currentTag)
		return "", fmt.Errorf("invalid current tag format: %s", currentTag)
	}

	err := updateVersion(version, bumpType, suffix)
	if err != nil {
		return "", err
	}

	nextTag := fmt.Sprintf("v%d.%d.%d%s", version.Major, version.Minor, version.Patch, version.Suffix)
	return nextTag, nil
}

// updateVersion updates a semantic version based on the given bump type and suffix.
func updateVersion(version *tagVersion, bumpType, suffix string) error {
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
		return fmt.Errorf("unknown bump type: %s", bumpType)
	}

	if suffix != "" {
		version.Suffix = "-" + suffix
	} else {
		version.Suffix = ""
	}

	return nil
}

// parseInt converts a string to an integer, defaulting to 0 on error.
func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// TagAndPush creates a new git tag with the given tag and pushes it to the remote repository.
func TagAndPush(repoPath, tag string) error {
	if err := createTag(tag); err != nil {
		return err
	}

	if err := pushTag(); err != nil {
		return err
	}

	return nil
}

// createTag creates a new git tag with the given tag.
func createTag(tag string) error {
	cmdTag := exec.Command("git", "tag", tag)
	if err := cmdTag.Run(); err != nil {
		log.Error("failed to create tag", "err", err)
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

// pushTag pushes the latest git tag to the remote repository.
func pushTag() error {
	cmdPush := exec.Command("git", "push", "--tags")
	if err := cmdPush.Run(); err != nil {
		log.Error("failed to push tag", "err", err)
		return fmt.Errorf("failed to push tag: %w", err)
	}
	return nil
}
