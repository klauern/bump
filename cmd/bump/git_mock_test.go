package main

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// MockGitRepository is a mock implementation of GitRepository for testing.
type MockGitRepository struct {
	TagsFunc      func() (storer.ReferenceIter, error)
	CreateTagFunc func(string) error
	PushTagsFunc  func() error
	WorktreeFunc  func() (GitWorktree, error)
	PathFunc      func() string
}

// Tags calls the mock function if set, otherwise returns nil.
func (m *MockGitRepository) Tags() (storer.ReferenceIter, error) {
	if m.TagsFunc != nil {
		return m.TagsFunc()
	}
	return NewMockTagIterator([]string{}), nil
}

// CreateTag calls the mock function if set, otherwise returns nil.
func (m *MockGitRepository) CreateTag(name string) error {
	if m.CreateTagFunc != nil {
		return m.CreateTagFunc(name)
	}
	return nil
}

// PushTags calls the mock function if set, otherwise returns nil.
func (m *MockGitRepository) PushTags() error {
	if m.PushTagsFunc != nil {
		return m.PushTagsFunc()
	}
	return nil
}

// Worktree calls the mock function if set, otherwise returns a mock worktree.
func (m *MockGitRepository) Worktree() (GitWorktree, error) {
	if m.WorktreeFunc != nil {
		return m.WorktreeFunc()
	}
	return &MockGitWorktree{}, nil
}

// Path calls the mock function if set, otherwise returns "/mock/repo".
func (m *MockGitRepository) Path() string {
	if m.PathFunc != nil {
		return m.PathFunc()
	}
	return "/mock/repo"
}

// MockGitWorktree is a mock implementation of GitWorktree for testing.
type MockGitWorktree struct {
	AddFunc    func(string) (plumbing.Hash, error)
	CommitFunc func(string, *git.CommitOptions) (plumbing.Hash, error)
}

// Add calls the mock function if set, otherwise returns a zero hash.
func (m *MockGitWorktree) Add(path string) (plumbing.Hash, error) {
	if m.AddFunc != nil {
		return m.AddFunc(path)
	}
	return plumbing.ZeroHash, nil
}

// Commit calls the mock function if set, otherwise returns a zero hash.
func (m *MockGitWorktree) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	if m.CommitFunc != nil {
		return m.CommitFunc(msg, opts)
	}
	return plumbing.ZeroHash, nil
}

// MockTagIterator is a mock implementation of storer.ReferenceIter for testing.
type MockTagIterator struct {
	tags  []string
	index int
}

// NewMockTagIterator creates a new MockTagIterator with the given tag names.
func NewMockTagIterator(tags []string) *MockTagIterator {
	return &MockTagIterator{
		tags:  tags,
		index: 0,
	}
}

// Next returns the next tag reference or io.EOF when done.
func (m *MockTagIterator) Next() (*plumbing.Reference, error) {
	if m.index >= len(m.tags) {
		return nil, fmt.Errorf("EOF")
	}

	tag := m.tags[m.index]
	m.index++

	// Create a mock reference
	ref := plumbing.NewHashReference(
		plumbing.ReferenceName("refs/tags/"+tag),
		plumbing.ZeroHash,
	)

	return ref, nil
}

// ForEach iterates over all tags and calls the function for each.
func (m *MockTagIterator) ForEach(fn func(*plumbing.Reference) error) error {
	for m.index < len(m.tags) {
		ref, err := m.Next()
		if err != nil {
			break
		}
		if err := fn(ref); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the iterator.
func (m *MockTagIterator) Close() {
	m.index = len(m.tags)
}

// Helper functions for common mock scenarios

// NewMockRepoWithTags creates a mock repository that returns the specified tags.
func NewMockRepoWithTags(tags []string) *MockGitRepository {
	return &MockGitRepository{
		TagsFunc: func() (storer.ReferenceIter, error) {
			return NewMockTagIterator(tags), nil
		},
	}
}

// NewMockRepoWithError creates a mock repository that returns errors.
func NewMockRepoWithError(tagsErr, createErr, pushErr error) *MockGitRepository {
	return &MockGitRepository{
		TagsFunc: func() (storer.ReferenceIter, error) {
			if tagsErr != nil {
				return nil, tagsErr
			}
			return NewMockTagIterator([]string{}), nil
		},
		CreateTagFunc: func(string) error {
			return createErr
		},
		PushTagsFunc: func() error {
			return pushErr
		},
	}
}

// NewMockWorktreeWithError creates a mock worktree that returns errors.
func NewMockWorktreeWithError(addErr, commitErr error) *MockGitWorktree {
	return &MockGitWorktree{
		AddFunc: func(string) (plumbing.Hash, error) {
			if addErr != nil {
				return plumbing.ZeroHash, addErr
			}
			return plumbing.ZeroHash, nil
		},
		CommitFunc: func(string, *git.CommitOptions) (plumbing.Hash, error) {
			if commitErr != nil {
				return plumbing.ZeroHash, commitErr
			}
			return plumbing.ZeroHash, nil
		},
	}
}
