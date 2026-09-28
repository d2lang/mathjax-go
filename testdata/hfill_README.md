# Array HFill alignment

`hfill_mathjax_3_2_2.json` contains 2,112 complete, unmodified SVG responses
from D2's frozen MathJax 3.2.2 bundle, upstream commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. There are 1,682 valid expressions
and 430 original error SVGs. Each expression uses a fresh VM with the
original D2 setup, `em=16`, `ex=8`, no font cache, and its recorded display
mode. The generator verifies the three asset hashes stored in the fixture.
No reference is synthesized by Go or normalized before comparison.

The primary sources are
[`BaseMethods.HFill`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts),
[`ArrayItem.EndEntry` and `EqnArrayItem.EndEntry`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseItems.ts),
[`MultlineItem.EndEntry`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ams/AmsItems.ts),
and [`AmsCdMethods.arrow` and `cell`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/amscd/AmsCdMethods.ts).
The base map binds `hfill`, `hfil`, and `hfilll` to the same method.

HFill emits no node. It records the current ArrayItem's node count, and is
an error unless the actual top item is an ArrayItem or subclass. Pending
functions, styles, positions, primes, negations, dots, fractions, Braket
owners, and AutoOpen fences therefore reject it before they reduce. Ordinary
groups and genuine child parsers cannot inherit the surrounding array.
SetFont does not push an item: font continuations preserve the preceding
array node count. User operator and paired-delimiter registrations retain
priority over the built-in handler.

ArrayItem.EndEntry right-aligns a cell when its first fill was at position
zero, left-aligns it when its last fill was at the final node count, and
centers it when both conditions hold. Interior fills do not override the
array's column alignment. The EqnArray and Multline subclasses override
EndEntry and intentionally ignore fill alignment; they still accept HFill
when their array item is on top.

CD uses the same ArrayItem alignment. A fill that emits no prefix nodes must
remain pending until the next actual cell closes: CD may close a padding
cell, or insert the arrow into the pending cell first. Its first-object-cell
strut also counts as a real node. Explicit `&` entries clear both lexical
state and fill positions. These rules preserve leading, trailing, repeated,
and empty fills around every arrow form and both row types.

All 2,112 SVG references match exactly. Compared with baseline main
`6a7adcd1c6cce9df99b3363d40f22c2f68857fd5`, 2,036 fail and 76 already match.
The corpus covers all three spellings; leading, trailing, interior and
repeated fills; multiple cells and rows; explicit and generated matrices;
AMS/Mathtools arrays; fonts; empty entries; nested scopes; command arguments;
script recipients; pending items; dynamic registrations in both orders;
and standalone error behavior. Every one of the 914 newly captured CD
comparisons matches. All 36 historical CD HFill observations from
`cd_entries_residuals.json` are covered as passing references here, with 30
additional unique inputs beyond the new sweep. Their historical receipts
remain unchanged.

`hfill_residuals.json` preserves 106 other valid-original observations from
the complete 2,218-input inventory, including raw original, baseline, and
candidate outputs. None was previously exact. Matching no-HFill controls
identify the independent gaps: 36 no-node final environment rows, 20
alignedat/flalign layout cases, 20 unsupported eqnarray environment cases,
10 displaylines ampersand scanning cases, eight raw nested-environment cell
splits, and 12 dollar/Braket/script ownership cases. These are not accepted
as passing references or replaced with Go-generated goldens.

A visible witness gives all three columns the same width:
`\begin{array}{c|c|c}\hfill x&x\hfill&\hfill x\hfill\\abcdefgh&abcdefgh&abcdefgh\end{array}`.
Its top row is right-, left-, and center-aligned, respectively.

Regenerate both inventories with
`node --jitless testdata/generate_hfill.cjs PINNED_ASSETS`.
The checked-in input inventory is fixed; the generator does not consult or
filter on candidate results. Workers use batches of 24 fresh VMs. Run
`go test ./... -run TestHFillReferences`.
