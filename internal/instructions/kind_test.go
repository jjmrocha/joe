package instructions

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKinds(t *testing.T) {
	t.Run("holds every kind joe accepts", func(t *testing.T) {
		// given
		expected := []Kind{KindAgents, KindClaude}
		// when
		result := slices.Sorted(Kinds.Values())
		// then
		assert.Equal(t, expected, result)
	})

	t.Run("agrees with what Validate accepts", func(t *testing.T) {
		for kind := range Kinds.Values() {
			t.Run(string(kind), func(t *testing.T) {
				// given
				value := kind
				// when
				err := value.Validate()
				// then
				assert.NoError(t, err)
			})
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("accepts the kinds joe knows", func(t *testing.T) {
		testCases := []struct {
			name  string
			value Kind
		}{
			{name: "claude", value: KindClaude},
			{name: "agents", value: KindAgents},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				value := testCase.value
				// when
				err := value.Validate()
				// then
				assert.NoError(t, err)
			})
		}
	})

	t.Run("rejects any other value", func(t *testing.T) {
		testCases := []struct {
			name  string
			value Kind
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
				err := value.Validate()
				// then
				assert.ErrorIs(t, err, ErrInvalidKind)
			})
		}
	})
}
