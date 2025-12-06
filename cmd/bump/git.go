package main

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/klauern/bump"
)

// GitRepository defines the interface for git repository operations.
// This abstraction allows for testing with mocks instead of real git repos.
type GitRepository interface {
	// Tags returns an iterator over all tags in the repository
	Tags() (storer.ReferenceIter, error)

	// CreateTag creates a new annotated tag at HEAD
	CreateTag(name string) error

	// PushTags pushes all tags to the remote repository
	PushTags() error

	// Worktree returns the working tree for this repository
	Worktree() (GitWorktree, error)

	// Path returns the filesystem path to the repository
	Path() string
}

// GitWorktree defines the interface for git working tree operations.
// This abstraction allows for testing file staging and commits with mocks.
type GitWorktree interface {
	// Add stages a file for commit
	Add(path string) (plumbing.Hash, error)

	// Commit creates a new commit with the staged changes
	Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error)
}

// GoGitRepository is the real implementation of GitRepository using go-git.
type GoGitRepository struct {
	repo *git.Repository
	path string
}

// NewGoGitRepository creates a new GoGitRepository by opening an existing git repo.
func NewGoGitRepository(repoPath string) (*GoGitRepository, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &GoGitRepository{
		repo: repo,
		path: repoPath,
	}, nil
}

// Tags returns an iterator over all tags in the repository.
func (r *GoGitRepository) Tags() (storer.ReferenceIter, error) {
	return r.repo.Tags()
}

// CreateTag creates a new annotated tag at HEAD using the bump package.
func (r *GoGitRepository) CreateTag(name string) error {
	return bump.CreateTag(name)
}

// PushTags pushes all tags to the remote repository using the bump package.
func (r *GoGitRepository) PushTags() error {
	return bump.PushTag()
}

// Worktree returns the working tree for this repository.
func (r *GoGitRepository) Worktree() (GitWorktree, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get working tree: %w", err)
	}
	return &GoGitWorktree{worktree: wt}, nil
}

// Path returns the filesystem path to the repository.
func (r *GoGitRepository) Path() string {
	return r.path
}

// GoGitWorktree is the real implementation of GitWorktree using go-git.
type GoGitWorktree struct {
	worktree *git.Worktree
}

// Add stages a file for commit.
func (w *GoGitWorktree) Add(path string) (plumbing.Hash, error) {
	return w.worktree.Add(path)
}

// Commit creates a new commit with the staged changes.
func (w *GoGitWorktree) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	// If no options provided, use default author
	if opts == nil {
		opts = &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Bump CLI",
				Email: "bump@localhost",
				When:  time.Now(),
			},
		}
	}
	return w.worktree.Commit(msg, opts)
}
