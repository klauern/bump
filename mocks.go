package bump

import (
	"io"

	"github.com/go-git/go-git/v5/plumbing"
)

type MockReferenceIter struct {
	refs []plumbing.Reference
	idx  int
}

func NewMockReferenceIter(refs []plumbing.Reference) *MockReferenceIter {
	return &MockReferenceIter{
		refs: refs,
		idx:  -1,
	}
}

func (m *MockReferenceIter) Next() (*plumbing.Reference, error) {
	m.idx++
	if m.idx >= len(m.refs) {
		return nil, io.EOF
	}
	return &m.refs[m.idx], nil
}

func (m *MockReferenceIter) ForEach(cb func(*plumbing.Reference) error) error {
	for _, ref := range m.refs {
		if err := cb(&ref); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockReferenceIter) Close() {
	m.idx = -1
}
