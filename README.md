<div align="center">

<img src="logo.svg" alt="joe" width="360">

**A coding agent with a workflow of its own.**

[![CI](https://github.com/jjmrocha/joe/actions/workflows/ci.yml/badge.svg)](https://github.com/jjmrocha/joe/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.27%2B-00ADD8)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

[Why joe](#why-joe-the-workflow) · [What makes it practical](#what-makes-it-practical) · [Is joe for you?](#is-joe-for-you) · [Quickstart](#quickstart) · [Manual](docs/manual.md)

</div>

joe is one engineer's way of building software, turned into an agent. It doesn't adapt to
your workflow; it brings its own and follows it on every request. If that workflow fits
you, joe enforces it so you don't have to.

```
$ joe

❯ add a /version command that prints joe's version

  Route: a new capability (a new /version command) → Loading brainstorm

● skill_load(skill_name="brainstorm")
  ⎿ <6.5 KB> · 379ms

  Loading knowledge-base (kb_path configured)

● file_read(path="wiki/joe/index.md")
  ⎿ <10.5 KB> · 240ms

● serena__find_symbol(include_body=true, name_path_pattern="buildCommands", relative_path="internal/engine/commands.go")
  ⎿ <755 B> · 303ms

  ⋮ 10 more tool calls

  Findings from the repo and KB:

  • No version string exists anywhere in joe's code — no const, no ldflags, no debug.ReadBuildInfo use.
  • Slash commands are registered in internal/engine/commands.go (buildCommands, line 37) — the
    pattern in wiki/ai-chat/patterns/add-slash-command.

  The one real design decision is where the version value comes from:

  1. Hardcoded const — bumped by hand. Simple, drifts if forgotten. (my default)
  2. ldflags -X — set at build time. Accurate, but go run / go install print empty.
  3. runtime/debug.ReadBuildInfo — reports (devel) for local make build.

  Where should the version number come from?

──────────────────────────────────────────────────────────────────────────
 13 tool calls · 26s llm · 3s tools · ↑33.08K ↓616 tokens
```

No code yet. joe checked what exists, found the pattern to follow, and asked the one
question that matters.

---

## Why joe: the workflow

Every request is routed to a skill, a written procedure for that kind of work, and joe
says which one it loaded. Skills hand off to each other, so work moves through the same
loop every time:

```mermaid
flowchart LR
    Q([Question]) --> R[Research]
    I([Idea]) --> D[Design]
    R --> D
    D -->|approved plan| B[Build]
    B --> A[Review]
    A -->|findings| F[Decide]
    F -->|approved fix| B
    A -->|clean| S([Ship])
    B -.->|on request| T[Test by hand]
    T -->|bug| B
```

| Step | Skill | What joe does |
|---|---|---|
| Research | `research` | Answers from your code and notes, citing `file:line`. Never from memory |
| Design | `brainstorm` | Asks one question at a time, proposes a design, writes no code until you approve. Saves the plan |
| Build | `using-software-specialists` | Implements the plan test-first, checks each new interface's depth, guards against the usual LLM failure modes |
| Review | `analyze-code` | Audits the change through five lenses and reports. Fixes nothing |
| Decide | `addressing-findings` | Walks the findings one at a time. Nothing changes until you decide |
| Test by hand | `guiding-manual-testing` | Proposes one step at a time on a real environment; you run it, joe reads the output |

"Too small to need a skill" is never an excuse. A precise request still gets a short
design conversation: it confirms joe understood and surfaces what you didn't consider.

---

## What makes it practical

### A knowledge base of your organisation

One repository rarely tells the whole story. Point joe at a folder and it keeps a wiki of
your systems: which service owns a piece of data, who consumes an event, which process
writes a table. Across a dozen microservices, joe knows how the pieces fit before it
changes one. It also keeps the plans `brainstorm` writes, so work picks up where it left
off. Code stays the truth: a note that disagrees with the code is reported, not trusted.

### A second opinion at the critical calls

Some judgements matter more than others. At three of them joe must ask a separate
classification model (JEV), whose verdict is binding unless a named fact contradicts it:

- after a test goes green, whether a function needs more refactoring;
- how severe each review finding is;
- how deep a new interface is before it's built.

The same model screens tool calls before they run: files change only in your repository and
knowledge base, nothing remote changes, no secret leaves the machine.

### Code by symbol, not by file

joe works through [Serena](https://github.com/oraios/serena), which uses language servers
to look up a function, its body or its references instead of reading whole files. Less
context spent, more precise edits, 40+ languages. Other repositories are read-only.

### Profiles for every situation

A profile is one JSON file holding the model, knowledge base, MCP servers and classifier.
Keep one per situation: `joe work` with the company knowledge base and a frontier model,
`joe local` on Ollama, with no cloud model. Bring OpenRouter, Anthropic or
Ollama, and switch models mid-session.

joe also reads the `CLAUDE.md` or `AGENTS.md` files you already keep, and saves and resumes
sessions.

---

## Is joe for you?

**Yes, if you:**

- want the design agreed before code is written;
- want to approve every change, one decision at a time;
- work across several services and want the agent to remember how they connect;
- prefer a fixed, predictable process to an endlessly configurable one.

**No, if you:**

- want an agent that adapts to your way of working;
- want a general-purpose assistant or a Claude Code replacement;
- want quick edits without a conversation first.

---

## Quickstart

You need Go 1.27+, `git`, [uv](https://github.com/astral-sh/uv) (`uvx` on `PATH`) and an
API key, unless you use a local Ollama.

```bash
git clone https://github.com/jjmrocha/joe.git && cd joe
make build                       # writes ./bin/joe
export OPEN_ROUTER_KEY=sk-...    # the profile stores the variable's name, never the key

cd ~/your/repo && /path/to/joe/bin/joe
```

The first run asks a few questions, writes your profile to `~/.config/joe/`, clones the
skills and opens the chat. Details in the [manual](docs/manual.md#install-and-first-run).

---

## Learn more

The [manual](docs/manual.md) covers everyday use, the workflow in depth, the knowledge
base, the guard, configuration and troubleshooting.

joe itself is small: the prompt, the configuration and the wiring. It's built on
[ai-toolkit](https://github.com/jjmrocha/ai-toolkit) (agent loop, LLM clients, tools),
[ai-chat](https://github.com/jjmrocha/ai-chat) (terminal UI, commands) and
[coding-skills](https://github.com/jjmrocha/coding-skills) (the skills).

---

Named after [Joe Armstrong](https://en.wikipedia.org/wiki/Joe_Armstrong_(programmer)),
creator of Erlang.

[MIT](LICENSE) © 2026 Joaquim Rocha
