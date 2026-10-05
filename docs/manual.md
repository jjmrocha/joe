# joe manual

For engineers who already use coding agents. For what joe is and why, see the [README](../README.md).

- [Install and first run](#install-and-first-run)
- [Everyday use](#everyday-use)
- [The workflow](#the-workflow)
- [Knowledge base](#knowledge-base)
- [Second opinion and guard](#second-opinion-and-guard)
- [Code access](#code-access)
- [Your own instructions](#your-own-instructions)
- [Profiles and configuration](#profiles-and-configuration)
- [Troubleshooting](#troubleshooting)
- [Development](#development)

---

## Install and first run

### Prerequisites

| What | Why |
|---|---|
| `git` | Finding the repository root; cloning the skills on first run |
| `uvx` on `PATH` ([uv](https://github.com/astral-sh/uv)) | Starts Serena, which serves the coding tools. **joe will not start without it** |
| An API key | [OpenRouter](https://openrouter.ai) or [Anthropic](https://console.anthropic.com); none for a local Ollama |

### Install

Download the archive for your system from [Releases](https://github.com/jjmrocha/joe/releases)
and put `joe` on your `PATH`.

To build from source instead (Go 1.27+), run `make build` in a clone of the repository; it writes `./bin/joe`.

### Setup questions

With no `default.json`, the first run asks:

```
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

| Question | Decides |
|---|---|
| Provider, Model, API key variable | The model joe talks to. The key question is skipped on Ollama |
| Harness | Which instruction files joe reads ([Your own instructions](#your-own-instructions)). Pick `claude` if you keep a `~/.claude/CLAUDE.md` |
| Knowledge base | Whether joe gets the `file_*` tools. The folder must exist; joe asks until it does. `~` is expanded and the full path stored |
| Classifier model | Whether joe gets the guard and the `classify_*` tools. Runs on OpenRouter, so the key is an OpenRouter one (JEV or equivalent model) |

The folder and classifier follow-ups appear only after `yes`. Setup writes `effort: medium`, registers two MCP servers (both off) and clones the skills:

```
~/.config/joe/
├── default.json    your profile
├── AGENTS.md       empty, for harness: agents
├── coding-skills/  git clone of coding-skills: joe's skills
└── skills/         empty, for your own skills
```

`$XDG_CONFIG_HOME/joe/` replaces `~/.config/joe/` when the variable is set.

If the clone fails, joe stops with `clone skills: …` and keeps your answers; the next run retries the clone without asking again. Nothing already in the folder is overwritten.

### Optional MCP servers

The default profile registers two servers and starts neither. Turn one on with `/mcp on <name>`, or list it in `mcps-on` to start it at launch.

| Server | Provides | Needs |
|---|---|---|
| [donsetch](https://github.com/dondai44423/donsetch) | Web search, page fetching, crawling | `npm install -g donsetch` or `brew tap dondai44423/donsetch && brew install donsetch` |
| [context7](https://github.com/upstash/context7) | Current docs for libraries and frameworks | [Node.js](https://nodejs.org) (runs via `npx`) |

Without them joe works offline; a request needing the web fails when the tool is called.

---

## Everyday use

Run joe from the repository you want it to work on. It works at the git root, with Serena's project already active. Outside a repository, or without `git`, it uses the current directory.

```
Usage: joe [profile] [-resume <id>]
```

| Command | Does |
|---|---|
| `/help` | List commands |
| `/model [name]` | Show the model, or switch to one from `llm.models` |
| `/effort [level]` | Show or change reasoning effort |
| `/clear` | Reset the conversation and start a new session |
| `/compact` | Compact the context now |
| `/export` | Save the session to `~/.config/joe/sessions/<id>.json`; exporting again updates the file |
| `/mcp [on\|off] [name]` | Show MCP servers, or start/stop one |
| `/skills` | List loaded skills |
| `/<skill> [request]` | Run the request through that skill (see below) |
| `/exit` | Quit |

### Skill commands

When you already know the procedure, name it: `/brainstorm add a --verbose flag`, `/research how does Load resolve the profile?`.

| Command | Use it to |
|---|---|
| `/research` | Answer a question about the code base, with sources |
| `/brainstorm` | Turn an idea into an approved design |
| `/using-software-specialists` | Plan and implement a code change |
| `/analyze-code` | Review code for bugs, security and tech debt |
| `/addressing-findings` | Work through review findings one at a time |
| `/guiding-manual-testing` | Guide manual testing of a change |
| `/writing-unit-tests` | Add unit tests to existing code |
| `/knowledge-base` | Query or update the knowledge base (only when `kb-path` is set) |

Your own skills have no command; joe routes to them by description.

### Resume a session

```bash
joe -resume <id>          # same as: joe default -resume <id>
joe work -resume <id>
```

The model gets the whole conversation back; the screen starts empty, and `/export` keeps writing to the same file. joe resumes only in the repository the session was exported from. The profile, not the file, decides model and effort.

---

## The workflow

### Routing

Every request that reads, changes, judges or documents code starts by loading one entry skill. joe picks it by what you want to end up with, announces it (`Route: … → Loading <skill>`), then follows it. Only a rename, typo or comment-only edit skips routing; "too small for a skill" is not an exemption. A precise request still goes through `brainstorm`: precision shortens it, never skips it.

| You want | Entry skill |
|---|---|
| An answer about your code | `research` |
| A new capability, interface, flag, refactor or migration | `brainstorm` |
| A bug fixed, or an approved plan implemented | `using-software-specialists` |
| A review | `analyze-code` |
| To go through findings or PR comments | `addressing-findings` |
| To verify a change by hand on a real environment | `guiding-manual-testing` |
| Tests for existing code, no production change | `writing-unit-tests` |
| To query or update the knowledge base | `knowledge-base` |

The routing table and tie-breakers live in [`internal/prompt/skills.go`](../internal/prompt/skills.go). joe's routing wins over a skill's own "When NOT to use".

### The loop

How the entry skills hand off to each other is in the README: [Why joe: the workflow](../README.md#why-joe-the-workflow).

### Supporting skills

Loaded by the entry skills at the step that needs them, never as entry points.

| Skill | Loaded |
|---|---|
| `coding-discipline` | Before any diff: names six LLM failure modes and their counters |
| `designing-interfaces` | Before a new or widened interface: a four-line contract; sent back if it hides nothing |
| `test-driven-development` | During implementation: red, green, refactor |
| `writing-unit-tests` | By TDD, or as an entry for tests only |
| `style-checker` | As the style lens of `analyze-code` |
| `knowledge-base` | By `brainstorm`, `research`, `using-software-specialists` and `analyze-code` when `kb-path` is set |

### Your own skills

List extra skills by name in the profile's `skills`; joe loads each from `~/.config/joe/skills/<name>/SKILL.md`. A name colliding with one of joe's twelve is rejected. Your skills are routed by their description when it fits a request better than any table row.

### Updating the skills

joe loads its twelve skills from `~/.config/joe/coding-skills/<name>/SKILL.md` and never updates the clone itself:

```bash
git -C ~/.config/joe/coding-skills pull
```

A newer joe can need a skill an older clone lacks; pull after upgrading.

---

## Knowledge base

The knowledge base is a folder of Markdown outside your repositories. It holds what one repository can't tell an agent: which service owns a piece of data, who consumes an event, what a plan intended.

| Folder | Answers | Holds |
|---|---|---|
| `wiki/` | What exists? | Per-repo pages: entities, interfaces, jobs, dependencies, events, rules, helpers, patterns |
| `plans/` | What's intended? | Plans written by `brainstorm`, often cross-repo |
| `manuals/` | How do I use this? | Plain-language docs for operators, written only when you ask |

joe reads the knowledge base before it works. It writes plans on its own; the wiki changes only when you ask. Every wiki page lists its `sources:`, and joe treats code as truth: a wiki claim is checked against the code it cites, and a disagreement is reported, not silently resolved. Deletes need your approval.

### Getting started

1. Create an empty folder and set `kb-path` to it (or answer `yes` at setup).
2. On first use joe asks before creating the layout (`wiki/`, `plans/`, `manuals/`), and before adding each repository's folder under `wiki/`.
3. In each repository: `/knowledge-base ingest this repository`. joe reads every file, writes the pages, links them to other services already in the wiki, and reports coverage (`Ingested 42/42 files`).

| To | Ask |
|---|---|
| Add documents | `/knowledge-base ingest docs/architecture/` |
| Record a change you just made | `/knowledge-base update the wiki with what we just changed` |
| Find stale or orphaned pages | `/knowledge-base audit the KB` |
| Look something up | `/knowledge-base who consumes order-created?` |

### Tools

With `kb-path` set, joe gets seven tools confined to that folder: `file_read`, `file_write`, `file_edit`, `file_list`, `file_search`, `file_delete`, `file_workdir`. The repository stays Serena's.

### `kb-path` rules

- Absolute, and the folder must exist; joe never creates it. A relative path fails profile validation; a missing folder fails with `opening root: …`.
- Omit it and joe starts without the `file_*` tools and `/knowledge-base`.
- The profile is the only source. A `kb_path` line in `CLAUDE.md` or `AGENTS.md` is quoted like any other line and otherwise ignored: a repository can't ship a KB root. The prompt names the path joe uses and tells the model to ignore others.
- Each profile can name its own, so `joe work` and `joe personal` keep separate knowledge bases.

---

## Second opinion and guard

Both need a `classifier` in the profile: a classification model, different from the model that writes code. It writes no text; it answers a typed question (yes/no, a choice, a score) with a calibrated probability. joe reaches it through OpenRouter's decisions API, so any classification model OpenRouter serves works; today that is TypeSafe's [Jev](https://openrouter.ai/blog/insights/what-is-jev/) (`typesafe/jev-1.13`). Without a classifier, neither feature exists.

### Second opinion

joe gets three tools backed by the classifier: `classify_yes_no`, `classify_choice`, `classify_score`. Its prompt makes three calls required and binding:

| When | Call | Asks |
|---|---|---|
| `test-driven-development` REFACTOR, per function changed in GREEN | `classify_yes_no`, repeated | Would a senior engineer refactor this function further? |
| `analyze-code`, per surviving finding | `classify_choice` | Severity, from the skill's scale |
| `designing-interfaces`, after the contract | `classify_score`, repeated | How deep is this interface? |

joe overrides a verdict only on a named contradicting fact, and reports overrides and failures. Each call is billed.

### Guard

Before every tool call, joe sends the classifier the tool, its arguments and the session's rules:

- files change only inside the repository and the knowledge base;
- nothing remote or shared changes: no push, deploy, publish, merge or message;
- no secret or credential leaves the machine.

A call judged to break them (probability ≥ 0.8) is refused with `rejected by joe`. The agent stops and tells you what was refused instead of trying another route. A call whose classifier input exceeds 64 KB is refused with `rejected by joe: the request is too big`.

**The guard fails open.** If the classifier errors or takes longer than five seconds, the call runs. It is a second line of defence, not a sandbox.

**Tools earn trust.** The first call to each tool also asks whether *any* call to it could break the rules. A tool judged unable to (`current_date`, say) is not checked again until joe exits; `/clear` doesn't reset it. That verdict is read from the tool's own description, so an MCP server can talk the classifier into trusting a tool it shouldn't: **only add MCP servers you trust.**

---

## Code access

joe reads and edits the repository through [Serena](https://github.com/oraios/serena), which works on symbols via language servers (40+ languages). It looks up a function, its body or its references instead of reading whole files, which keeps context small, and renames across files in one call. Serena always starts with joe and isn't configurable. joe also has `shell_run` for builds and tests.

### Other repositories

joe changes only the repository it started in. Others it reads through `serena__query_project`, which refuses every editing tool. Register each repository once:

```bash
uvx --from git+https://github.com/oraios/serena serena project create /path/to/repo
```

That covers reading and searching files. Symbol lookups in another repository need Serena's project server, which joe doesn't start:

```bash
uvx --from git+https://github.com/oraios/serena serena start-project-server
```

Without it, joe searches the other repository as text.

---

## Your own instructions

joe quotes your instruction files into its prompt verbatim, least specific first, skipping missing ones:

| `harness` | Files |
|---|---|
| `claude` | `~/.claude/CLAUDE.md`, `<repo>/CLAUDE.md`, `<repo>/CLAUDE.local.md` |
| `agents` | `~/.config/joe/AGENTS.md`, `<repo>/AGENTS.md`, `<repo>/AGENTS.local.md` |

A later file wins where two disagree. joe's own instructions win over all of them: your files can add rules or narrow joe's, never override or relax them.

Imports aren't followed: `@RTK.md` is passed through as text.

---

## Profiles and configuration

A profile is one JSON file in `~/.config/joe/`. `joe` reads `default.json`; `joe <name>` reads `<name>.json`. Use one per situation:

```bash
joe                      # default.json
joe work                 # work.json: company KB, stronger model, classifier
joe local                # local.json: Ollama, no API key
joe local -resume <id>
```

Each profile is complete: joe runs exactly what the file says, with no merging and no hidden defaults. A profile with no `mcps` gets no MCP servers. Only `default.json` is written for you; copy it for others. Keep `default.json`: deleting it makes the next run ask the setup questions again.

```json
{
  "harness": "claude",
  "kb-path": "/Users/you/Documents/LLM_WIKI",
  "llm": {
    "provider": "openrouter",
    "api-key-env": "OPEN_ROUTER_KEY",
    "model": "z-ai/glm-5.3-flash",
    "models": ["z-ai/glm-5.3-flash", "z-ai/glm-5.3"],
    "effort": "max"
  },
  "skills": [],
  "mcps": {
    "context7": { "command": "npx", "args": ["-y", "@upstash/context7-mcp"], "timeout": 60 },
    "donsetch": { "command": "donsetch", "args": ["mcp", "--supervised"], "timeout": 900 }
  },
  "mcps-on": [],
  "classifier": {
    "provider": "openrouter",
    "api-key-env": "OPEN_ROUTER_KEY",
    "model": "typesafe/jev-1.13"
  }
}
```

| Key | Meaning |
|---|---|
| `harness` | `claude` or `agents`: which instruction files joe reads |
| `kb-path` | Absolute path to the knowledge base. Omit for none |
| `llm.provider` | `openrouter`, `ollama` or `anthropic` |
| `llm.base-url` | Overrides the provider's endpoint |
| `llm.api-key-env` | **Name** of the variable holding the key. Required except on Ollama |
| `llm.model` | Model joe starts with |
| `llm.models` | Models `/model` switches between |
| `llm.effort` | `off`, `low`, `medium` or `max`. Setup writes `medium` |
| `skills` | Your extra skills by name, from `~/.config/joe/skills` |
| `mcps` | MCP servers: `command`, `args`, `env` (variables inherited from joe), `timeout` in seconds (`0` or absent: 60) |
| `mcps-on` | Servers started at launch; start others with `/mcp on <name>` |
| `classifier` | Model behind the guard and `classify_*`. Omit for neither |
| `classifier.provider` | `openrouter` |
| `classifier.base-url` | Overrides the provider's endpoint |
| `classifier.api-key-env` | Name of the variable holding the key. Required |
| `classifier.model` | Classification model, e.g. `typesafe/jev-1.13` |

joe validates the whole profile before the session opens and lists every fault at once: unknown provider, effort, harness or `mcps-on` server; a skill that isn't a bare name; a relative `kb-path`; an empty model; an unset API-key variable. Unknown keys are rejected.

---

## Troubleshooting

| Message | Cause | Fix |
|---|---|---|
| `skill folder not found: …` | A skill is missing from `coding-skills/`, or an extra from `skills/` | After upgrading joe, `git -C ~/.config/joe/coding-skills pull`. Otherwise restore it, or delete `coding-skills/` to re-clone. All missing ones are listed |
| `skill is one of joe's own` | `skills` names one of joe's twelve | Remove it from `skills` |
| `clone skills: …` | First-run clone failed; git's message is above it | Check network and `git`, run joe again. Only the clone is retried |
| `profile not found` | `joe <name>` with no `<name>.json` | Create it; only `default.json` is written for you |
| `profile name is not a bare name` | Name has characters other than letters, digits, `-`, `_` | Rename the profile |
| `api key variable is not set` | `api-key-env` names an empty variable | `export` it, or point `api-key-env` at the one you use |
| `provider is not openrouter, ollama or anthropic` | Unknown `llm.provider` | Use one of the three |
| `classifier provider is not openrouter` | Any other `classifier.provider` | Use `openrouter` |
| `effort is not off, low, medium or max` | Unknown `llm.effort` | Use one of the four |
| `model is not set` | Empty `llm.model` or `classifier.model` | Set it |
| `mcps-on names a server that is not registered` | `mcps-on` lists a name missing from `mcps` | Add the server to `mcps` or remove the name |
| `harness is not claude or agents` | Unknown `harness` | Use `claude` or `agents` |
| `kb-path is not absolute` | Relative or `~`-prefixed path | Spell it out in full |
| `opening root: …` | `kb-path` folder missing or unreadable | Create it, or drop the key |
| `no answer to read` | Setup ran with no terminal on stdin (pipe, redirect, Ctrl-D) | Run joe from a terminal; the next run asks again |
| `git rev-parse: …` | `git` refuses the repository (dubious ownership, unreadable `.git`) | Fix the repository, or run joe elsewhere |
| `coding tools: …` | Serena failed to start, usually no `uvx` on `PATH` | Install [uv](https://github.com/astral-sh/uv); check `uvx` runs |
| `invalid session id` / `session not found` | `-resume` with a malformed or unknown id | Use the id from a file in `~/.config/joe/sessions/` |
| `session belongs to another repository` | `-resume` outside the session's repository | Run joe from that repository |

- An `mcps-on` server that fails to start doesn't stop joe: the error prints before the chat opens, `/mcp` shows it `off`, and `/mcp on <name>` retries.
- To start over, delete `~/.config/joe` (or just `default.json`). An interrupted setup finishes on the next run; an existing `AGENTS.md` is never overwritten.

---

## Development

| `make` target | Effect |
|---|---|
| `build` | Build joe into `./bin` |
| `clean` | Remove `./bin` |
| `test` | All tests with the race detector |
| `bench` | Benchmarks |
| `lint` | golangci-lint (v2.13.2+; older builds refuse Go 1.27) |
| `deps` | Update dependencies |
| `tidy` | Tidy `go.mod` |

CI runs `go test -race ./...`, golangci-lint, govulncheck and a gitleaks scan on every push and pull request ([`ci.yml`](../.github/workflows/ci.yml)).
