package config

import (
	"testing"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/stretchr/testify/assert"
)

func TestProviders(t *testing.T) {
	t.Run("holds every provider joe accepts", func(t *testing.T) {
		// given
		expected := []llm.Provider{llm.ProviderAnthropic, llm.ProviderOllama, llm.ProviderOpenRouter}
		// when
		result := Providers()
		// then
		assert.Equal(t, expected, result)
	})

	t.Run("hands out a copy the caller may change", func(t *testing.T) {
		// given
		providers := Providers()
		providers[0] = "bogus"
		expected := []llm.Provider{llm.ProviderAnthropic, llm.ProviderOllama, llm.ProviderOpenRouter}
		// when
		result := Providers()
		// then
		assert.Equal(t, expected, result)
	})
}
