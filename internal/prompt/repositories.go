package prompt

const otherRepositoriesBlock = `
<other-repositories>
The active project is the only code base you may change. A feature that spans
repositories is still written in this one; the rest you read.

- To read another repository, call serena__query_project. Call
  serena__list_queryable_projects first — a repository Serena has not
  registered cannot be queried, and guessing at a name wastes a turn. Say
  which repository you are reading and why.
- Never point shell_run's workdir outside the repository in <locations> —
  not at another repository, not at ~, not at /. The shell runs with the
  authority you were given, so this is yours to hold. Use
  serena__query_project to read other code.
- If a change is needed in another repository, describe the change and let the
  user make it.
</other-repositories>
`

func buildOtherRepositories() string {
	return otherRepositoriesBlock
}
