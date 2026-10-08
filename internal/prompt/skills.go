package prompt

import (
	"strings"
	"unicode/utf8"

	"github.com/jjmrocha/joe/internal/skills"
)

const (
	skillsStartTag = "<skills>"
	skillsEndTag   = "</skills>"

	routesWantsHeader = "The user wants..."
	routesLoadHeader  = "Load"

	skillsIntro = `Skills are how you work, not reference material. Every request, in this order:

1. Route it with the table below: load the entry skill, or none.
2. Follow the loaded skill.

Every request that will read, change, judge or document code starts with one
entry skill, loaded with skill_load before you explore anything. Only a rename,
a typo or a comment-only edit skips the table; "too small to need a skill" is
not otherwise an exemption.

Announce the route before loading the entry skill:
"Route: <the row, in a few words> → Loading <skill>". Announce every other skill
you load: "Loading <skill>". The skill_load call goes in the same turn as its
announcement — an announced skill you never load is a skipped skill.

## Pick the entry skill
Route on what the user wants to end up with, not on the words they use. Rows run
from specific to general: take the first that matches, top to bottom.

`

	tieBreakers = `
Tie-breakers:
- "Review and fix" is analyze-code. It only reports; the fixes follow its
  Suggested Next Actions. Going through the findings it reported, one by one,
  is addressing-findings.
- Comments already on a PR are addressing-findings. A fresh review of that PR
  is analyze-code.
- Explaining how something works is research. Explaining why it is broken is
  using-software-specialists.
- Tests only is writing-unit-tests. Tests that come with a code change follow
  that change's route; the skill it lands on loads writing-unit-tests itself.
- A precise request is still brainstorm: it confirms you understood it and
  surfaces what the user did not consider. Precision shortens the
  brainstorm; it never skips it.
`

	nonEntrySkills = `
## Skills that are not entry points
coding-discipline, designing-interfaces and test-driven-development are never
entry skills. The entry skills load them at the step that needs them.

For every skill named in <skills>, the table decides the entry. Their
descriptions in <available-skills> only say when a step inside another skill
applies — never a reason to load one first:
- "Implement X" and "fix X" go to using-software-specialists, even though
  test-driven-development's description names them.
- writing-unit-tests is an entry only for tests with no production change. Its
  "when modifying code" clause is a step inside another skill.

A skill in <available-skills> not named anywhere in <skills> was added by the user.
Descriptions route only those: load one directly when its description fits the
request better than any row above.
`

	examplesHeading = `
## Examples
A routed request is one turn: the announcement, then the skill_load call — a
real tool call, never written out as text:

  "Which Go TUI library should we use?"
  Route: an answer from outside this code base → Loading using-software-specialists
  → tool call skill_load, skill_name "using-software-specialists"

The list below is shorthand for that turn — request → skill to load:
`

	followUps = `
Follow-ups:
(after research) "ok, now fix it"                    → using-software-specialists
(after analyze-code) "go through them"               → addressing-findings
(during brainstorm) "what about rate limits?"        → stay in brainstorm
`

	duringTheWork = `
## During the work
- This prompt's routing and rules win over a loaded skill. A skill's
  "When NOT to Use", or a step that contradicts this prompt, does not send you
  elsewhere or override it — follow this prompt, and the rest of the skill as
  written.
- The entry skill runs the work and names the other skills to load, and when.
  Do not load them ahead of it.
- When a skill step says to load another skill, load it at that step.
  Skipping it is the same failure as skipping the entry skill.
- When a skill hands off — analyze-code to addressing-findings or to
  using-software-specialists, addressing-findings to using-software-specialists,
  guiding-manual-testing to using-software-specialists, brainstorm to
  planning — load the skill it names.
- Apply each fix approved in addressing-findings through
  using-software-specialists' Implementation phase: coding-discipline,
  test-driven-development, and designing-interfaces when an interface changes.
- Load a skill only when its text is not already in this conversation — a
  routed request or a /<skill> command for a skill already loaded means
  following that copy, not loading it again.
- A follow-up that continues the same work stays in the loaded skill. Route
  again only when the request changes kind — research turning into "now fix it".
- A skill lists the files it ships. Read the ones it tells you to with
  skill_load_file, giving the skill's name and the file's path — naming a
  reference file is not reading it.
- When a step needs a tool this session does not have — a web search to answer
  from outside the code base, say — tell the user which tool is missing and
  stop there; never answer that step from memory instead.
`
)

const noSkill = "no skill"

type route struct {
	wants  string
	load   string
	needKB bool
}

var routes = []route{
	{wants: `A skill they named ("use brainstorm", "run analyze-code"), or a message that starts with /<skill-name> ("/research how does Load work") — a request to use that skill on the rest of the message. If it is coding-discipline, designing-interfaces or test-driven-development, route with the rows below and load the named skill at its step`, load: "that skill"},
	{wants: "The knowledge base written to or audited — ingest, update, lint, write a manual", load: skills.KnowledgeBase, needKB: true},
	{wants: "Findings already reported worked through one at a time — an analyze-code report, PR review comments, an audit or issue list", load: skills.AddressingFindings},
	{wants: "A review of existing code — a diff, branch, PR, module; quality, security, tech debt", load: skills.AnalyzeCode},
	{wants: "To be guided through testing a change by hand on a local or staging environment", load: skills.GuidingManualTesting},
	{wants: "Style, formatting or naming checked, or a style question answered", load: skills.StyleChecker},
	{wants: `Tests added to existing code, with no production change — "write tests for X", "cover this edge case"`, load: skills.WritingUnitTests},
	{wants: `Something broken fixed or explained — bug, failing test, crash, vulnerability, CI or build failure — even when asked as "why does X fail?"`, load: skills.UsingSoftwareSpecialists},
	{wants: "An existing plan file changed", load: skills.Brainstorm},
	{wants: `An existing plan implemented — "implement plans/X.md"`, load: skills.UsingSoftwareSpecialists},
	{wants: "A new capability, a change callers will see, or a restructuring — feature, new or widened interface, flag, contract change, refactor, migration — however precisely the user specified it", load: skills.Brainstorm},
	{wants: "Any other code change — perf tuning, dependency bump", load: skills.UsingSoftwareSpecialists},
	{wants: "An answer about this code base or what is documented about it, plans included", load: skills.Research},
	{wants: `An answer from outside this code base — best practice, library choice, "is X true?"`, load: skills.UsingSoftwareSpecialists},
	{wants: "A concept or term explained, with nothing to decide and nothing to change here", load: noSkill},
}

type example struct {
	request string
	route   string
	needKB  bool
}

var examples = []example{
	{request: `"How does Load resolve the profile?"`, route: skills.Research},
	{request: `"Why does Load pick the wrong profile?"`, route: skills.UsingSoftwareSpecialists},
	{request: `"Is there a plan for PROJ-1234?"`, route: skills.Research},
	{request: `"/research how does Load resolve the profile?"`, route: skills.Research},
	{request: `"Why does TestLoad fail on CI?"`, route: skills.UsingSoftwareSpecialists},
	{request: `"Fix the SQL injection in the search handler"`, route: skills.UsingSoftwareSpecialists},
	{request: `"Fix the nil panic in Load and add a test for it"`, route: skills.UsingSoftwareSpecialists},
	{request: `"Bump golang.org/x/net to v0.40"`, route: skills.UsingSoftwareSpecialists},
	{request: `"Which Go TUI library should we use?"`, route: skills.UsingSoftwareSpecialists},
	{request: `"What is the difference between a mutex and a channel?"`, route: noSkill},
	{request: `"I'd like plugin support, not sure what shape yet"`, route: skills.Brainstorm},
	{request: `"Add a --verbose flag"`, route: skills.Brainstorm},
	{request: `"Add an Instructions() method to mcp.Client"`, route: skills.Brainstorm},
	{request: `"Refactor Load into smaller functions"`, route: skills.Brainstorm},
	{request: `"Use TDD to add a --verbose flag"`, route: "brainstorm, test-driven-development at its step"},
	{request: `"Add a rollback step to plans/proj-12.md"`, route: skills.Brainstorm},
	{request: `"Implement plans/proj-12.md"`, route: skills.UsingSoftwareSpecialists},
	{request: `"Review my branch before I open the PR"`, route: skills.AnalyzeCode},
	{request: `"Review PR #12 and fix what you find"`, route: skills.AnalyzeCode},
	{request: `"Go through those findings one at a time"`, route: skills.AddressingFindings},
	{request: `"Address the comments on PR #12"`, route: skills.AddressingFindings},
	{request: `"Does internal/config follow Go naming conventions?"`, route: skills.StyleChecker},
	{request: `"Write tests for config.Load"`, route: skills.WritingUnitTests},
	{request: `"Cover the empty-profile edge case in load_test"`, route: skills.WritingUnitTests},
	{request: `"Help me check the new setup flow on my machine"`, route: skills.GuidingManualTesting},
	{request: `"Update the wiki with what we just changed"`, route: skills.KnowledgeBase, needKB: true},
	{request: `"Write a manual for running joe"`, route: skills.KnowledgeBase, needKB: true},
	{request: `"Rename cfg to conf in config.go"`, route: noSkill},
}

func buildSkills(r Request) string {
	withKB := r.KBPath != ""

	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(skillsStartTag)
	builder.WriteString("\n")
	builder.WriteString(skillsIntro)
	writeRoutes(&builder, withKB)
	builder.WriteString(tieBreakers)
	builder.WriteString(nonEntrySkills)
	builder.WriteString(examplesHeading)
	writeExamples(&builder, withKB)
	builder.WriteString(followUps)
	builder.WriteString(duringTheWork)
	builder.WriteString(skillsEndTag)
	builder.WriteString("\n")

	return builder.String()
}

func writeRoutes(builder *strings.Builder, withKB bool) {
	writeRow(builder, routesWantsHeader, routesLoadHeader)
	builder.WriteString("|---|---|\n")

	for _, row := range included(routes, withKB, func(r route) bool { return r.needKB }) {
		writeRow(builder, row.wants, row.load)
	}
}

func writeRow(builder *strings.Builder, wants, load string) {
	builder.WriteString("| " + wants + " | " + load + " |\n")
}

func writeExamples(builder *strings.Builder, withKB bool) {
	rows := included(examples, withKB, func(e example) bool { return e.needKB })

	requestWidth := 0
	for _, row := range rows {
		requestWidth = max(requestWidth, width(row.request))
	}

	for _, row := range rows {
		builder.WriteString(pad(row.request, requestWidth) + " → " + row.route + "\n")
	}
}

func included[T any](items []T, withKB bool, needKB func(T) bool) []T {
	var kept []T

	for _, item := range items {
		if withKB || !needKB(item) {
			kept = append(kept, item)
		}
	}

	return kept
}

func pad(text string, to int) string {
	return text + strings.Repeat(" ", to-width(text))
}

func width(text string) int {
	return utf8.RuneCountInString(text)
}
