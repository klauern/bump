package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/klauern/bump"
)

// BumpService coordinates version bumping operations using dependency injection.
// This service layer separates business logic from I/O, making it fully testable.
type BumpService struct {
	repo    GitRepository
	updater *VersionFileUpdater
	output  io.Writer
}

// NewBumpService creates a new BumpService with the given dependencies.
func NewBumpService(repo GitRepository, updater *VersionFileUpdater, output io.Writer) *BumpService {
	if output == nil {
		output = os.Stdout
	}
	if updater == nil {
		updater = NewVersionFileUpdater()
	}

	return &BumpService{
		repo:    repo,
		updater: updater,
		output:  output,
	}
}

// BumpOptions contains all options for a version bump operation.
type BumpOptions struct {
	BumpType   string // "patch", "minor", or "major"
	Suffix     string // Optional pre-release suffix (e.g., "beta", "rc1")
	UpdateFile string // Optional path to file containing Version constant
	Push       bool   // Whether to push tags to remote
	DryRun     bool   // Preview changes without making them
}

// BumpResult contains the result of a bump operation.
type BumpResult struct {
	NextTag      string // The tag that was (or would be) created
	Pushed       bool   // Whether the tag was pushed to remote
	FileUpdated  bool   // Whether a file was updated
	WouldPush    bool   // Dry-run: whether tag would be pushed
	WouldUpdate  bool   // Dry-run: whether file would be updated
	PreviousTag  string // The previous latest tag (empty if none)
}

// Bump performs a version bump operation.
// This is the main entry point for the service layer.
func (s *BumpService) Bump(opts BumpOptions) (*BumpResult, error) {
	// Get all tags from the repository
	tagRefs, err := s.repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}
	defer tagRefs.Close()

	// Find the latest tag
	latestTag, err := bump.GetLatestTag(tagRefs)
	if err != nil {
		return nil, fmt.Errorf("failed to determine latest tag: %w", err)
	}

	// Calculate the next version (pure function)
	nextTag, err := calculateNextVersion(latestTag, opts.BumpType, opts.Suffix)
	if err != nil {
		return nil, fmt.Errorf("failed to determine next tag: %w", err)
	}

	// Print starting message if no tags exist
	if latestTag == "" {
		if opts.DryRun {
			fmt.Fprintln(s.output, "No tags found, would start at v0.1.0")
		} else {
			fmt.Fprintln(s.output, "No tags found, starting at v0.1.0")
		}
	}

	// Dry-run mode: preview without making changes
	if opts.DryRun {
		fmt.Fprint(s.output, formatDryRunMessage(nextTag, opts.Push, opts.UpdateFile))
		return &BumpResult{
			NextTag:      nextTag,
			WouldPush:    opts.Push,
			WouldUpdate:  opts.UpdateFile != "",
			PreviousTag:  latestTag,
		}, nil
	}

	// Create the tag
	if err := s.repo.CreateTag(nextTag); err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	// Push tags if requested
	pushed := false
	if opts.Push {
		if err := s.repo.PushTags(); err != nil {
			return nil, fmt.Errorf("failed to push tags: %w", err)
		}
		pushed = true
	}

	// Print success message (pure function)
	fmt.Fprintln(s.output, formatBumpMessage(nextTag, pushed))

	// Update version file if requested
	fileUpdated := false
	if opts.UpdateFile != "" {
		if err := s.UpdateVersionFile(opts.UpdateFile, nextTag); err != nil {
			return nil, fmt.Errorf("failed to update file: %w", err)
		}
		fileUpdated = true
	}

	return &BumpResult{
		NextTag:      nextTag,
		Pushed:       pushed,
		FileUpdated:  fileUpdated,
		PreviousTag:  latestTag,
	}, nil
}

// UpdateVersionFile updates a Go source file with a new development version.
// This method handles path validation, file operations, and git operations.
func (s *BumpService) UpdateVersionFile(filePath, nextTag string) error {
	repoPath := s.repo.Path()

	// Validate file path to prevent security issues
	if err := validateFilePath(filePath, repoPath); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Clean the path
	cleanPath := filepath.Clean(filePath)

	// Resolve to absolute path for file operations
	absPath := filepath.Join(repoPath, cleanPath)

	// Calculate development version (pure function)
	devVersion, err := calculateDevVersion(nextTag)
	if err != nil {
		return fmt.Errorf("failed to calculate dev version: %w", err)
	}

	// Parse, update, and write the file (using absolute path)
	node, fset, err := s.updater.ParseGoFile(absPath)
	if err != nil {
		return err
	}

	if err := s.updater.UpdateVersionConstant(node, devVersion); err != nil {
		return err
	}

	if err := s.updater.WriteFormattedFile(absPath, fset, node); err != nil {
		return err
	}

	// Stage and commit the file
	worktree, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get working tree: %w", err)
	}

	// Get relative path for git operations
	relPath, err := filepath.Rel(repoPath, absPath)
	if err != nil {
		return fmt.Errorf("failed to determine relative path: %w", err)
	}

	// Stage the file
	if _, err := worktree.Add(relPath); err != nil {
		return fmt.Errorf("failed to stage file: %w", err)
	}

	// Commit the change
	commitMsg := fmt.Sprintf("Bump version to %s", devVersion)
	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Bump CLI",
			Email: "bump@localhost",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit file: %w", err)
	}

	return nil
}
