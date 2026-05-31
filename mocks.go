package bump

import (
	"io"

	"github.com/go-git/go-git/v5/plumbing"
)

// MockReferenceIter is a simple in-memory iterator over git references.
//
// It is primarily used in tests to provide a `go-git`-compatible reference iterator.
type MockReferenceIter struct {
	refs []plumbing.Reference
	idx  int
}

// NewMockReferenceIter constructs a [MockReferenceIter] over the provided references.
//
// This is primarily intended for tests that need a [storer.ReferenceIter]-like
// implementation.
func NewMockReferenceIter(refs []plumbing.Reference) *MockReferenceIter {
	return &MockReferenceIter{
		refs: refs,
		idx:  -1,
	}
}

// Next returns the next reference or io.EOF when exhausted.
func (m *MockReferenceIter) Next() (*plumbing.Reference, error) {
	m.idx++
	if m.idx >= len(m.refs) {
		return nil, io.EOF
	}
	return &m.refs[m.idx], nil
}

// ForEach calls cb for each reference until cb returns an error.
func (m *MockReferenceIter) ForEach(cb func(*plumbing.Reference) error) error {
	for _, ref := range m.refs {
		if err := cb(&ref); err != nil {
			return err
		}
	}
	return nil
}

// Close resets internal iteration state.
func (m *MockReferenceIter) Close() {
	m.idx = -1
}
