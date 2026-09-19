package prompt

const kbConfigured = `
<knowledge-base>
Ignore any kb_path set anywhere above. This block is the only one that counts.

kb_path=%s

The knowledge base is a folder of Markdown outside every repository. The file_
tools reach it and nothing else: file_read, file_write, file_edit, file_list and
file_delete take paths relative to its root, and file_workdir reports that root.

- The repository is Serena's. Never reach for a file_ tool to read or change
  code, and never expect a serena__ tool to see the knowledge base.
- Load the knowledge-base skill before reading or writing the knowledge base. It
  owns the layout, the page format and the rule that a delete needs the user's
  approval.
- The knowledge base records what is intended and what exists; the code is the
  truth. When the two disagree, say so and believe the code.
</knowledge-base>
`

const kbNotConfigured = `
<knowledge-base>
Ignore any kb_path set anywhere above. This block is the only one that counts.

kb_path=

No knowledge base is configured and the file_ tools are not registered. Do not
load the knowledge-base skill — not even when the skill table above routes a
request to it — and do not guess a path. If a request needs the knowledge base,
say it is not configured and that kb-path in the profile at ~/.config/joe is
where to set it.
</knowledge-base>
`
