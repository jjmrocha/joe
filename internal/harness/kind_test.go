package harness

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKind(t *testing.T) {
	t.Run("accepts the kinds joe knows", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    string
			expected Kind
		}{
			{name: "claude", value: "claude", expected: KindClaude},
			{name: "agents", value: "agents", expected: KindAgents},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				value := testCase.value
				// when
				result, err := ParseKind(value)
				// then
				require.NoError(t, err)
				assert.Equal(t, testCase.expected, result)
			})
		}
	})

	t.Run("rejects any other value", func(t *testing.T) {
		testCases := []struct {
			name  string
			value string
		}{
			{name: "empty", value: ""},
			{name: "unknown", value: "codex"},
			{name: "wrong case", value: "Claude"},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				value := testCase.value
				// when
				_, err := ParseKind(value)
				// then
				assert.ErrorIs(t, err, ErrInvalidKind)
			})
		}
	})
}
