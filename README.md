# joe

**The Opinionated Coding Agent**

[![CI](https://github.com/jjmrocha/joe/actions/workflows/ci.yml/badge.svg)](https://github.com/jjmrocha/joe/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.27%2B-00ADD8)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

joe is a terminal coding agent. It reads and changes real code bases through
[Serena](https://github.com/oraios/serena)'s symbolic tools, and it works through skills —
loading the one that matches the request before it touches anything.

Named after [Joe Armstrong](https://en.wikipedia.org/wiki/Joe_Armstrong_(programmer)),
creator of Erlang.

Built on [ai-toolkit](https://github.com/jjmrocha/ai-toolkit) (agent, LLM client, tools,
skills) and [ai-chat](https://github.com/jjmrocha/ai-chat) (chat core, TUI, slash commands).

## Quickstart

```bash
git clone https://github.com/jjmrocha/joe.git && cd joe
make build                      # writes ./bin/joe

export OPEN_ROUTER_KEY=sk-...
./bin/joe
```

That first run writes `~/.config/joe/default.json`, creates an empty `~/.config/joe/skills/`,
and then stops — naming every one of the eleven skills it could not find there. Copy those
eleven folders in from
[jjmrocha/coding-skills](https://github.com/jjmrocha/coding-skills), each as
`~/.config/joe/skills/<name>/SKILL.md`, and run it again.

From then on, run joe from the repository you want it to work on: it calls `repo_info` at the
start of a session to find the git root, and activates Serena on that path. Outside a
repository, and when `git` is not installed at all, joe uses the current directory instead.
But a `git` that is present and *refuses* — dubious ownership, an unreadable `.git` — stops
joe rather than letting it mistake a subdirectory for the root and quietly skip the
repository's `CLAUDE.md`.

`go install github.com/jjmrocha/joe/cmd@latest` also works, but installs a binary named
`cmd`, after its directory. `make build` names it `joe`.

## Prerequisites

| What | Why | How |
|---|---|---|
| Go 1.27+ | Building joe | [go.dev/dl](https://go.dev/dl/) |
| `OPEN_ROUTER_KEY` | joe talks to [OpenRouter](https://openrouter.ai) | `export OPEN_ROUTER_KEY=sk-...` |
| `uvx` on `PATH` | Starts Serena, which serves joe's coding tools | [uv](https://github.com/astral-sh/uv) |
| Skills in `~/.config/joe/skills` | joe loads a skill before acting | see below |
| `npx` on `PATH` | Starts the context7 MCP, which serves library documentation | [Node.js](https://nodejs.org) |
| `donsetch` on `PATH` | Starts the DonSeTch MCP, which serves web search and fetching | [donsetch](https://github.com/dondai44423/donsetch) |

The key is the one your profile names in `api-key-env`; with the default profile that is
`OPEN_ROUTER_KEY`, and an Ollama profile needs no key at all. The last two rows are needed
only on demand: Serena starts with joe and a missing `uvx` fails at launch, while an MCP
server starts the first time you use it unless the profile lists it in `mcps-on`.

The eleven skills joe loads by name are `analyze-code`, `brainstorm`, `coding-discipline`,
`designing-interfaces`, `guiding-manual-testing`, `knowledge-base`, `research`,
`style-checker`, `test-driven-development`, `using-software-specialists` and
`writing-unit-tests`. A missing one stops joe at startup, and no folder other than
`~/.config/joe/skills` is ever searched.

## Configuration

joe starts from a profile: a JSON file in `~/.config/joe/` — or in `$XDG_CONFIG_HOME/joe/`
when that variable is set. `./bin/joe` reads `default.json`, and `./bin/joe local` reads
`local.json`, so a second setup is a second file:

```bash
./bin/joe            # ~/.config/joe/default.json
./bin/joe local      # ~/.config/joe/local.json
```

The first run creates the folder, writes `default.json`, and creates an empty `skills/` and
an empty `AGENTS.md`. Nothing that already exists is overwritten. Asking for a profile that
does not exist is an error — only `default.json` is ever created for you.

That `AGENTS.md` is read only under `harness: agents`, and the `default.json` written beside
it says `claude`, so on the default profile it stays unread until you switch the harness.

**Each profile is complete.** joe runs exactly what the file says: a profile with no `mcps`
section gets no MCP servers. There is no merging between profiles and no hidden default.

```json
{
  "harness": "claude",
  "llm": {
    "provider": "openrouter",
    "base-url": "",
    "api-key-env": "OPEN_ROUTER_KEY",
    "model": "z-ai/glm-5.3-flash",
    "models": ["z-ai/glm-5.3-flash", "deepseek/deepseek-v4-pro"],
    "effort": "medium"
  },
  "skills": ["removing-ai-tells"],
  "mcps": {
    "context7": { "command": "npx", "args": ["-y", "@upstash/context7-mcp"], "timeout": "60s" },
    "github": { "command": "github-mcp-server", "args": ["stdio"], "env": ["GITHUB_TOKEN"] }
  },
  "mcps-on": ["github"]
}
```

| Key | What it does |
|---|---|
| `harness` | `claude` or `agents` — which instruction files joe reads, see below |
| `llm.provider` | `openrouter`, `ollama` or `anthropic` |
| `llm.base-url` | Overrides the provider's endpoint; empty uses the standard one |
| `llm.api-key-env` | The **name** of the variable holding the key, never the key itself. Required except on Ollama |
| `llm.model` | The model joe starts with |
| `llm.models` | The models `/model` switches between |
| `llm.effort` | `off`, `low`, `medium` or `max` |
| `skills` | Extra skills by name, loaded from `~/.config/joe/skills` beside joe's own |
| `mcps` | MCP servers joe registers: `command`, `args`, `env` (variables inherited from joe), `timeout` |
| `mcps-on` | The servers started at launch; the rest start on first use |

A profile that names an unknown provider, effort, harness or `mcps-on` server, a skill that is
not a bare name, or leaves the model empty, or names an API-key variable that is not set, stops
joe before the session opens — and the message lists every fault in the file, not just the
first. A server in `mcps-on` that fails to start is the exception: joe reports it and carries
on, and `/mcp` shows it as `off` so you can retry with `/mcp on <name>`.

Serena is not configurable here: it always starts with joe.

## Your instructions

At startup joe reads your standing instruction files and puts each one into its prompt
verbatim, in order, skipping any that are absent. Which three it reads depends on `harness`:

| `harness` | Files, least specific first |
|---|---|
| `claude` | `~/.claude/CLAUDE.md`, `<repo>/CLAUDE.md`, `<repo>/CLAUDE.local.md` |
| `agents` | `~/.config/joe/AGENTS.md`, `<repo>/AGENTS.md`, `<repo>/AGENTS.local.md` |

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

The **last** `kb_path` line in a file wins, and the search does not skip fenced code blocks —
so a file that documents the setting after using it is read as setting it twice, and the
example wins. Keep one line per file.

The path must be absolute and spelled out in full — `~` is not expanded, so `kb_path=~/wiki`
is rejected. A relative path, a `~`-relative one, or a folder that cannot be opened stops joe
at startup rather than running without the knowledge base. Without `kb_path` joe starts
normally and the `file_*` tools are simply absent.

The `knowledge-base` skill looks for `kb_path` in a CLAUDE file of its own accord, so on
`harness: agents` keep the line in `~/.claude/CLAUDE.md` as well — joe reads it from the
AGENTS files, the skill reads it from there.

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

## Troubleshooting

joe validates what it can before the session opens, and the message names the fault. These
are the ones you are most likely to meet.

| Message | Cause | Fix |
|---|---|---|
| `skill folder not found: …` | One or more of the eleven skills is missing from `~/.config/joe/skills` | Copy the folders from [coding-skills](https://github.com/jjmrocha/coding-skills). Every missing skill is listed at once |
| `api key variable is not set` | `api-key-env` names a variable with no value | `export` it, or point `api-key-env` at the one you use |
| `profile not found` | `joe <name>` with no `<name>.json` | Create the file; only `default.json` is written for you |
| `kb_path is not absolute` | A `kb_path` line is relative or starts with `~` | Spell the path out in full |
| `harness is not claude or agents` | Unknown `harness` value | Use `claude` or `agents` |
| `git rev-parse: …` | `git` is installed but refusing — dubious ownership, unreadable `.git`. Stops the launch, and fails `repo_info` if it happens mid-session | Fix the repository, or run joe somewhere else |

A failure to start Serena is the one that does not name itself clearly: it surfaces as an
error from the coding tool pack at launch, and the usual cause is `uvx` missing from `PATH`.

An `mcps-on` server that fails to start does *not* stop joe. The failure prints before the
TUI opens and `/mcp` shows the server as `off`; `/mcp on <name>` retries it.

## Development

`make` on its own lists every target:

| Target | Effect |
|---|---|
| `build` | Build joe into `./bin` |
| `clean` | Remove `./bin` |
| `test` | Run all tests |
| `bench` | Run benchmarks |
| `lint` | Run golangci-lint |
| `deps` | Update dependencies |
| `tidy` | Tidy `go.mod` |

CI runs `go test -race ./...`, `golangci-lint` and `govulncheck` on every push and pull
request ([`.github/workflows/ci.yml`](.github/workflows/ci.yml)). Run the `-race` form before
pushing — `make test` does not. golangci-lint must be v2.13.2 or newer; earlier releases are
built with an older Go and refuse this module.

## License

[MIT](LICENSE) © 2026 Joaquim Rocha
