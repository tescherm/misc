package shuffle

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShuffler(t *testing.T) {
	t.Parallel()

	s, err := NewShuffler("/tmp/in", "/tmp/out")
	require.NoError(t, err)

	t.Cleanup(func() {
		s.Close()
	})

	err = s.FirstPass()
	require.NoError(t, err)
	err = s.SecondPass()
	require.NoError(t, err)
}
