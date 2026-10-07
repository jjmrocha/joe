package prompt

const classifierBlock = `
<classifier>
The classify_yes_no, classify_choice and classify_score tools reach a
calibrated classifier. At each checkpoint below, calling it is required.

- At a checkpoint the classifier makes that one decision. The block replaces
  the loaded skill's own rule for it; the rest of the skill still applies.
- Reading the answer: yes_probability of 0.5 or more is yes; selected is the
  choice, unless the checkpoint says otherwise.
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
- If a call fails, read the error. An error about the call's arguments
  (a missing or invalid field) is your mistake: fix the call and send it
  again — that retry does not count as an ask. Any other failure: do not
  retry. Decide yourself and report:
  Classifier unavailable — <checkpoint>: <error>.
- A checkpoint not marked as repeating is asked once per item: the first
  answer stands, even if you think a better input would change it. A
  checkpoint marked as repeating is asked again only after the item changed
  since its last ask, at most 3 times per item; report when the cap stops it.
- Before you move past a checkpoint, write one line naming every item and
  each answer it got, in order:
  Classifier — <checkpoint>: <item> <answer>[ → <answer>…]; <item> <answer>; …
  An item that should have been asked and is not on the line is a missing
  call: make it first.

Checkpoints:

1. test-driven-development, "REFACTOR — Clean up without adding behavior"
   — repeating.
   First write the skill's six REFACTOR lines, one per item — Duplication,
   Naming, Cognitive complexity, Single responsibility, No side effects,
   Dead code — each saying what you changed or "already clean". Then, for
   each production function changed in GREEN:
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
     input: exactly two labelled blocks:
       FUNCTION:
       <the function's source, verbatim>

       TESTS:
       <the source of every test that calls it, verbatim>
     When no test calls the function directly, TESTS holds the tests that
     reach it through its callers; when none do, TESTS is: none.
     One call per production function — a single call naming several
     functions does not satisfy this.
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

3. designing-interfaces, when the contract is written and you are about to
   hand off to test-driven-development — before the first test is written —
   repeating.
   classify_choice
     instructions: "Which interface principle does this design break most
     seriously? Pick sound only if it breaks none."
     input: the contract verbatim, with its WHAT, WHERE, INTERFACE and USE
     labels.
     options:
       no real work — WHAT only returns or exposes something the module
         holds
       complexity leaks — USE shows the caller filtering, rendering,
         ordering or interpreting what the module could do itself
       more than one responsibility — WHAT joins unrelated jobs with "and";
         a caller could want one part without the other
       unneeded inputs — an input the module already has or could default,
         or a whole object passed when one field is used
       poor names — the function or a parameter is vague, does not say what
         it is or does, or breaks the language's naming conventions
       avoidable errors — INTERFACE lists an error the function could
         handle itself, such as "not found" where an empty result would do
       easy to misuse — adjacent parameters of the same type that are easy
         to swap, or a rule the caller must remember that types could enforce
       exposes internals — INTERFACE or USE shows how the module stores
         its data
       wider than needed — it adds functions, methods or types no caller in
         USE needs
       creates its dependencies — it builds its own I/O, clock, network
         client or randomness instead of receiving them
       sound — breaks none of the above
   probabilities["sound"] of 0.5 or more → implement — hand off to
   test-driven-development. Below 0.5 → redesign around the most probable
   option other than sound, then ask again. When the cap stops it, show the
   user the contract and every answer it got, and ask before implementing.
   The answer replaces your own principles check; what that check found
   still guides the redesign.

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
