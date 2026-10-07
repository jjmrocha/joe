package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jjmrocha/joe/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func configHome(t *testing.T) {
	t.Helper()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
}

func writeSkill(t *testing.T, dir, name string) {
	t.Helper()

	folder := filepath.Join(dir, name)
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatalf("MkdirAll(%s): %v", folder, err)
	}

	content := "---\nname: " + name + "\ndescription: " + name + " skill\n---\n\nBody.\n"
	if err := os.WriteFile(filepath.Join(folder, "SKILL.md"), []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%s): %v", folder, err)
	}
}

func TestLoad(t *testing.T) {
	t.Run("reports every missing skill in one pass", func(t *testing.T) {
		// given
		configHome(t)
		// when
		_, err := Load(nil)
		// then
		require.Error(t, err)

		for _, expected := range skillNames {
			assert.Contains(t, err.Error(), expected)
		}
	})

	t.Run("names a missing extra skill beside the missing built-ins", func(t *testing.T) {
		// given
		configHome(t)
		// when
		_, err := Load([]string{"removing-ai-tells"})
		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "removing-ai-tells")
	})

	t.Run("loads every skill the profile names", func(t *testing.T) {
		// given
		configHome(t)

		codingSkillsDir, err := config.CodingSkillsDir()
		require.NoError(t, err)

		skillsDir, err := config.SkillsDir()
		require.NoError(t, err)

		for _, name := range skillNames {
			writeSkill(t, codingSkillsDir, name)
		}

		writeSkill(t, skillsDir, "removing-ai-tells")
		// when
		result, err := Load([]string{"removing-ai-tells"})
		// then
		require.NoError(t, err)

		catalog := result.Catalog()
		for _, name := range append(skillNames, "removing-ai-tells") {
			assert.Contains(t, catalog, "<name>"+name+"</name>")
		}
	})

	t.Run("rejects an extra skill that is one of joe's own", func(t *testing.T) {
		// given
		configHome(t)

		codingSkillsDir, err := config.CodingSkillsDir()
		require.NoError(t, err)

		skillsDir, err := config.SkillsDir()
		require.NoError(t, err)

		for _, name := range skillNames {
			writeSkill(t, codingSkillsDir, name)
		}

		writeSkill(t, skillsDir, "brainstorm")
		// when
		_, err = Load([]string{"brainstorm"})
		// then
		require.ErrorIs(t, err, ErrReservedSkill)
		assert.Contains(t, err.Error(), "brainstorm")
	})
}
