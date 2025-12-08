// Package bump provides functionality for semantic versioning and git tagging.
package bump

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/ini.v1"

	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// execCommand is a variable to hold the exec.Command function for easier testing and mocking.
var execCommand = exec.Command

// semanticVersionRegex is a regular expression for semantic versioning.
var semanticVersionRegex = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(-[0-9A-Za-z-.]+)?$`)

// gitLocks stores file-based locks per repository to prevent concurrent git operations.
var gitLocks = make(map[string]*sync.Mutex)

// gitLocksMutex protects concurrent access to the gitLocks map.
var gitLocksMutex sync.RWMutex

// GitLock represents a file-based lock for git operations.
type GitLock struct {
	lockFile string       // lockFile is the path to the lock file
	acquired bool         // acquired indicates whether the lock has been successfully acquired
	mutex    *sync.Mutex // mutex is the in-process mutex for this repository
}

// acquireGitLock acquires a file-based lock for git operations on the specified repository.
// This prevents concurrent git operations that could corrupt the repository state.
func acquireGitLock(repoPath string) (*GitLock, error) {
	// Validate repository path first
	if err := validateRepositoryPath(repoPath); err != nil {
		return nil, fmt.Errorf("invalid repository for git lock: %w", err)
	}

	absRepoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repository path: %w", err)
	}

	// Get or create a mutex for this repository
	gitLocksMutex.Lock()
	if gitLocks[absRepoPath] == nil {
		gitLocks[absRepoPath] = &sync.Mutex{}
	}
	repoMutex := gitLocks[absRepoPath]
	gitLocksMutex.Unlock()

	// Acquire the in-process mutex first
	repoMutex.Lock()

	lockFile := filepath.Join(absRepoPath, ".git", "bump.lock")

	// Try to acquire file-based lock with timeout
	const maxAttempts = 30
	const lockTimeout = 100 * time.Millisecond

	var lockFileHandle *os.File
	for i := 0; i < maxAttempts; i++ {
		lockFileHandle, err = os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			break
		}

		if !os.IsExist(err) {
			repoMutex.Unlock()
			return nil, fmt.Errorf("failed to create lock file: %w", err)
		}

		// Check if existing lock file is stale (older than 5 minutes)
		if stat, statErr := os.Stat(lockFile); statErr == nil {
			if time.Since(stat.ModTime()) > 5*time.Minute {
				log.Warn("Removing stale lock file", "lockFile", lockFile, "age", time.Since(stat.ModTime()))
				if err := os.Remove(lockFile); err != nil {
					log.Error("failed to remove stale lock file", "lockFile", lockFile, "err", err)
				}
				continue
			}
		}

		time.Sleep(lockTimeout)
	}

	if lockFileHandle == nil {
		repoMutex.Unlock()
		return nil, fmt.Errorf("failed to acquire git lock after %d attempts: repository may be busy", maxAttempts)
	}

	// Write process info to lock file
	if _, err := fmt.Fprintf(lockFileHandle, "pid: %d\ntime: %s\n", os.Getpid(), time.Now().Format(time.RFC3339)); err != nil {
		log.Error("failed to write to lock file", "lockFile", lockFile, "err", err)
	}
	if err := lockFileHandle.Close(); err != nil {
		log.Error("failed to close lock file", "lockFile", lockFile, "err", err)
	}

	return &GitLock{
		lockFile: lockFile,
		acquired: true,
		mutex:    repoMutex,
	}, nil
}

// Release releases the git lock, removing the lock file and releasing the mutex.
func (lock *GitLock) Release() error {
	if !lock.acquired {
		return nil
	}

	// Remove lock file
	if err := os.Remove(lock.lockFile); err != nil && !os.IsNotExist(err) {
		log.Error("failed to remove lock file", "lockFile", lock.lockFile, "err", err)
	}

	// Release in-process mutex
	lock.mutex.Unlock()
	lock.acquired = false

	return nil
}

// tagVersion represents a semantic version of a git tag.
type tagVersion struct {
	Major  int    // Major is the major version number
	Minor  int    // Minor is the minor version number
	Patch  int    // Patch is the patch version number
	Suffix string // Suffix is the optional pre-release suffix (e.g., "-alpha", "-beta.1")
	Tag    string // Tag is the original git tag string
}

// NewGitInfo scans the git repository at the given path and returns all semantic version tags.
// It opens the repository, fetches all tags, parses them as semantic versions, and returns
// the tag strings in descending order (newest first). Returns an error if the repository
// cannot be opened or tags cannot be fetched.
func NewGitInfo(path string) ([]string, error) {
	r, err := openGitRepo(path)
	if err != nil {
		return nil, err
	}

	tagRefs, err := getTags(r)
	if err != nil {
		return nil, err
	}
	defer tagRefs.Close()

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
	err := tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		if tagRef.Name().IsTag() && strings.HasPrefix(tagRef.Name().Short(), "v") {
			log.Debug("adding tag", "tag", tagRef.Name().String())
			versions = append(versions, tagRef.Name().String())
		}
		return nil
	})
	if err != nil {
		log.Error("Error iterating tags", "err", err)
		return nil
	}
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

// compareSuffixes compares two suffixes in semantic versions according to SemVer 2.0 spec.
// Returns true if suffix1 > suffix2 (for descending sort order).
func compareSuffixes(suffix1, suffix2 string) bool {
	// Per SemVer 2.0: stable version (no suffix) > any pre-release version
	if suffix1 == "" && suffix2 != "" {
		return true
	}
	if suffix1 != "" && suffix2 == "" {
		return false
	}

	// Both have suffixes - compare according to SemVer 2.0 rules
	// Strip leading dashes and split by dots
	ids1 := strings.Split(strings.TrimPrefix(suffix1, "-"), ".")
	ids2 := strings.Split(strings.TrimPrefix(suffix2, "-"), ".")

	// Compare identifiers left to right
	for i := 0; i < len(ids1) && i < len(ids2); i++ {
		id1 := ids1[i]
		id2 := ids2[i]

		// Check if identifiers are numeric
		num1, isNum1 := parseNumericIdentifier(id1)
		num2, isNum2 := parseNumericIdentifier(id2)

		if isNum1 && isNum2 {
			// Both numeric: compare numerically
			if num1 != num2 {
				return num1 > num2
			}
		} else if isNum1 && !isNum2 {
			// Numeric has lower precedence than alphanumeric
			return false
		} else if !isNum1 && isNum2 {
			// Alphanumeric has higher precedence than numeric
			return true
		} else {
			// Both alphanumeric: compare lexically
			if id1 != id2 {
				return id1 > id2
			}
		}
	}

	// All compared identifiers are equal; longer list has higher precedence
	return len(ids1) > len(ids2)
}

// parseNumericIdentifier checks if an identifier consists only of digits
// and returns its numeric value if so.
func parseNumericIdentifier(id string) (int, bool) {
	if id == "" {
		return 0, false
	}

	// Check if all characters are digits
	for _, ch := range id {
		if ch < '0' || ch > '9' {
			return 0, false
		}
	}

	// Parse as integer
	num, err := strconv.Atoi(id)
	if err != nil {
		return 0, false
	}

	return num, true
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

// CreateTag creates a new git tag with the given tag.
// Uses concurrency protection to prevent concurrent git operations.
func CreateTag(tag string) error {
	repoPath, err := findGitRepoRoot(".")
	if err != nil {
		return fmt.Errorf("failed to find git repository: %w", err)
	}

	return createTagWithLock(repoPath, tag)
}

// PushTag pushes the latest git tag to the remote repository.
// Uses concurrency protection to prevent concurrent git operations.
func PushTag() error {
	repoPath, err := findGitRepoRoot(".")
	if err != nil {
		return fmt.Errorf("failed to find git repository: %w", err)
	}

	return pushTagWithLock(repoPath)
}

// createTagWithLock creates a new git tag with the given tag using git operation locking.
func createTagWithLock(repoPath, tag string) error {
	lock, err := acquireGitLock(repoPath)
	if err != nil {
		return fmt.Errorf("failed to acquire git lock: %w", err)
	}
	defer func() {
		if releaseErr := lock.Release(); releaseErr != nil {
			log.Error("failed to release git lock", "err", releaseErr)
		}
	}()

	return createTag(tag)
}

// pushTagWithLock pushes tags to remote using git operation locking.
func pushTagWithLock(repoPath string) error {
	lock, err := acquireGitLock(repoPath)
	if err != nil {
		return fmt.Errorf("failed to acquire git lock: %w", err)
	}
	defer func() {
		if releaseErr := lock.Release(); releaseErr != nil {
			log.Error("failed to release git lock", "err", releaseErr)
		}
	}()

	return pushTag()
}

// createTag creates a new git tag with the given tag.
func createTag(tag string) error {
	cmdTag := execCommand("git", "tag", "-m", tag, tag)
	if output, err := cmdTag.CombinedOutput(); err != nil {
		log.Error("failed to create tag", "err", err, "output", string(output))
		return fmt.Errorf("failed to create tag: %w; %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// pushTag pushes the latest git tag to the remote repository.
func pushTag() error {
	cmdPush := execCommand("git", "push", "--tags")
	if output, err := cmdPush.CombinedOutput(); err != nil {
		log.Error("failed to push tag", "err", err, "output", string(output))
		return fmt.Errorf("failed to push tag: %w; %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// findGitRepoRoot finds the root directory of the git repository.
func findGitRepoRoot(startPath string) (string, error) {
	currentPath := startPath
	for {
		if _, err := os.Stat(filepath.Join(currentPath, ".git")); err == nil {
			return currentPath, nil
		}

		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			return "", fmt.Errorf("not inside a git repository")
		}
		currentPath = parentPath
	}
}

// GetDefaultPushPreference reads the [bump] defaultPush value from .git/config in the given repo path.
// Returns (value, isSet, error) where isSet indicates if the preference was explicitly configured.
func GetDefaultPushPreference(repoPath string) (bool, bool, error) {
	// Validate repository path
	if err := validateRepositoryPath(repoPath); err != nil {
		return false, false, fmt.Errorf("invalid repository path: %w", err)
	}

	configPath := filepath.Join(repoPath, ".git", "config")

	// Check if config file exists and is readable
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return false, false, fmt.Errorf("git config file not found: %s", configPath)
		}
		return false, false, fmt.Errorf("cannot access git config file: %w", err)
	}

	cfg, err := ini.Load(configPath)
	if err != nil {
		return false, false, fmt.Errorf("failed to load git config: %w", err)
	}

	section := cfg.Section("bump")
	if !section.HasKey("defaultPush") {
		// Return false, false (not set) when preference is not configured
		return false, false, nil
	}

	val := section.Key("defaultPush").String()
	switch val {
	case "true":
		return true, true, nil // value=true, isSet=true
	case "false":
		return false, true, nil // value=false, isSet=true (explicitly set to false)
	default:
		return false, false, fmt.Errorf("invalid defaultPush value: %s (must be 'true' or 'false')", val)
	}
}

// SetDefaultPushPreference writes the [bump] defaultPush value to .git/config in the given repo path.
// Uses atomic writes to prevent corruption.
func SetDefaultPushPreference(repoPath string, value bool) error {
	// Validate repository path
	if err := validateRepositoryPath(repoPath); err != nil {
		return fmt.Errorf("invalid repository path: %w", err)
	}

	configPath := filepath.Join(repoPath, ".git", "config")

	// Check if config file exists and is writable
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("git config file not found: %s", configPath)
		}
		return fmt.Errorf("cannot access git config file: %w", err)
	}

	// Create backup file path for atomic operation
	backupPath := configPath + ".bump.tmp"

	// Load current config
	cfg, err := ini.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load git config: %w", err)
	}

	// Update the configuration
	section := cfg.Section("bump")
	section.Key("defaultPush").SetValue(fmt.Sprintf("%v", value))

	// Write to temporary file first (atomic operation)
	if err := cfg.SaveTo(backupPath); err != nil {
		return fmt.Errorf("failed to write temporary config: %w", err)
	}

	// Atomic rename to replace original file
	if err := os.Rename(backupPath, configPath); err != nil {
		// Clean up temporary file on failure
		if rmErr := os.Remove(backupPath); rmErr != nil {
			log.Error("failed to clean up temporary config file", "backupPath", backupPath, "err", rmErr)
		}
		return fmt.Errorf("failed to update git config atomically: %w", err)
	}

	return nil
}

// validateRepositoryPath validates that the given path is a valid Git repository.
func validateRepositoryPath(repoPath string) error {
	if repoPath == "" {
		return fmt.Errorf("repository path cannot be empty")
	}

	// Clean and validate the path
	cleanPath := filepath.Clean(repoPath)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if it's a valid git repository
	gitDir := filepath.Join(absPath, ".git")
	stat, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not a git repository: %s", absPath)
		}
		return fmt.Errorf("cannot access .git directory: %w", err)
	}

	if !stat.IsDir() {
		return fmt.Errorf(".git is not a directory: %s", gitDir)
	}

	return nil
}
