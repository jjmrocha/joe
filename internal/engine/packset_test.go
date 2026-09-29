package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingPack struct {
	name   string
	closed *[]string
}

func (p recordingPack) Close() error {
	*p.closed = append(*p.closed, p.name)

	return nil
}

func (recordingPack) Instructions(context.Context) (*mcp.Instruction, error) { return nil, nil }

func TestPackSet(t *testing.T) {
	t.Run("keeps each pack that opened", func(t *testing.T) {
		// given
		var closed []string
		var result packSet
		// when
		errFirst := result.add(recordingPack{name: "date", closed: &closed}, nil)
		errSecond := result.add(recordingPack{name: "shell", closed: &closed}, nil)
		// then
		require.NoError(t, errFirst)
		require.NoError(t, errSecond)
		assert.Len(t, result, 2)
	})

	t.Run("returns the error and keeps nothing when a pack fails to open", func(t *testing.T) {
		// given
		expected := errors.New("serena missing")
		var set packSet
		// when
		result := set.add(nil, expected)
		// then
		assert.ErrorIs(t, result, expected)
		assert.Empty(t, set)
	})

	t.Run("closes the packs in reverse order of opening", func(t *testing.T) {
		// given
		var result []string
		var set packSet
		require.NoError(t, set.add(recordingPack{name: "first", closed: &result}, nil))
		require.NoError(t, set.add(recordingPack{name: "second", closed: &result}, nil))
		require.NoError(t, set.add(recordingPack{name: "third", closed: &result}, nil))
		expected := []string{"third", "second", "first"}
		// when
		set.close()
		// then
		assert.Equal(t, expected, result)
	})
}
