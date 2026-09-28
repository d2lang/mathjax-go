# Array horizontal rules and their frame-layout dependencies

These references use D2's frozen MathJax 3.2.2 component at
`ad8f5c21cb810236551da8c6512ba733e67357ee`. The public test compares complete,
unaltered original SVG strings. Separate residual files preserve differences;
they are not expected Go outputs and are not normalized into passing results.

## Source behavior

[BaseMethods.HLine](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts#L1298)
handles `hline` and `hdashline` as side effects on the actual ArrayItem. The
current cell must be empty and the array must be the current stack item. A
pending function, style or genuine child parser cannot receive a line on the
array's behalf. HLine produces no node and does not close an entry or row.
Completed rows determine whether the rule is a top frame or an internal line;
the final line at a shared boundary wins.

[ArrayItem.checkLines and createMml](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseItems.ts#L964)
move a terminal rowline into the bottom frame and stop a repeated interior rule
from spilling into a following gap. Frame sides are a list, including repeated
top sides. Exactly four entries select an mtable frame; other nonempty lists
produce a menclose. The full-frame dashed flag comes from the column definition.
Partial top/bottom frames are solid, including when requested with hdashline.
The original mtable rowlines spelling is retained; trimming repeated trailing
`none` entries affects only partial-frame padding.

The same owner state is carried through ordinary arrays, matrix commands, AMS,
Mathtools, CD, and NumCases. Finalization occurs after the actual rows and table
attributes are complete, before delimiters or the final MML item are delivered.
For NumCases this is before copying the closed table and applying its left
expression, as in CasesMethods.NumCases and EmpheqUtil.left. A resulting
menclose is not unwrapped to force the old table shape.

## Coupled source corrections

Newly renderable framed NumCases exposed three existing representation/layout
differences. The implementation follows the original contracts rather than
accepting the displaced formula as a residual:

* Empheq's copied top-row `setChildren(slice(0, 1))` clears the same inferred
  row before flattening it. The correction is local to that copy route;
  generic Node.SetChildren behavior is unchanged.
* SVGWrapper.handleAttributes consults per-node and global defaults, the source
  skip map, and existing SVG element attributes. Attributes without a default
  pass through; known MathML defaults do not leak to SVG. Class handling is a
  separate unchanged source gap.
* SVGmtr starts at the unscaled left frame offset, while individual cell
  positions still use the source scale calculation.

Restoring the default-aware attribute loop requires SVG-created operators to
have the same registered defaults as parser-created nodes. The MathML registry
was moved mechanically into `internal/mml`; TeX retains forwarding entry points.
Registration bodies, order and defaults are unchanged. Radicals, bevel slashes
and mfenced operators now use that shared factory.

[MmlMfenced](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/core/MmlTree/MmlNodes/mfenced.ts)
owns its fake open/close/separator nodes separately from its authored children.
Each receives the original incoming inheritance context, not the resolved
mfenced attributes, and its initial TeX class is a field rather than a
`texClass` property. Source class traversal interleaves these nodes with the
authored children. The temporary SVG row owns wrappers only; it has no semantic
children. Clone and repeated layout preserve those ownership distinctions.

[CommonWrapper.getScale](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrapper.ts)
reads explicit-aware `mathsize` only for tokens and mstyle. Other containers
read the inherited layer, including default/global prototype values. The
correction changes only this selection; script-level arithmetic and existing
CSS handling remain unchanged. This resolves the four retained own-size fence
differences after fake-node inheritance was corrected.

## Public inventory and measured results

There are 2,444 distinct TeX/display pairs: **2,412 complete original SVG
assertions** (1,662 valid renderings and 750 original TeX-error renderings) and
**32 unchanged raw observations**. Against actual main169
`dcdb525519b6a36b83b0eb2af303174e7bae6021`, the measured source
`acca3b7a37c4991c36810ae4695c72494ea85a1b` makes 1,398 of these inputs exact.
The remaining strict inputs were already exact. Original outputs from every
retained capture are preserved, including early preview failures.

The union contains the 2,190 owner/frame/attribute controls, 84 public root
controls, and distinct promoted published/upstream inputs. Coverage includes
top/bottom/interior lines, repeated rules, mixed column frames, empty/ragged
arrays, generated entries/rows, script/style/font recipients, dynamic command
overrides, table subclasses, NumCases copy paths, and ordinary unchanged
attribute consumers. Per-source counts overlap before this explicit union.

The raw file retains 24 token font/CSS differences and eight class-handler
differences (truthiness, JavaScript trim and splitting on ASCII spaces). All 32
candidate strings equal the measured baseline strings. Their complete original,
baseline and candidate objects remain separate; no test requires reproducing a
wrong baseline SVG. None is a newly changed valid residual.

Overlap was scanned against every tracked JSON or JSON.gz below any testdata
directory at actual main170 `8393014a3f2dd05d0b2881141c85a8153ddd333e`: 453 files.
Of the 2,412 strict inputs, **1,910 are first input publications, 196 promote
prior raw originals, and 306 already have full original references**. All prior
original strings agree; there are no ambiguous-only matches. These are
per-change input counts, not globally unique MathJax coverage. This overlap
base is newer than the measured main169 replay base; it is not a claim of a
compiled main170 candidate.

## Constructed MathML controls and replay boundaries

The separate internal fixtures contain 96 factory/fence observations: 88 strict
original SVGs and eight unchanged CSS residuals. Thirty-six cases additionally
assert original fake-node attributes/properties and actual Go inheritance/clone
ownership. Another 128 strict original SVGs exercise size selection across
tokens, styles, containers, scripts, fractions, roots, tables and fences. Eight
default-maction calls fail in the original lite DOM because addEventListener is
unavailable; those full original runtime observations remain compressed and
separate. They are not SVG assertions. The 216 strict constructed inputs are
distinct within these two files and are not counted as new public TeX inputs.

On the measured main169 comparison, all 2,190 owner/frame/attribute inputs have
2,158 exact outputs and the same 32 residuals; all 84 public root controls pass.
All 88 strict fence/factory and 128 scale controls pass, including the six
earlier registry-only fence regressions. The 4,326 upstream inputs have 30
genuine fixes, no regressions, and no changed nonexact results. The 6,438 complete
published residual inputs have 196 genuine fixes, no regressions, and no changed
nonexact results; 24 metadata-only records are excluded from this SVG count.

Cancel/enclose attribute insertion order produces apparent byte differences.
Nine retained controls were repeated 64 times in each binary: all 1,152 results
match the complete parsed original XML, including every attribute, text node
and child order. One apparent fix and three apparent regressions are therefore
excluded from the behavioral counts. This qualification is separate from all
strict SVG assertions, which perform no XML normalization.

The current receipts establish focused and prebuilt replay results. Final
current-base composition and full frozen-oracle/race/vet/WASM gates remain
required before publication.

## Regeneration and visual evidence

```sh
node --jitless testdata/generate_array_horizontal_rules.cjs PINNED_ASSETS
node --jitless internal/tex/testdata/generate_synthetic_operator_factory.cjs PINNED_ASSETS
go test ./... -run 'TestArrayHorizontalRuleOriginalReferences|TestSyntheticOperator|TestFencedIncoming'
```

Both generators verify the three frozen asset hashes and run fresh original
VMs in bounded subprocess batches. They never invoke Go, choose cases according
to candidate results, or normalize outputs. The public exact and raw files
regenerate byte-identically. Existing historical fixture bytes remain unchanged.

D2 omits MathJax's table-line CSS. Partial menclose frame sides have explicit
SVG strokes and provide direct D2 witnesses. Internal table lines and complete
table frames need a separately labeled render with the original MathJax CSS;
that supplemental image must not be described as a D2 rendering fix.
