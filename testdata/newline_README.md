# Closing cell events and row separators

The source is MathJax 3.2.2 commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`.
[BaseMappings](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMappings.ts)
registers `newline` as `CrLaTeX` with `nobrackets=true`. The existing `\\` uses
ordinary `CrLaTeX`, and `cr` uses `Cr`. These handlers deliver a closing
[CellItem](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseItems.ts),
not an immediate space or a raw text split.

For example, `a\over b\\c` must complete the fraction before continuing with
`c` on the surrounding row. The old parser put `c` in the denominator.
`\begin{matrix}a\text\\b&c\end{matrix}` must let `\text` consume its argument
before deciding whether a row separator exists. The old raw scanner split it
before argument parsing. Macro-generated separators and overridden separator
commands have the same ownership requirement.

## Source semantics

[BaseMethods.CrLaTeX](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts)
reads only an immediately adjacent `*` and then an immediately adjacent optional
bracket dimension. `newline` consumes neither. Dimensions are validated before
the CellItem is pushed. Empty `[]` differs from explicit `[0pt]` and invalid
`[ ]`. The post-action produces an optional depth space and a newline space only
after the close event has reached its open recipient. Real arrays instead close
an entry and row, then adjust row spacing using their initial source spacing.

Pending Fn, Not, Dots, Prime, Nonscript, Style and Position items reduce before
the close event is delivered. OverItem produces its fraction and replays the
same CellItem. Groups, Left/Right and AutoOpen owners swallow a linebreak cell
and remain open. A single Braket can close on an earlier final MML and replay
the event to its caller. Its copied lexical environment then ends; a denominator
continuation in the same owner retains its environment. Plain `cr` and Entry
remain errors when the actual open owner is not an array.

The shared array driver resumes the expanded parser source after a delivered
Entry/CR. Each array kind retains its EndEntry, EndRow, EndTable, tagging, HFill,
column-count and spacing callbacks. Cells share configuration, Stack.global and
the parser macro budget while resetting their lexical environment. Explicit
empty rows and pending empty final rows remain distinct. Matrix command bodies,
ordinary arrays, EqnArray, Flalign, Multline, the cases command, Mathtools starred cases text cells and CD row endings
now use this event path. CD's separate `@` arrow tokenizer and environment-body
capture are retained; this change does not claim to replace those scanners.

D2's frozen package order selects **AMS HandleShove**, overriding the Mathtools
handler. The selected handler sets a property on an actual MultlineItem and
consumes no argument. A following group or bracket stays on the caller;
`multlined` is rejected. This distinction preserves the grouped operand and
prevents a new regression introduced by using the unselected source handler.

## Complete original references and qualifications

The fixture inventory combines the initial source audit, additional array and
ownership controls, independent review, macro budgets, explicit-owner controls,
and the complete changed cohort from the historical residual replay. Every
input remains in either `newline_mathjax_3_2_2.json` or
`newline_residuals.json`; the generator does not filter based on Go output.
The recorded fixture metadata reports its exact baseline, counts and partition.

The current inventory has 9,299 inputs: 9,071 complete-original references and
228 preserved residuals. Against merged PR 153, 5,461 references change from an
incorrect baseline to the exact original SVG. All 7,859 previously reviewed
candidate outputs remain byte-identical after composing Bqty and text MtLap.
The added Bqty, MathLap/MtLap, equation-array and independent array/font
controls match the original SVGs exactly. MtLap cases distinguish literal
text, embedded math and outer-array ownership; each input retains its
provenance and ownership annotations.

A subsequent token-boundary audit identified 1,512 earlier inputs containing
joined control words such as `\newlinec` or `\crb`, including occurrences in
raw text, command arguments and comments. They remain exact-input error or
text/argument controls; their results do **not** establish dispatch of
`\newline` or `\cr`. Their original SVGs are never rewritten. Including the
later MathLap and Bqty controls, 2,184 such inputs remain explicitly marked.
Separately captured versions with explicit keyword boundaries retain the input's original
generator-family provenance. The 38 additional valid residuals have both
original and candidate SVGs identical to previously reviewed residual
counterparts, and each has a complete-original-exact literal control. The
existing `\\` forms and genuinely terminated keyword forms are unaffected.
Keyword occurrences inside text, command arguments or comments remain ownership
controls; a separated spelling alone does not establish direct row dispatch.

The public audit includes starred and ordinary options, all supported array
families, generated and overridden commands, text/argument ownership, comments,
scoped fonts/colors/styles, root and script callers, pending items, and
999/1,001 macro calls split across row separators. The latter produce twelve
valid originals and twelve original macro-limit errors, proving that row
resumption does not reset the caller's budget.

Raw residuals retain the original, historical baseline and candidate SVGs.
They are not passing parity references. The changed valid families were also
inspected at a common physical scale and checked against direct source controls:

- Quantity and Braket/Set still have inherited Over/Style finalization and
  lexical-state gaps. Explicit completed-owner formulas have **identical full
  original SVGs**, and the candidate renders those formulas exactly. Those
  bindings are retained and tested. No-separator controls reproduce the same
  inherited differences without the new row behavior.
- MathFont around subarray now resets to the original glyphs and bounds. A
  remaining 0.1-unit table-row translation difference reproduces without the
  outer font wrapper; it is a separate renderer-rounding discrepancy.
- Escaped literal LF in control-sequence names retains the existing GetCS and
  declaration-name validation differences. Removing the Position command gives
  byte-unchanged baseline/candidate controls. Original-error and original-valid
  observations remain distinguished.
- Original error SVGs containing `Misplaced &` serialize a bare ampersand in the
  error attribute. Go retains safe XML. Qualified tests permit only that exact
  attribute escape and compare the rest of the complete SVG unchanged.

Independent source review covers close-item resumption, Braket environment
lifetime, selected AMS shove handling and array callback ownership. The saved
independent public corpus is included in the fixture union. Historical raw
inputs were replayed without changing the originals. Among 4,638 published raw
inputs, 984 become complete-original exact and 26 changed residual outputs
remain identical to previously qualified candidates. Three observed changes
in cancel SVGs are only nondeterministic data-attribute order: two apparent
fixes and one apparent regression. Both binaries reproduce both orders over
64 repeated renders of each of the 16 cancel controls, with full XML trees
and geometry equal to the original. Those three artifacts are excluded from
the source-fix and regression counts. Final gates must be checked at handoff;
the source checkpoints and discovery comparisons remain separate audit receipts.

## Regeneration

```sh
node --jitless testdata/generate_newline.cjs PINNED_ASSETS
go test ./... -run 'TestNewline'
```

The generator verifies all three frozen D2 asset hashes, starts a fresh original
VM per conversion, and captures complete SVGs with font cache none, em=16, ex=8
and the stored display flag. It also regenerates raw originals and literal
bindings while preserving historical Go observations. JSONL framing uses only
LF, retaining literal Unicode line separators inside the saved inputs.
