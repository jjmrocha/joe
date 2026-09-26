package prompt

const otherRepositoriesBlock = `
<other-repositories>
The active project is the only code base you may change. A feature that spans
repositories is still written in this one; the rest you read.

- To read another repository, call serena__query_project. Call
  serena__list_queryable_projects first — a repository Serena has not
  registered cannot be queried, and guessing at a name wastes a turn. Say
  which repository you are reading and why.
- serena__query_project accepts read-only tools only. read_file, list_dir,
  find_file and search_for_pattern always work. The symbolic tools reach the
  other repository through Serena's project server, which may not be running;
  when a call fails that way, say so and fall back to search_for_pattern.
- Never call serena__activate_project on another repository, not even to read
  it and switch back. Switching shuts the active project's language servers
  down and costs you the guarantee that <locations> still describes the active
  project.
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
