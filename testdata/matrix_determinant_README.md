# Physics determinant macro aliases

Pinned [PhysicsMappings.ts264–266](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMappings.ts#L264)
registers three one-argument macros:

| Command | Original expansion |
| --- | --- |
| `matrixdeterminant` | `\vmqty{#1}` |
| `mdet` | `\vmqty{#1}` |
| `smdet` | `\svmqty{#1}` |

The implementation adds these registrations to the existing macro map. It
preserves GetArgument ownership, caller substitution charging, late-bound
helper lookup, and the priority of user operators and paired delimiters.
There is no new determinant algorithm, matrix parser, or child-budget reset.
The helper names resolve normally through the existing Physics matrix macros.

For example, `\Huge\mdet{a&b\\c&d}\quad\smdet{a&b\\c&d}` renders a full-size and
a small determinant with bar fences. The parser-owned row driver and source
MatrixQuantity behavior were prerequisites merged in PR 154; the new aliases
reuse that behavior.

## Complete original references

The source is MathJax 3.2.2 commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`, rendered using the frozen D2 runtime.
The explicit inventory has 898 distinct TeX/display inputs: 414 original alias
controls, 300 additional parser/budget controls, 100 whitespace/helper controls,
and 102 independent controls, with 18 overlaps. It contains 696 valid original
SVGs, 182 original error SVGs, and 20 original runtime exceptions. Every input is retained in the exact
fixture or the raw observation file; the generator never filters based on Go.

Against merged PR 154, 804 complete-original SVG references are exact: 608 fixes
and 196 unchanged controls, with no formerly exact regression. Coverage includes
all three names, grouped and real unbraced operands, scripts and pending items,
font/style/text-lap callers, generated matrices, helper and alias overrides in
both declaration orders, and effective caller/child macro-budget boundaries.
The six initial joined spellings such as `\mdetx` remain explicitly labeled
undefined-command controls. Separately delimited inputs establish actual
unbraced determinant dispatch; no original string is rewritten.

The aliases and their vmqty/svmqty helpers each incur one caller macro charge.
Repeated-call, same-parser, and genuine-child controls use supported
DeclareMathOperator definitions. They distinguish the two caller expansions
from MatrixQuantity's independent child parser and its array-opening charge.
The raw literal helpers are not treated as budget-equivalent near a limit:
removing the determinant alias removes one caller charge.

Compared with merged PR 157 (`42932ee918ca4db7f3edf9da30e59b0c9cd61f86`),
780 strict inputs are first complete-original publications and 24 already have
full references. No prior raw input is promoted; all overlapping originals agree.
The 30 qualified safe-XML controls are also first complete-original publications
and are counted separately from strict SVG assertions. These provenance counts
are separate from the merged-154 rendering baseline; the rebased source still
requires current-base validation before publication.

## Retained boundaries

The residual file retains 64 raw observations and 30 separately qualified
safe-XML controls. The safe-XML controls differ only in the mandatory escaping
of the single bare ampersand in `data-mjx-error="Misplaced &"`: 12 are newly
reachable and 18 are unchanged. Their tests retain the entire original SVG and
assert exactly that one attribute substitution. They are not strict byte-exact
references.

The 64 other raw observations are separate from passing assertions. All 48 changed
raw targets have complete-original-equivalent direct helper controls, and the
candidate output equals the helper's unchanged merged-154 output. These
bindings are retained with their original, baseline, and candidate SVGs.

- Twelve changed valid BOM inputs and eight unchanged helper inputs expose the
  inherited generic Macro GetArgument whitespace boundary. JavaScript GetNext
  accepts BOM as whitespace; the existing Go reader does not. This is a shared
  mandatory-reader issue, not a determinant-specific parsing rule.
- Twenty NEL inputs cause original runtime exceptions. Twelve are newly
  reachable aliases and eight are unchanged helpers. They are not counted as
  passing SVG references or as original-valid behavior.
- Twenty-four original-error targets preserve inherited unbraced frac/text and
  commented environment-capture diagnostics. Their complete helper bindings
  demonstrate the existing source-boundary difference without accepting Go
  output as an original golden.

The source map registration and the inherited helper limitations remain
separate claims. No normalized or shortened SVG is used for exact assertions.

## Regeneration

```sh
node --jitless testdata/generate_matrix_determinant.cjs PINNED_ASSETS
go test ./... -run TestMatrixDeterminantOriginalReferences
```

The generator verifies the three frozen asset SHA256 hashes, starts a fresh
original VM per conversion, and regenerates exact SVGs, raw originals, and
literal-control originals. JSONL uses LF framing and preserves literal Unicode
line separators. Public counts exclude any augmented-package configuration.
