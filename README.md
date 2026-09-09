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
| `npx` on `PATH` | Starts the context7 MCP, which serves library documentation | [Node.js](https://nodejs.org) |
| `donsetch` on `PATH` | Starts the DonSeTch MCP, which serves web search and fetching | [donsetch](https://github.com/dondai44423/donsetch) |

The last two are needed only on demand. Serena starts with joe and a missing `uvx` fails at
launch; the two MCP servers start the first time you use them, so joe runs without `npx` or
`donsetch` and a missing binary surfaces at that point instead.

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

## Your instructions

At startup joe reads the CLAUDE files you already keep for Claude Code and puts each one
into its prompt verbatim, in this order, skipping any that are absent:

| File | Scope |
|---|---|
| `~/.claude/CLAUDE.md` | You, everywhere |
| `<repo>/CLAUDE.md` | The repository joe was started in |
| `<repo>/CLAUDE.local.md` | That repository, not checked in |

A later file wins where two disagree, and joe's own instructions win over all of them on
tools, skills, Serena and the knowledge base — the rest is yours.

Imports are **not** followed. A line like `@RTK.md` is passed through as text; joe never
opens the file it names.

## The knowledge base

If any of those files sets `kb_path`, joe registers a second set of tools — `file_read`,
`file_write`, `file_edit`, `file_list`, `file_delete`, `file_workdir` — rooted at that
folder and unable to leave it, and tells the model that the knowledge base is theirs while
the repository stays Serena's.

Both spellings are accepted, anywhere in the file:

```
kb_path=/Users/you/Documents/LLM_WIKI
kb_path: /Users/you/Documents/LLM_WIKI
```

The path must be absolute and spelled out in full — `~` is not expanded, so `kb_path=~/wiki`
is rejected. A relative path, a `~`-relative one, or a folder that cannot be opened stops joe
at startup rather than running without the knowledge base. Without `kb_path` joe starts
normally and the `file_*` tools are simply absent.

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
