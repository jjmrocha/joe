package toolpacks

import (
	"context"
	"errors"
	"testing"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPackName = "serena"

type recordingPack struct {
	name        string
	closed      *[]string
	instruction *mcp.Instruction
}

func (p recordingPack) Close() error {
	*p.closed = append(*p.closed, p.name)

	return nil
}

func (p recordingPack) Instructions(context.Context) *mcp.Instruction {
	return p.instruction
}

func TestSetAdd(t *testing.T) {
	t.Run("returns the error and keeps nothing when a pack fails to open", func(t *testing.T) {
		// given
		var closed []string
		var set Set
		failure := errors.New("serena missing")
		// when
		result := set.Add(recordingPack{name: testPackName, closed: &closed}, failure)
		// then
		assert.ErrorIs(t, result, failure)
		set.Close()
		assert.Empty(t, closed)
	})

	t.Run("keeps a pack that opened", func(t *testing.T) {
		// given
		var closed []string
		var set Set
		// when
		result := set.Add(recordingPack{name: testPackName, closed: &closed}, nil)
		// then
		require.NoError(t, result)
		set.Close()
		assert.Equal(t, []string{testPackName}, closed)
	})
}

func TestSetClose(t *testing.T) {
	t.Run("closes the packs in reverse order of opening", func(t *testing.T) {
		// given
		var result []string
		var set Set
		require.NoError(t, set.Add(recordingPack{name: "first", closed: &result}, nil))
		require.NoError(t, set.Add(recordingPack{name: "second", closed: &result}, nil))
		require.NoError(t, set.Add(recordingPack{name: "third", closed: &result}, nil))
		expected := []string{"third", "second", "first"}
		// when
		set.Close()
		// then
		assert.Equal(t, expected, result)
	})
}

func TestSetInstructions(t *testing.T) {
	t.Run("collects the instructions of every pack that has some, in order", func(t *testing.T) {
		// given
		var closed []string
		var set Set
		date := mcp.Instruction{Name: "date", Text: "Ask for the date."}
		shell := mcp.Instruction{Name: "shell", Text: "Run commands."}
		require.NoError(t, set.Add(recordingPack{name: "date", closed: &closed, instruction: &date}, nil))
		require.NoError(t, set.Add(recordingPack{name: "silent", closed: &closed}, nil))
		require.NoError(t, set.Add(recordingPack{name: "shell", closed: &closed, instruction: &shell}, nil))
		expected := []mcp.Instruction{date, shell}
		// when
		result := set.Instructions(t.Context())
		// then
		assert.Equal(t, expected, result)
	})

	t.Run("returns nothing when no pack has instructions", func(t *testing.T) {
		// given
		var closed []string
		var set Set
		require.NoError(t, set.Add(recordingPack{name: "silent", closed: &closed}, nil))
		// when
		result := set.Instructions(t.Context())
		// then
		assert.Empty(t, result)
	})
}
