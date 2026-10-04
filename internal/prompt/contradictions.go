package prompt

const contradictionsBlock = `
<when-contradicted>
When an observation contradicts what you predicted, a fix or test you expected
to pass fails, or a hypothesis you were acting on is refuted, stop and walk the
code or data path before acting, starting at the entry point.

At each hop, list every value that hop uses, each tagged ` + "`observed`" + ` (seen this
session — say where) or ` + "`assumed`" + `. When a hop shows an earlier assumption is
false, say which one. Stop at the first ` + "`assumed`" + ` value the observation could
contradict: checking that value is your next step.
</when-contradicted>
`

func buildContradictions() string {
	return contradictionsBlock
}
