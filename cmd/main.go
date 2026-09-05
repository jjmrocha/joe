// Command joe launches an opinionated coding agent.
//
// It connects to OpenRouter, reading the API key from the OPEN_ROUTER_KEY
// environment variable, and takes its coding tools from Serena, which needs the
// uvx executable on PATH. Its skills are read from the user's Claude skills
// folder, ~/.claude/skills.
//
// Usage:
//
//	export OPEN_ROUTER_KEY=sk-...
//	joe
package main

import (
	"context"
	"log"

	"github.com/jjmrocha/joe/internal/agent"
)

func main() {
	if err := agent.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
