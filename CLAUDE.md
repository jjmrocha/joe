# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

joe is an application, not a library: `cmd/main.go` is the only entry point, and everything
else lives under `internal/`. The chat core, TUI and slash commands come from
`github.com/jjmrocha/ai-chat`; the agent, LLM client, tool packs and skills come from
`github.com/jjmrocha/ai-toolkit`. Both are used unchanged — a change that belongs in either
of them is made there, not worked around here.

## Commands

- `make build` writes `./bin/joe`. `go build ./cmd` would name the binary `cmd`, after its directory.
- `make test` runs `go test ./...` — but CI runs `go test -race ./...`. Use the `-race` form before pushing.
- `make lint` runs `golangci-lint run`. The config in `.golangci.yml` is strict (`errcheck`, `gosec`, `errorlint`, `gocritic`, `revive`, `govet` with `shadow`, `modernize`) and reports every issue (`max-issues-per-linter: 0`). Lint must be clean.
- **golangci-lint must be v2.13.2 or newer.** Earlier releases are built with Go 1.26 and refuse this module outright: `the Go language version (go1.26) used to build golangci-lint is lower than the targeted Go version (1.27.1)`. The same applies to `govulncheck` — rebuild it with the local toolchain if it reports `requires newer Go version`.
- `make bench` runs benchmarks; `make tidy` runs `go mod tidy`.

## Dependencies

- `github.com/stretchr/testify` is in `go.mod` with no importer, so **`make tidy` removes it**. That is expected, not a mistake: it is there so the first test written has its dependency already chosen. Put it back with `go get github.com/stretchr/testify@v1.11.1` after tidying, until a test imports it.
- `golang.org/x/text` and `github.com/yuin/goldmark` are pinned above the versions ai-chat resolves, because govulncheck reports the older ones as reachable through the TUI's markdown rendering (GO-2026-5970, GO-2026-5320). Do not lower them. When ai-chat bumps its own, the pins here become redundant and can go.
- Prefer helpers from `github.com/jjmrocha/go-algo` (`fn.Map`, `sets`, `future`) over hand-rolled loops where they fit. It is not a direct dependency yet — add it when code actually needs it.

## Documentation

- Every exported symbol gets a doc comment. Reference related symbols with doc links: `[Run]`, `[Register]`.
- The package comment lives in that package's `doc.go`, which holds nothing else.
- Unexported symbols get no doc comments. Test functions get no comments either (beyond the given/when/then markers below).
- `README.md` documents the prerequisites and how to build and run joe. Anything that changes what an operator must install or set goes there in the same change.

## Conventions

- Sentinel errors live in a per-package `errors.go` as `var ErrX = errors.New("lower case message")`, each with a doc comment naming the function that returns it.
- `internal/agent` owns the whole lifecycle — model, skills, Serena, session, chat core, UI — and `Run` is its only exported symbol. `cmd/main.go` stays a call to it plus `log.Fatal`. Do not return the chat core to `main`.
- ai-toolkit's `agent` and `tools` packages collide by name with this repo's own; the imports are aliased (`toolkitagent`, `joetools`) at the point of use.

## Tests

- `testify` `assert`/`require`, in the same package as the code under test (no `_test` package suffix).
- One `Test<Func>` per exported function, with `t.Run` subtests for cases; table-driven for families of error cases.
- Mark sections with `// given`, `// when`, `// then` comments.
- Name the value under test `result` and the comparison value `expected`.
