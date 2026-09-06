# joe

The Opinionated Coding Agent

Named after [Joe Armstrong](https://en.wikipedia.org/wiki/Joe_Armstrong_(programmer)), creator of Erlang.

joe is a terminal coding agent. It reads and changes real code bases through
[Serena](https://github.com/oraios/serena)'s symbolic tools, and it works through skills —
loading the one that matches the request before it touches anything.

Built on [ai-toolkit](https://github.com/jjmrocha/ai-toolkit) (agent, LLM client, tools,
skills) and [ai-chat](https://github.com/jjmrocha/ai-chat) (chat core, TUI, slash commands).

## Prerequisites

| What | Why | How |
|---|---|---|
| Go 1.27+ | Building joe | [go.dev/dl](https://go.dev/dl/) |
| `OPEN_ROUTER_KEY` | joe talks to [OpenRouter](https://openrouter.ai) | `export OPEN_ROUTER_KEY=sk-...` |
| `uvx` on `PATH` | Starts Serena, which serves joe's coding tools | [uv](https://github.com/astral-sh/uv) |
| Skills in `~/.claude/skills` | joe loads a skill before acting | see below |

joe loads eleven skills by name from `~/.claude/skills`: `analyze-code`, `brainstorm`,
`coding-discipline`, `designing-interfaces`, `guiding-manual-testing`, `knowledge-base`,
`research`, `style-checker`, `test-driven-development`, `using-software-specialists`, and
`writing-unit-tests`. They live in
[jjmrocha/coding-skills](https://github.com/jjmrocha/coding-skills) — copy the folders into
`~/.claude/skills/`. A missing skill stops joe at startup.

## Build and run

```bash
make build          # writes ./bin/joe
export OPEN_ROUTER_KEY=sk-...
./bin/joe
```

Run it from the repository you want joe to work on — it calls `repo_info` at the start of a
session to find the git root, and activates Serena on that path.

`go install github.com/jjmrocha/joe/cmd@latest` also works, but installs a binary named
`cmd`, after its directory. `make build` names it `joe`.

## Reading other repositories

A feature that spans repositories is still written in the one joe was started in. joe reads
the others and cannot change them: it reaches them through Serena's `query_project`, which
refuses every editing tool.

`query_project` only reaches repositories Serena has registered. Register each one once:

```bash
uvx --from git+https://github.com/oraios/serena serena project create /path/to/repo
```

Reading and searching files needs nothing more. Symbol lookups in another repository go
through Serena's project server, which joe does not start — run it alongside joe if you
want them:

```bash
uvx --from git+https://github.com/oraios/serena serena start-project-server
```

Without it, joe falls back to searching the other repository as text.

## Make targets

`make` on its own lists them:

| Target | Effect |
|---|---|
| `build` | Build joe into `./bin` |
| `clean` | Remove `./bin` |
| `test` | Run all tests |
| `bench` | Run benchmarks |
| `lint` | Run golangci-lint |
| `deps` | Update dependencies |
| `tidy` | Tidy `go.mod` |

## License

[MIT](LICENSE) © 2026 Joaquim Rocha
