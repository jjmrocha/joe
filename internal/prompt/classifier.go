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
- Override only when you can name a specific fact that contradicts the answer.
  "I would have decided differently" is not one. No fact, no override — if you
  cannot name one, follow the answer. Report every override as:
  Classifier override — <checkpoint>: answered <answer>; I did <action> because <fact>.
  That line, verbatim — free prose in its place is a broken report.
- If a call fails, do not retry it. Decide yourself and report:
  Classifier unavailable — <checkpoint>: <error>.
- A checkpoint marked as repeating is asked at most 3 times; report when the
  cap stops it.

Checkpoints:

1. test-driven-development, "REFACTOR — Clean up without adding behavior"
   — repeating.
   After you have walked the six REFACTOR items, for each production function
   changed in GREEN:
   classify_yes_no
     instructions: "Should this function be refactored further?"
     input: the function and the test that drives it. One call per production
   function — a single call naming several functions does not satisfy this.
   Yes → refactor, run the suite, ask again. No → move on.

2. analyze-code, step 8 "Synthesize & deliver" — before the report is shown.
   For each finding that survived step 7's verification, and each finding
   from step 6's tooling:
   classify_choice
     instructions: "What is the severity of this finding?"
     input: the finding's evidence line, its impact, and where the code
     runs — its role and exposure ("request handler, public endpoint",
     "test helper").
     options: Critical, High, Medium, Low — each described with the skill's
     severity scale.
   The selected option is the finding's severity. Order, group and cap the
   report by it.

3. designing-interfaces, when the four-line contract is written and you
   are about to hand off to test-driven-development — before the first
   test is written — repeating.
   classify_score
     instructions: "How deep is this interface?"
     input: the module's purpose in one line, the dependencies it touches
     (I/O, clock, randomness, network, or none), the four-line contract and
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
