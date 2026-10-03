package prompt

const classifierBlock = `
<classifier>
The classify_yes_no, classify_choice and classify_score tools reach a
calibrated classifier. At each checkpoint below, calling it is required.

- At a checkpoint the classifier makes that one decision. The block replaces
  the loaded skill's own rule for it; the rest of the skill still applies.
- Reading the answer: yes_probability of 0.5 or more is yes; selected is the
  choice; a score is compared with the checkpoint's cutoff.
- Input: the smallest self-contained context that lets the classifier decide.
  It sees nothing else — no files, no conversation. Never include secrets,
  keys or tokens; redact them.
- What a checkpoint names as input — source, a contract, an evidence line —
  goes in verbatim, labels included: copied from what you read in this
  session, never described, summarised or rebuilt from memory — read all of
  it before you send it. Leave out test results, lint status and your own
  verdict: the classifier judges the artifact, not your opinion of it.
- Instructions, labels and options go in as the checkpoint gives them and
  name the outcomes neutrally. Never write your hypothesis, the item or its
  context into one — that belongs in input; otherwise the answer measures
  the label, not the input.
- Override only when you can name a specific fact that contradicts the answer.
  "I would have decided differently" is not one. No fact, no override — if you
  cannot name one, follow the answer. Report every override as:
  Classifier override — <checkpoint>: answered <answer>; I did <action> because <fact>.
  That line, verbatim — free prose in its place is a broken report.
- If a call fails, do not retry it. Decide yourself and report:
  Classifier unavailable — <checkpoint>: <error>.
- A checkpoint not marked as repeating is asked once per item: the first
  answer stands, even if you think a better input would change it. A
  checkpoint marked as repeating is asked at most 3 times; report when the
  cap stops it.
- Before you move past a checkpoint, write:
  Classifier — <checkpoint>: <n> calls for <n> <functions|findings|interfaces>.
  The two numbers must match; if they do not, make the missing calls first.

Checkpoints:

1. test-driven-development, "REFACTOR — Clean up without adding behavior"
   — repeating.
   After you have walked the six REFACTOR items, for each production function
   changed in GREEN:
   classify_yes_no
     instructions: "Would a senior software engineer refactor this function
     further? Judge it against these principles:
       - Duplication: the same knowledge expressed in two places.
       - Naming: each name says what it is, not how it works.
       - Cognitive complexity: nesting depth, branch count, boolean
         operators per condition; the function reads top-to-bottom in
         one screen.
       - Single responsibility: the function does one thing.
       - No side effects: no mutation of arguments, globals or receiver
         state that the name doesn't advertise.
       - Dead code and speculative generality: none is left."
     input: the function's source and the source of the tests that drive it,
     both verbatim. One call per production function — a single call naming
     several functions does not satisfy this.
   Yes → refactor, run the suite, ask again. No → move on.

2. analyze-code, step 8 "Synthesize & deliver" — before the report is shown.
   For each finding that survived step 7's verification:
   classify_choice
     instructions: "What is the severity of this finding?"
     input: the finding's evidence line, its impact, and where the code
     runs — its role and exposure ("request handler, public endpoint",
     "test helper").
     options: Critical, High, Medium, Low — each described with its row of
     the skill's Severity Scale table, copied word for word, examples
     included.
   The selected option is the finding's severity. Order, group and cap the
   report by it.
   Skip the call, and keep the skill's severity, for a finding whose severity
   the skill sets by rule — intent conformance (step 2), a breaking change
   (step 3), convention drift or duplication (step 4), a hardcoded secret —
   and for step 6's tooling output. Count only the findings you call for.

3. designing-interfaces, when the four-line contract is written and you
   are about to hand off to test-driven-development — before the first
   test is written — repeating.
   classify_score
     instructions: "How deep is this interface?"
     input: the module's purpose in one line, the dependencies it touches
     (I/O, clock, randomness, network, or none), the four-line contract
     verbatim with its CALLER LEARNS, HIDDEN, SEAM and TEST CALLS labels, and
     the proposed signatures.
     levels:
       0 shallow — HIDDEN is empty or one clause; a conduit
       1 leaky — CALLER LEARNS is longer than HIDDEN
       2 adequate — hides more than it exposes, but TEST CALLS reaches past
         CALLER LEARNS or a needed seam is missing
       3 deep — small CALLER LEARNS, substantial HIDDEN, the test calls only
         the interface
   Below 2.0 → redesign, then ask again. 2.0 or above → implement —
   hand off to test-driven-development.
   The score replaces the read-back's verdict; what the read-back found still
   guides the redesign.

Outside these checkpoints you may call the classifier for any other judgement
a labelled answer settles — a yes/no, a pick from named options, a rating on
ordered levels. There the answer is advice, not a decision: weigh it against
what you know, and say when you went against it. The input rules above still
apply.
</classifier>
`

func buildClassifier() string {
	return classifierBlock
}
