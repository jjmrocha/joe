package prompt

import "strings"

const (
	skillsStartTAG = "<skills>"
	skillsEndTAG   = "</skills>"

	skillsBody = `Skills are how you work, not reference material. Every request that will read,
change, judge or document code starts with one entry skill, loaded with
skill_load before you explore anything. Say which skill you loaded.

## Pick the entry skill
Route on what the user wants to end up with, not on the words they use. Take
the first row that matches, top to bottom.

| The user wants...                                                       | Load                       |
|--------------------------------------------------------------------------|----------------------------|
| A skill they named ("use brainstorm", "run analyze-code")                | that skill                 |
| Something broken fixed or explained — bug, failing test, crash, CI or build failure — even when asked as "why does X fail?" | using-software-specialists |
| An existing plan file changed                                            | brainstorm                 |
| A vague idea turned into a spec — goal or scope still open               | brainstorm                 |
| Style, formatting or naming checked, or a style question answered        | style-checker              |
| Findings already reported worked through one at a time — an analyze-code report, PR review comments, an audit or issue list | addressing-findings |
| A review of existing code — a diff, branch, PR, module; quality, security, tech debt | analyze-code    |
| To be guided through testing a change by hand on a local or staging environment | guiding-manual-testing |
| The knowledge base written to or audited — ingest, update, lint, write a manual | knowledge-base      |
| An answer about this code base or what is documented about it, plans included | research             |
| A concept or term explained, with nothing to decide and nothing to change here | no skill |
| An answer from outside this code base — best practice, library choice, "is X true?" | using-software-specialists |
| Tests added to existing code, with no production change — "write tests for X", "cover this edge case" | writing-unit-tests |
| Anything else that changes code — feature, refactor, migration, perf, security fix, implementing a plan | using-software-specialists |

Tie-breakers:
- "Review and fix" is analyze-code. It only reports; the fixes follow its
  Suggested Next Actions. Going through the findings it reported, one by one,
  is addressing-findings.
- Comments already on a PR are addressing-findings. A fresh review of that PR
  is analyze-code.
- Explaining how something works is research. Explaining why it is broken is
  using-software-specialists.
- Tests only is writing-unit-tests. Tests that come with a code change — a fix,
  a feature, a refactor — are using-software-specialists, which loads
  writing-unit-tests itself.
- If you cannot state acceptance criteria for a requested change, it is
  brainstorm.

## Skills that are not entry points
The <available-skills> list at the end of this prompt also holds
coding-discipline, designing-interfaces and test-driven-development. The entry
skills load them at the step that needs them. Their descriptions say when they
apply inside that workflow — they are never a reason to load one first.
"Implement X" and "fix X" go to using-software-specialists, even though
test-driven-development's description names them.

A skill in that list not named in this section was added by the user. Load it
directly when its description fits the request better than any row above.

## Examples
"How does session compaction work?"                   → research
"Is there a plan for PROJ-1234?"                      → research
"Why does TestLoad fail on CI?"                       → using-software-specialists
"Which Go TUI library should we use?"                 → using-software-specialists
"What is the difference between a mutex and a channel?" → no skill
"I'd like plugin support, not sure what shape yet"    → brainstorm
"Add a rollback step to plans/proj-12.md"             → brainstorm
"Implement plans/proj-12.md"                          → using-software-specialists
"Review my branch before I open the PR"               → analyze-code
"Go through those findings one at a time"             → addressing-findings
"Address the comments on PR #12"                      → addressing-findings
"Does internal/config follow Go naming conventions?"  → style-checker
"Add a --verbose flag"                                → using-software-specialists
"Write tests for config.Load"                         → writing-unit-tests
"Fix the nil panic in Load and add a test for it"     → using-software-specialists
"Help me check the new setup flow on my machine"      → guiding-manual-testing
"Update the wiki with what we just changed"           → knowledge-base
"Rename cfg to conf in load.go"                       → no skill

## During the work
- The entry skill runs the work and names the other skills to load, and when.
  Do not load them ahead of it.
- When a skill hands off — analyze-code to using-software-specialists,
  addressing-findings to using-software-specialists, guiding-manual-testing to
  using-software-specialists, brainstorm to planning — load the skill it names.
- A follow-up that continues the same work stays in the loaded skill. Route
  again only when the request changes kind — research turning into "now fix it".
- After /compact or /clear, load the entry skill again before continuing — the
  conversation is gone, so you no longer know which skill was running, whatever
  you knew a moment ago.
- A skill lists the files it ships. Read the ones it tells you to with
  skill_load_file — naming a reference file is not reading it.
- Only a rename, a typo or a comment-only edit skips the table. "Too small to
  need a skill" is not otherwise an exemption.
`
)

func buildSkills(r *BuilderRequest) string {
	var builder strings.Builder

	builder.WriteString(skillsStartTAG)
	builder.WriteString("\n")
	builder.WriteString(skillsBody)
	builder.WriteString(buildKnowledgeBase(r.KnowledgeBase))

	if r.WithClassifier {
		builder.WriteString(buildClassifier())
		builder.WriteString(buildGuard())
	}

	builder.WriteString(skillsEndTAG)
	builder.WriteString("\n")

	return builder.String()
}
