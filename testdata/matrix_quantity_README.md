# Physics matrix quantities and their parser boundaries

The original is MathJax 3.2.2 commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`, rendered with D2's frozen assets.
[PhysicsMethods.MatrixQuantity](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMethods.ts#L706)
and [PhysicsMappings](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMappings.ts#L251)
define four handlers and eight convenience macros. The macros remain late-bound:
user replacements of `mqty`, `smqty`, `left`, `right`, `begin`, and `end` affect
the generated expression through normal dispatch.

MatrixQuantity creates an independent TeX parser for an exact array expansion.
Ordinary quantities use `array` with an explicitly empty column specification;
small quantities use `smallmatrix` with an initial empty group. Braces omit
outer fences; parentheses, brackets, and bars add the corresponding fences.
Only starred parentheses change to group delimiters. Unsupported following
input produces an empty fenced matrix and stays in the caller. GetStar/GetNext
use JavaScript whitespace; GetUpTo tracks braces, not nested punctuation.
Completed inferred rows are delivered as individual items through the existing
child-parser path, preserving script and pending-item ownership.

This source correction and the parser-owned row events in `newline_README.md`
are coupled. Before the row-event correction, generated rows could remain in a
single cell. Correcting their rows alone exposed the old MatrixQuantity path's
full-size metrics in small matrices. Correcting only the fences could also move
a still-flattened column farther from the original. The combined implementation
matches all 24 generated small-matrix examples and all 48 independently captured
literal/source expansions exactly. The intermediate failures are retained in
the development audit; neither intermediate is presented as a correct result.

## Necessary shared source boundaries

[BaseMethods.BeginEnd](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts)
charges the opening environment after reading and validating its name. A
registered closing environment with a falsy first handler argument returns an
EndItem without a charge; unknown, custom, and truthy-handler extra ends charge
before dispatch. Retained source maps identify this distinction. Matched
truthy-handler ends still have an inherited capture-path limitation, documented
in the raw controls; this change does not claim complete BeginEnd parity.

The new opening charge requires the actual source-created children to have
independent budgets. ParseArg, MathFont, root indices, arrow labels, separate
under/overbracket copies, Physics exponent/Quantity/Expectation children,
Physics braket expansions, and CD labels now save, reset, and restore their
macro count at those specific boundaries. Generated Physics operands share
one child budget. Groups, font declarations, styles, continuations, Eval and paired-delimiter
reinsertion keep the caller's budget. The generic `parseString` and shared
expansion helpers are deliberately unchanged. Counter controls at 998–1,001
calls distinguish these cases, including the previously exact child examples
that would otherwise regress after the opening charge.

[BaseMethods.Array](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts)
materializes `columnalign` even when it is empty. Normal matrix defaults remain
`c`; authored `array{}` remains explicitly empty. MtMatrix/MtSmallMatrix pass
an optional alignment to Array, which reads the next argument only when that
value is literally empty. Thus `[]` differs from absent, `[ ]`, `[x]`, and
`[{}]`. This fallback occurs before environment-body capture and preserves the
active `\begin` diagnostic and JavaScript whitespace behavior.

## Original references and retained differences

The explicit, deduplicated inventory contains 2,892 inputs: 2,674 complete
original SVG assertions and 218 raw observations. The exact cases include
2,286 valid expressions and 388 original error SVGs. Against merged PR 156,
1,764 become exact and 910 are unchanged exact controls. The raw inventory
contains 32 original runtime exceptions; none is counted as a passing SVG
reference. The generator regenerates every original independently of Go and
does not discard inputs based on a parity result.

Coverage includes all core and convenience spellings, delimiters/stars,
mandatory and optional arguments, fallback tokens and whitespace, unbraced
scripts, scoped fonts/styles, nested arrays, shared/child macro budgets,
registered command overrides and both declaration orders. The independent
empty-option review contains 704 inputs: 616 exact, 400 fixed, and no formerly
exact regression. All 160 caller/default controls are exact, including the
16 formerly wrong-glyph empty-option cases.

Thirty raw valid expressions become renderable where the baseline produced an
error. Their glyph paths agree with the original; their remaining inner-array
rule/frame discrepancy exactly reproduces in unchanged standalone controls.
The complete inner table, including child transforms and lines, matches those
controls after excluding its outer placement. This remains a separate column
rule/frame issue, not an exact-reference claim. Other retained observations
are unchanged inherited cases, original-runtime inputs, or original-error
capture/override diagnostics. Raw original and historical Go outputs remain
unmodified.

Together with the newline fixture, this branch contains 12,247 distinct exact
inputs (9,573 plus 2,674, with no overlap). Against all complete-original fixture
records on merged PR 156, 11,235 are first complete-original publications,
914 promote earlier raw observations, and 98 were already complete references.
All overlapping original strings agree. The two inventories contain 12,693 distinct inputs in total, including raw
observations. Safe-XML qualified controls in the newline fixture remain
separate from the exact counts: 70 are first publications and two were
previously raw observations.

The main156 historical replay covers 4,650 unique raw inputs: 984 genuine
source fixes and 26 changed nonexact outputs identical to the previously
qualified newline candidates, with no rendering regressions. Cancel/enclose
data-attribute order is nondeterministic in the existing renderer. Repeated
renders of both binaries preserve complete XML and geometry; apparent exact
successes as well as apparent regressions from ordering are excluded from
source-fix counts. The original serialized strings are never normalized.

## Regeneration

```sh
node --jitless testdata/generate_matrix_quantity.cjs PINNED_ASSETS
go test ./... -run 'TestMatrixQuantityOriginalReferences|TestEnvironmentCounterBoundaries'
```

The generator verifies all three pinned asset SHA256 hashes and starts a fresh
original VM for every conversion. It captures the complete SVG with font cache
none, em=16, ex=8, and each stored display flag. LF-only JSONL framing preserves
literal Unicode line separators. Final gate results belong to the tested
composed commit, not any earlier standalone MatrixQuantity checkpoint.
