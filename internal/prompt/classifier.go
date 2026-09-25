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
  "I would have decided differently" is not one. Report every override as:
  Classifier override — <checkpoint>: answered <answer>; I did <action> because <fact>.
- If a call fails, do not retry it. Decide yourself and report:
  Classifier unavailable — <checkpoint>: <error>.
- A checkpoint marked as repeating is asked at most 3 times; report when the
  cap stops it.

Checkpoints:

1. test-driven-development, REFACTOR — repeating.
   After GREEN, for each production function changed in GREEN:
   classify_yes_no
     instructions: "Should this function be refactored further?"
     input: the function and the test that drives it.
   Yes → refactor, run the suite, ask again. No → move on.

2. analyze-code — before the report is shown.
   For each finding that survived verification:
   classify_choice
     instructions: "What is the severity of this finding?"
     input: the finding's evidence line and its impact.
     options: Critical, High, Medium, Low — each described with the skill's
     severity scale.
   The selected option is the finding's severity. Order, group and cap the
   report by it.

3. designing-interfaces, after the contract — repeating.
   classify_score
     instructions: "How deep is this interface?"
     input: the four-line contract and the proposed signatures.
     levels:
       0 shallow — HIDDEN is empty or one clause; a conduit
       1 leaky — CALLER LEARNS is longer than HIDDEN
       2 adequate — hides more than it exposes, but TEST CALLS reaches past
         CALLER LEARNS or a needed seam is missing
       3 deep — small CALLER LEARNS, substantial HIDDEN, the test calls only
         the interface
   Below 2.0 → redesign, then ask again. 2.0 or above → implement.
</classifier>
`
