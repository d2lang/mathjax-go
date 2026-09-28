# FrameBox width and alignment

The pinned MathJax 3.2.2 source commit is
`ad8f5c21cb810236551da8c6512ba733e67357ee`.
[BaseMappings](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMappings.ts)
registers `framebox` as
[BaseMethods.FrameBox](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts#L1032-L1048).
The name was listed in Go's source inventory but had no executable handler.

A visible example is:

```tex
\Huge\framebox[6em][l]{left}\quad\framebox[6em][c]{$x^2+y^2$}\quad\framebox[6em][r]{right}
```

The baseline reports an undefined command. The candidate and original produce
identical complete SVGs containing three boxes with left, center and right
aligned contents. Widthless `\framebox{hello}` uses the content's natural width.

## Source boundaries

FrameBox reads two optional bracket arguments and one required argument.
[TexParser.GetBrackets and GetArgument](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/TexParser.ts)
supply the lexical boundaries and error order. The local GetArgument boundary
uses JavaScript whitespace, including BOM and excluding NEL, without changing
ordinary math tokenization or other consumers.

[ParseUtil.internalMath](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ParseUtil.ts)
handles text and embedded `$...$` or `\(...\)` math. The existing implementation
is reused: literal text follows the active font, while embedded math uses a
fresh empty lexical environment and independent macro counter, sharing the
configuration. Public declared-operator controls cover the 1,000-expansion
limit, a child overflow, independent successive embedded children and a caller
already at the limit.

A nonempty width wraps the content in `mpadded` with the authored width string.
Alignment is selected by
[Options.lookup](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/util/Options.ts#L334-L343):
only the exact strings `l` and `r` mean left and right; every other value means
center. The options are not expanded or trimmed, and inherited object-property
names do not match. An empty or missing width creates no padding wrapper.
FrameBox then creates a boxed `menclose` inside an ORD TeXAtom. Existing padding,
alignment and enclosure rendering need no changes.

## Original references and unresolved cases

The inventory has **1,344 unique TeX/display inputs**. Against merged main139,
`b24e36d0f65fcbe61488fe2a17392e7639a98429`, **1,286** match the entire original
SVG: **1,214 valid expressions and 72 original error renderings**. Of those,
**1,138 become exact and 148 were already exact**. There are no formerly exact
regressions. The earlier main137 and main138 baselines produce identical outputs on the
entire inventory. Independent source/public review is included in these counts. All 2,325 unique
inputs from existing published residual inventories also remain byte-unchanged.

The references cover natural, fixed, relative, percentage, negative and invalid
widths; literal alignments and prototype-property names; text/internal math and
unbraced ownership; scripts, roots, fractions, positions, fonts, colors and
sizes; matrices and CD; declaration priority and shared configuration; malformed
options and contents; Unicode and JavaScript whitespace. Existing FBox, boxed,
text and HBox controls verify unchanged behavior.

`framebox_residuals.json` preserves **58 raw original/baseline/candidate
observations**, rather than treating them as parity assertions. **28 are
unchanged and 30 change without becoming exact**. Four originals throw at
runtime on NEL; these are not valid-rendering claims. Four error renderings
have only the existing safe XML escaping of an ampersand in an error attribute.

The remaining differences are existing box font/internal-math behavior and
newly reachable EOF or enclosure metric issues. The shared GetArgument helper
omits the original's synthetic control-space at a terminal backslash. Explicit
braced controls preserve both outcomes: the original EOF output equals
`\framebox[3em][r]{\ }`, while the candidate EOF output equals
`\framebox[3em][r]{\\}`. Existing FBox has the same EOF issue.

For literal `~`, `^` and `_` text, existing enclosure metrics differ from the
original. [CommonMenclose.computeBBox](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrappers/menclose.ts#L291-L301)
combines the child into the initial zero box. Go empties that box first, so
negative child heights or depths produce a smaller rectangle. The glyph data
itself is unchanged. `framebox_geometry_proof.json` records seven suffixes in both modes:
original FrameBox and FBox have matching view boxes, rectangles and glyphs;
candidate FrameBox matches the unchanged baseline FBox geometry. All underlying
raw references are retained. These shared parser/enclosure-metric issues are kept
separate from the missing FrameBox handler.

An affine SVG comparison, composing transforms and comparing geometry/paint,
found **1,138 visual fixes and zero visual regressions among 148 previously
visually exact inputs**. There are 24 changed visible residuals, all in the
EOF/enclosure-metric families above; no structural-only change is claimed as a
visible fix.

## Regeneration

```sh
node --jitless testdata/generate_framebox.cjs PINNED_ASSETS
go test ./... -run 'TestFrameBox'
```

The generator verifies all three frozen asset hashes and starts a fresh original
VM for each conversion, in bounded batches of 24. Complete original SVGs use
font cache none, em=16, ex=8 and the stored display mode. Unicode separators are
escaped only on the JSONL wire. Raw original runtime failures are regenerated;
historical Go observations and the derived geometry report remain unchanged.
