<div align="center">

# joe

**The opinionated coding agent for your terminal.**

[![CI](https://github.com/jjmrocha/joe/actions/workflows/ci.yml/badge.svg)](https://github.com/jjmrocha/joe/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.27%2B-00ADD8)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

[Quickstart](#quickstart) · [Using joe](#using-joe) · [Configure](#configure) · [Troubleshooting](#troubleshooting)

</div>

Point joe at a repository, describe what you want, and it reads the code, changes it, runs
the tests and tells you what it did.

```
$ cd ~/src/my-api
$ joe

❯ the /users endpoint 500s when the page param is missing — fix it
```

joe loads its debugging skill, looks up `paginate` through Serena, reproduces the panic,
writes a failing test, fixes it, runs `go test ./...` and reports back with the file and line
it changed.

Named after [Joe Armstrong](https://en.wikipedia.org/wiki/Joe_Armstrong_(programmer)),
creator of Erlang.

---

## Why "opinionated"

**It works through skills.** Before touching anything, joe loads the skill that matches the
request — a written procedure for that kind of work. Asked to fix a bug, it loads the
debugging procedure; asked to build a feature, it loads the one that starts by pinning down
requirements. The skills are not joe's own: they live in
[coding-skills](https://github.com/jjmrocha/coding-skills), joe clones them on first run, and
you can edit them. **Change a skill and you change how joe works.**

**It reads code symbolically.** joe works through [Serena](https://github.com/oraios/serena),
so it looks up a function or a type and its references instead of grepping and guessing.

**You bring the model.** joe talks to OpenRouter, Anthropic or a local Ollama, whichever your
profile names.

joe itself is small — the prompt, the configuration and the wiring:

| Built on | What it provides |
|---|---|
| [ai-toolkit](https://github.com/jjmrocha/ai-toolkit) | The agent loop, the LLM clients (OpenRouter, Anthropic, Ollama), the tool packs and the skill loader |
| [ai-chat](https://github.com/jjmrocha/ai-chat) | The chat core, the terminal UI and the slash commands |
| [coding-skills](https://github.com/jjmrocha/coding-skills) | The twelve skills joe loads by name |

---

## Quickstart

Three steps, about five minutes.

### 1. Check the prerequisites

| What | Why | How |
|---|---|---|
| Go 1.27+ | Building joe | [go.dev/dl](https://go.dev/dl/) |
| `git` | joe finds the repository root with it, and clones the skills on first run | Your package manager |
| `uvx` on `PATH` | Starts Serena, which serves joe's coding tools. **joe will not start without it** | [uv](https://github.com/astral-sh/uv) |
| An API key | Unless you run Ollama locally | [OpenRouter](https://openrouter.ai) or [Anthropic](https://console.anthropic.com) |

### 2. Build joe and export your key

```bash
git clone https://github.com/jjmrocha/joe.git && cd joe
make build                       # writes ./bin/joe

export OPEN_ROUTER_KEY=sk-...    # put this in your shell profile
```

Copy `./bin/joe` somewhere on your `PATH` to run it as `joe` from anywhere.

joe never stores the key — the profile holds the *name* of the variable, and joe reads the
variable at startup.

> `go install github.com/jjmrocha/joe/cmd@latest` works too, but names the binary `cmd`,
> after its directory.

### 3. Run joe

```bash
./bin/joe
```

It finds no configuration folder, so it says where it will write and asks you to describe
the setup:

```
joe — The opinionated coding agent for your terminal.

First run: a few questions to set up your profile, saved to

  /Users/you/.config/joe/default.json

which you can edit later. Then joe clones its skills into

  /Users/you/.config/joe/coding-skills

Provider [anthropic, ollama, openrouter]: openrouter
Model: z-ai/glm-5.3-flash
Name of the API key variable: OPEN_ROUTER_KEY
Harness [agents, claude]: claude

Knowledge base [yes, no]: yes
Knowledge base folder: /Users/you/Documents/LLM_WIKI

Classifier model [yes, no]: yes
Classifier model name: typesafe/jev-1.13
Name of the classifier API key variable: OPEN_ROUTER_KEY
```

The API-key question is skipped on Ollama, and the folder and classifier questions only
follow a `yes`.

- **`Harness`** decides which instruction files joe reads — see
  [Your own instructions](#your-own-instructions). Pick `claude` if you already keep a
  `~/.claude/CLAUDE.md`.
- **`Knowledge base`** decides whether joe gets the `file_*` tools — see
  [A knowledge base of your own](#a-knowledge-base-of-your-own). **joe does not create the
  folder**: it must already exist, and joe asks again until it does. A leading `~` is
  expanded before joe checks, and the profile stores the path in full.
- **`Classifier model`** decides whether joe screens tool calls and gets the
  `classify_*` tools — see [Configure](#configure). It runs on OpenRouter, so the key is an
  OpenRouter one.

From your answers joe writes the profile, then clones the skills:

```
~/.config/joe/
├── default.json    your profile
├── AGENTS.md       empty, for harness: agents
├── coding-skills/  a git clone of coding-skills — joe's twelve skills
└── skills/         empty — for extra skills of your own
```

Then the chat opens. **You're done.**

If the clone fails — no network, say — joe stops with `clone skills: …`. Your answers are
kept: the next run clones again without asking them.

<details>
<summary>The twelve skills</summary>

`addressing-findings`, `analyze-code`, `brainstorm`, `coding-discipline`,
`designing-interfaces`, `guiding-manual-testing`, `knowledge-base`, `research`,
`style-checker`, `test-driven-development`, `using-software-specialists`,
`writing-unit-tests`.

Each loads from `~/.config/joe/coding-skills/<name>/SKILL.md`. Edit them there; joe never
touches the clone again. To pick up upstream changes, merged with your edits:

```bash
git -C ~/.config/joe/coding-skills pull
```

</details>

### Optional extras

<details>
<summary>Web search and fetching · library documentation</summary>

The default profile registers two MCP servers. Both start the first time joe uses them, not
at launch — so a missing one costs you nothing until a request needs it. `/mcp` lists them
and their state.

**[DonSeTch](https://github.com/dondai44423/donsetch)** gives joe web search, page fetching
and crawling. Without it joe works fine, offline; a request that needs the web fails when the
tool is called.

```bash
npm install -g donsetch
# or: brew tap dondai44423/donsetch && brew install donsetch
```

**[context7](https://github.com/upstash/context7)** serves up-to-date documentation for
libraries and frameworks. It runs through `npx`, so installing
[Node.js](https://nodejs.org) is all it needs.

</details>

---

## Using joe

Run joe from the repository you want it to work on. It calls `repo_info` at the start of the
session to find the git root and activates Serena on that path. Outside a repository — and
when `git` is not installed at all — joe uses the current directory.

| Command | What it does |
|---|---|
| `/help` | List the commands |
| `/model [name]` | Show the current model, or switch to another from `llm.models` |
| `/effort [level]` | Show or change reasoning effort |
| `/clear` | Reset the conversation |
| `/compact` | Compact the context now, instead of waiting for joe to do it |
| `/mcp [on\|off] [name]` | Show the MCP servers, or start and stop one |
| `/exit` | Quit |

### Your own instructions

At startup joe reads your standing instruction files and puts each into its prompt verbatim,
least specific first, skipping any that are absent. Which three depends on `harness`:

| `harness` | Files, least specific first |
|---|---|
| `claude` | `~/.claude/CLAUDE.md`, `<repo>/CLAUDE.md`, `<repo>/CLAUDE.local.md` |
| `agents` | `~/.config/joe/AGENTS.md`, `<repo>/AGENTS.md`, `<repo>/AGENTS.local.md` |

A later file wins where two disagree, and joe's own instructions win over all of them on
tools, skills, Serena and the knowledge base — the rest is yours.

Imports are **not** followed. A line like `@RTK.md` is passed through as text; joe never
opens the file it names.

### A knowledge base of your own

When the profile sets `kb-path`, joe gets a second set of tools — `file_read`, `file_write`,
`file_edit`, `file_list`, `file_delete`, `file_workdir` — rooted at that folder and unable to
leave it. The repository stays Serena's; that folder is joe's to write in, which is where it
keeps notes, plans and manuals across sessions.

```json
{ "kb-path": "/Users/you/Documents/LLM_WIKI" }
```

The path must be absolute and the folder must already exist — joe never creates it. A
relative path stops joe at startup with every other fault in the profile; a folder that is
missing or unreadable stops it when the tools are registered. Omit the key and joe starts
normally with the `file_*` tools simply absent.

Different profiles can name different folders, so `joe work` and `joe personal` can keep
separate knowledge bases.

<details>
<summary>Why a <code>kb_path</code> line in CLAUDE.md is ignored</summary>

**The profile is the only place joe reads this from.** A `kb_path` line in a `CLAUDE.md` or
`AGENTS.md` is quoted into the prompt like any other line of those files, and otherwise
ignored — joe will not take a knowledge-base root from a file a repository can ship. If you
used to keep the line in `~/.claude/CLAUDE.md`, move the value into your profile; leaving the
line where it is does no harm.

Either way joe closes the prompt with a `<knowledge-base>` block naming the path it actually
uses, or naming none, so a stale `kb_path` line elsewhere in the prompt cannot mislead the
model:

```
<knowledge-base>
Ignore any kb_path set anywhere above. This block is the only one that counts.

kb_path=/Users/you/Documents/LLM_WIKI
...
```

</details>

### Reading other repositories

A feature that spans repositories is still written in the one joe was started in. joe can
read the others but not change them: it reaches them through Serena's `query_project`, which
refuses every editing tool.

`query_project` only reaches repositories Serena has registered. Register each one once:

```bash
uvx --from git+https://github.com/oraios/serena serena project create /path/to/repo
```

Reading and searching files needs nothing more. Symbol lookups in another repository go
through Serena's project server, which joe does not start — run it alongside joe if you want
them:

```bash
uvx --from git+https://github.com/oraios/serena serena start-project-server
```

Without it, joe falls back to searching the other repository as text.

---

## Configure

joe starts from a profile: a JSON file in `~/.config/joe/` — or in `$XDG_CONFIG_HOME/joe/`
when that variable is set. `joe` reads `default.json`; `joe <name>` reads `<name>.json`, so a
second setup is a second file:

```bash
joe                  # ~/.config/joe/default.json
joe local            # ~/.config/joe/local.json — an Ollama profile, say
```

Only `default.json` is ever written for you; create the others by hand, or copy that one.
**Keep it.** joe decides whether it needs to run setup by looking for `default.json`, so
deleting it — even if you only ever use named profiles — makes the next run ask the setup
questions again.

**Each profile is complete.** joe runs exactly what the file says: a profile with no `mcps`
section gets no MCP servers. There is no merging between profiles and no hidden default.

```json
{
  "harness": "claude",
  "kb-path": "/Users/you/Documents/LLM_WIKI",
  "llm": {
    "provider": "openrouter",
    "api-key-env": "OPEN_ROUTER_KEY",
    "model": "z-ai/glm-5.3-flash",
    "models": ["z-ai/glm-5.3-flash", "deepseek/deepseek-v4-pro"],
    "effort": "medium"
  },
  "skills": ["removing-ai-tells"],
  "mcps": {
    "context7": { "command": "npx", "args": ["-y", "@upstash/context7-mcp"], "timeout": 60 },
    "github": { "command": "github-mcp-server", "args": ["stdio"], "env": ["GITHUB_TOKEN"] }
  },
  "mcps-on": ["github"],
  "classifier": {
    "provider": "openrouter",
    "api-key-env": "OPEN_ROUTER_KEY",
    "model": "typesafe/jev-1.13"
  }
}
```

| Key | What it does |
|---|---|
| `harness` | `claude` or `agents` — which instruction files joe reads |
| `kb-path` | Absolute path to your knowledge base. Omit it for no knowledge base and no `file_*` tools |
| `llm.provider` | `openrouter`, `ollama` or `anthropic` |
| `llm.base-url` | Overrides the provider's endpoint; omit it to use the standard one |
| `llm.api-key-env` | The **name** of the variable holding the key, never the key itself. Required except on Ollama |
| `llm.model` | The model joe starts with |
| `llm.models` | The models `/model` switches between |
| `llm.effort` | `off`, `low`, `medium` or `max` — how much the model reasons before answering |
| `skills` | Extra skills by name, loaded from `~/.config/joe/skills`. Cannot name one of joe's own twelve |
| `mcps` | MCP servers joe registers: `command`, `args`, `env` (variables inherited from joe), `timeout` in seconds — `0` or absent means no limit |
| `mcps-on` | The servers started at launch; the rest start on first use |
| `classifier` | The classification model that screens tool calls and backs the `classify_*` tools. Omit it and every call runs unchecked, with no `classify_*` tools |
| `classifier.provider` | `openrouter` |
| `classifier.base-url` | Overrides the provider's endpoint; omit it to use the standard one |
| `classifier.api-key-env` | The **name** of the variable holding the key. Required |
| `classifier.model` | The classification model, e.g. `typesafe/jev-1.13` |

A profile that names an unknown provider, effort, harness or `mcps-on` server, a skill that
is not a bare name or is one of joe's own twelve, a `kb-path` that is not absolute, or leaves the model empty, or names an
API-key variable that is not set, stops joe before the session opens — and the message lists
every fault in the file, not just the first.

With `classifier` set, joe asks the classification model about a tool call before it runs: the tool,
its arguments, and the rules of the session — files change only inside the repository and
the knowledge base, nothing remote or shared changes, no secret leaves the machine. A call
judged to break them is refused, and the model is told `rejected by joe`. If the classification
model errors or takes longer than five seconds, the call runs: the check fails open.

The first call to each tool also asks whether *any* call to that tool could break the rules. A
tool judged unable to — `current_date`, say — is not checked again until joe exits; every other
tool keeps having each call checked. That verdict is read from the tool's own description, so an
MCP server can talk the model into trusting a tool it should not: only add servers you trust.

The same model is offered to joe as three tools — `classify_yes_no`, `classify_choice` and
`classify_score` — so it can hand a judgement call to a calibrated model instead of guessing.

Serena is not configurable here: it always starts with joe.

---

## Troubleshooting

joe validates what it can before the session opens, and the message names the fault.

| Message | Cause | Fix |
|---|---|---|
| `skill folder not found: …` | One of the twelve is missing from `~/.config/joe/coding-skills`, or an extra from `~/.config/joe/skills` | After upgrading joe, `git -C ~/.config/joe/coding-skills pull` — a newer joe can need a skill your clone predates. Otherwise restore it, or delete `coding-skills/` and joe clones it again. Every missing one is listed at once |
| `clone skills: …` | The first-run clone of coding-skills failed; git's own message is printed above it | Check the network and `git`, then run joe again — only the clone is retried |
| `skill is one of joe's own` | The profile's `skills` names one of the twelve | Remove it from `skills`; edit that skill in `coding-skills/` instead |
| `api key variable is not set` | `api-key-env` names a variable with no value | `export` it, or point `api-key-env` at the one you use |
| `classifier provider is not openrouter` | The profile's `classifier.provider` is anything else | Use `openrouter` |
| `json: unknown field "som"` | The profile predates the rename of `som` to `classifier` | Rename the key to `classifier`; its contents stay the same |
| `profile not found` | `joe <name>` with no `<name>.json` | Create the file; only `default.json` is written for you |
| `no answer to read` | Setup ran with nothing on stdin — a pipe, a redirect, or Ctrl-D at a question | Run joe from a terminal and answer the questions; nothing is left broken, the next run simply asks again |
| `kb_path is not absolute` | The profile's `kb-path` is relative or starts with `~` | Spell the path out in full |
| `opening root: …` | The profile's `kb-path` names a folder that is missing or unreadable | Create it, or drop the key |
| `harness is not claude or agents` | Unknown `harness` value | Use `claude` or `agents` |
| `git rev-parse: …` | `git` is present but refusing — dubious ownership, unreadable `.git` | Fix the repository, or run joe somewhere else |

A failure to start Serena is the one that does not name itself clearly: it surfaces as an
error from the coding tool pack at launch, and the usual cause is `uvx` missing from `PATH`.

An `mcps-on` server that fails to start does *not* stop joe. The failure prints before the
TUI opens and `/mcp` shows the server as `off`; `/mcp on <name>` retries it.

**To start over**, delete `~/.config/joe` — or just its `default.json` — and run joe again:
it asks the setup questions afresh. joe looks for `default.json`, not for the folder, so a
setup you interrupted is finished by the next run rather than leaving you stuck. Nothing
already in the folder is ever overwritten: an existing `default.json` means setup does not
run at all, and an `AGENTS.md` you have edited is left as it is.

---

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
