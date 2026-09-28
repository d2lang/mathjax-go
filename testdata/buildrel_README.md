# BuildRel annotated relations

The pinned MathJax 3.2.2 source commit is
`ad8f5c21cb810236551da8c6512ba733e67357ee`.
[BaseMappings](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMappings.ts)
registers `buildrel` as
[BaseMethods.BuildRel](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts).
It was present in Go's source inventory but had no executable handler.

A visible example is:

```tex
\Huge\boxed{A\buildrel\text{definition}\over=B\quad C\buildrel n\to\infty\over\longrightarrow D}
```

The baseline reports an undefined command. The candidate and original produce
the same complete SVG, with the annotation over a relation and relation spacing
between the surrounding operands.

## Source boundaries

[TexParser.GetUpTo](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/TexParser.ts)
looks for the literal `\over` control sequence outside braces. It does not
expand macros or skip comments while searching. Control-word boundaries,
escaped braces, the slice endpoint before whitespace, and unmatched-close
errors follow that lexical helper. This does not alter ordinary infix parsing
or the independent `\root ... \of` handler.

The annotation and the following single argument each use a fresh child parser.
They share configuration, copy the lexical font/color/root environment and
start independent macro counters. BuildRel reuses the existing `parseChild`
implementation. Its base argument uses GetNext's JavaScript whitespace before
GetArgument captures it; that local boundary includes BOM and excludes NEL,
without changing other argument consumers.

BuildRel produces a REL TeXAtom containing an under/over node with no under
child. [FilterUtil.cleanSubSup](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/FilterUtil.ts)
reduces it to `mover`. The [TeX postfilter order](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex.ts#L133-L138)
runs cleanSubSup at priority -6, before inheritance (-5), moveLimits (-4),
cleanStretchy (-3), attribute cleanup (-2) and relation combination (-1).
BuildRel sets no attributes on that temporary munderover, so no explicit
attributes or inherited state are discarded by constructing the final mover.
The temporary munderover registration is removed by cleanSubSup and the
replacement mover participates in subsequent filters; Go's existing tree-based
passes visit the final mover and preserve the child operator registrations.
The implementation therefore constructs that final shape directly, without
Overset's accent or movable-limit adjustments.

## References and unresolved cases

The full inventory has **952** unique TeX/display pairs. Against merged main137,
`6f14f09f1e15ffb5a4e94cca0eb6bfa4a790a139`, **898** match the complete original
SVG: **542 valid expressions and 356 original error renderings**. Of these,
**222 were already exact and 676 become exact**, with no formerly exact
regressions. Independent source and public reviews are included in that corpus.
All 2,088 existing published residual inputs also remain byte-unchanged.

The references cover grouped and unbraced arguments, literal terminator and
comment boundaries, macro expansion and declaration priority, invalid argument
precedence, labels, scripts, fonts, colors, sizes, pending parser items, roots,
matrices and CD compositions. Declared operators exercise the original Macro handler at the 1,000-call limit:
six cases are valid and six assert overflow in either child or the restored
caller continuation. These public cases and a
separate caller-state test verify both independent child counters and restoration
on success or error. The frozen original does not provide `\let` or `\newcommand`; their unchanged
undefined-command controls are retained as diagnostics, not counter evidence.

`buildrel_residuals.json` keeps the other **54 raw original/baseline/candidate
observations**. **24 are unchanged and 30 change without reaching parity**.
Eighteen originals fail at runtime on unsupported NEL tokens; these are not
valid-render parity assertions. Existing generic `\frac`/`\overset` BOM
boundaries remain unresolved outside the new handler.

For all **16 newly reachable valid BuildRel residuals**, fresh literal controls
using `\mathrel{\overset{...}{...}}` prove both that the original literal equals
the original BuildRel output and that the candidate equals the baseline literal
output, byte for byte. These cover nested arrow annotation layout, CD spacing,
and escaped LF/CR/U+2028 child tokenization. Their raw original references are
preserved, not normalized or replaced with Go output.

An affine SVG comparison found **zero visual regressions among 222 previously
visually exact inputs**, with **676 visual fixes**. The 16 changed visible
residuals are the literal-proven inherited families above. The audit compares
root dimensions and visible geometry/paint after composing transforms; it does
not classify structural-only changes as visible fixes.

## Regeneration

```sh
node --jitless testdata/generate_buildrel.cjs PINNED_ASSETS
go test ./... -run 'TestBuildRel'
```

The generator verifies the three frozen asset hashes, starts a fresh original
VM for each conversion, and bounds batches to 24 VMs. Original SVGs use font
cache none, em=16, ex=8 and the recorded display mode. Unicode line separators
are escaped only on the JSONL wire; the stored TeX and original SVG remain intact.
Raw original runtime failures and literal controls are regenerated as well;
historical Go observations are left untouched.
