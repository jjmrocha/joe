package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jjmrocha/joe/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKeyEnv = "JOE_TEST_KEY"

func testProfile(skills string) string {
	return `{
  "harness": "claude",
  "llm": {
    "provider": "openrouter",
    "api-key-env": "` + testKeyEnv + `",
    "model": "z-ai/glm-5.3-flash",
    "effort": "medium"
  },
  "skills": ` + skills + `,
  "mcps": {},
  "mcps-on": []
}`
}

func testConfig(t *testing.T, profile string) *config.Config {
	t.Helper()
	t.Setenv(testKeyEnv, "sk-test")

	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	dir := filepath.Join(base, "joe")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("MkdirAll(%s): %v", dir, err)
	}

	path := filepath.Join(dir, "local.json")
	if err := os.WriteFile(path, []byte(profile), 0o600); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}

	cfg, err := config.Load("local")
	require.NoError(t, err)

	return cfg
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

func TestNewSkillCollection(t *testing.T) {
	t.Run("reports every missing skill in one pass", func(t *testing.T) {
		// given
		cfg := testConfig(t, testProfile(`[]`))
		// when
		_, err := newSkillCollection(cfg)
		// then
		require.Error(t, err)

		for _, expected := range skillNames {
			assert.Contains(t, err.Error(), expected)
		}
	})

	t.Run("names a missing extra skill beside the missing built-ins", func(t *testing.T) {
		// given
		cfg := testConfig(t, testProfile(`["removing-ai-tells"]`))
		// when
		_, err := newSkillCollection(cfg)
		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "removing-ai-tells")
	})

	t.Run("loads every skill the profile names", func(t *testing.T) {
		// given
		cfg := testConfig(t, testProfile(`["removing-ai-tells"]`))

		codingSkillsDir, err := config.CodingSkillsDir()
		require.NoError(t, err)

		skillsDir, err := config.SkillsDir()
		require.NoError(t, err)

		for _, name := range skillNames {
			writeSkill(t, codingSkillsDir, name)
		}

		writeSkill(t, skillsDir, "removing-ai-tells")
		// when
		result, err := newSkillCollection(cfg)
		// then
		require.NoError(t, err)

		catalog := result.Catalog()
		for _, name := range append(skillNames, "removing-ai-tells") {
			assert.Contains(t, catalog, "<name>"+name+"</name>")
		}
	})
	t.Run("rejects an extra skill that is one of joe's own", func(t *testing.T) {
		// given
		cfg := testConfig(t, testProfile(`["brainstorm"]`))

		codingSkillsDir, err := config.CodingSkillsDir()
		require.NoError(t, err)

		skillsDir, err := config.SkillsDir()
		require.NoError(t, err)

		for _, name := range skillNames {
			writeSkill(t, codingSkillsDir, name)
		}

		writeSkill(t, skillsDir, "brainstorm")
		// when
		_, err = newSkillCollection(cfg)
		// then
		require.ErrorIs(t, err, ErrReservedSkill)
		assert.Contains(t, err.Error(), "brainstorm")
	})
}
