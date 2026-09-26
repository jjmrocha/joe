package guard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstraints(t *testing.T) {
	t.Run("lets files change in the repository and the knowledge base", func(t *testing.T) {
		// given
		expected := []string{
			"The agent may only create, modify or delete files inside /src/joe.",
			"The agent may also create, modify or delete files inside /srv/wiki.",
			"The agent must not change remote or shared state: push, deploy, publish, merge, or send messages.",
			"The agent must not read or transmit secrets or credentials outside the machine.",
		}
		// when
		result := Constraints(testRepoPath, testKBPath)
		// then
		assert.Equal(t, expected, result)
	})

	t.Run("names no knowledge base when none is configured", func(t *testing.T) {
		// given
		expected := []string{
			"The agent may only create, modify or delete files inside /src/joe.",
			"The agent must not change remote or shared state: push, deploy, publish, merge, or send messages.",
			"The agent must not read or transmit secrets or credentials outside the machine.",
		}
		// when
		result := Constraints(testRepoPath, "")
		// then
		assert.Equal(t, expected, result)
	})
}
