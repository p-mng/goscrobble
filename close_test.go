package main_test

import (
	"testing"

	main "github.com/p-mng/goscrobble"
	"github.com/stretchr/testify/require"
)

type FakeCloser struct {
	Called bool
}

func (f *FakeCloser) Close() error {
	f.Called = true
	return nil
}

func TestCloseLogged(t *testing.T) {
	closer := &FakeCloser{}
	require.False(t, closer.Called)

	main.CloseLogged(closer)
	require.True(t, closer.Called)
}
