# Row command-map priority

`row_command_priority_mathjax_3_2_2.json` retains all 2,556 expressions from
an audit of commands handled directly by Go's row parser, plus 48 exact
line-break boundary controls and 476 exact independent review controls. Every expected SVG
is the complete, unmodified output of D2's frozen MathJax 3.2.2 bundle at
source commit `ad8f5c21cb810236551da8c6512ba733e67357ee`. The fixture records
all three original asset hashes. Each expression uses a fresh runtime, no
font cache, `em=16`, `ex=8`, and its recorded display mode.

In the original parser,
[`ParseMethods.controlSequence`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ParseMethods.ts)
passes control sequences through the macro handler, and
[`SubHandler.parse`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/MapHandler.ts)
selects the first applicable command map. Dynamic registrations therefore
precede the builtin handlers that Go implements as row-parser shortcuts.

[`AmsConfiguration`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ams/AmsConfiguration.ts)
registers declared operators at priority -1.
[`MathtoolsConfiguration`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/mathtools/MathtoolsConfiguration.ts)
registers paired delimiters at priority -5, and
[`MathtoolsUtil.addPairedDelims`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/mathtools/MathtoolsUtil.ts)
adds entries to that separate map. A paired delimiter wins over a declared
operator in either declaration order.

The row parser now checks both registration maps once, before its builtin
shortcuts for `right`/`middle`, infix fractions, `color`, math styles, math
sizes, fonts, `limits`/`nolimits`, and a line break with pending positioning.
A registered command reaches the existing dispatcher. This shares the guard
already required for font declarations; it does not change ordinary command
or package dispatch order.

The references exercise declared operators, paired delimiters, both
registration orders, grouped declarations, ordinary groups, scripts, pending
functions and negation, positioning, outer sizes, and open `left`/`right`
frames. Paired delimiters also use starred and explicit-size forms. Adjacent
font, macro, strut, and package commands are retained as controls. All cases
run in both display and inline modes.

Against baseline `f23c7d5e1ca718de61102161b01178b6ba7f168c`, 1,382 references
already matched and 1,698 differed. All 3,080 now match exactly, with no
formerly exact output changed. The original 2,556-expression inventory
contains 2,552 valid renderings and four original error SVGs, all asserted.
Those four error cases involve a starred paired delimiter named `right` and a
self-referential declared `operatorname`.

A fresh comparison against merged CD baseline
`c0a304b7ff0fc538d9ca824ca74d421961322d96` produced identical baseline SVGs
for all 3,100 observed expressions. The final baseline counts are therefore
also 1,382 exact and 1,698 newly fixed against that commit, with zero
regressions; all 20 residuals remain unchanged. The historical baseline
metadata in the fixtures is retained.

The additional boundary audit covers a registered control-symbol `\\`,
`cr`, `newline`, and a literal escaped newline, alone and inside pending
position, size, and function contexts. It contributes 48 exact references,
including 18 newly fixed outputs. Its remaining 16 observations are retained
in `row_command_priority_residuals.json` with raw original, baseline, and
candidate SVGs. All 16 candidate outputs are byte-identical to baseline:
eight valid original renderings and eight declaration-error controls. They
are not assertions of Go correctness.

These residuals expose separate lexical and declaration-validation gaps.
Original [`TexParser.GetCS`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/TexParser.ts)
uses a regular expression whose dot does not match a newline; the fallback
returns a space control sequence. Go returns the raw newline instead.
Original paired-delimiter declarations also reject a control-sequence name
containing a newline. Fixing those token and name rules is separate from
respecting registrations already present in the maps. No original result
has been normalized or rewritten to conceal these residuals.

Independent review contributed another 480 fresh original comparisons:
476 exact results, 284 newly fixed results, and no regressions. These cover
pending Not/Dots/Prime items, positioning and size, empty font scopes, roots,
matrices, and Left/Right barriers. All 476 exact outputs are asserted in the
main fixture. The remaining four valid originals redefine `cr` inside a
plain `matrix`; Go's existing eager matrix row splitter consumes that command
before row parsing. Their raw SVGs are retained in the residual file, and
all four candidate outputs are byte-identical to baseline. This is a separate
array-entry dispatch boundary.

The complete asserted fixture contains 3,072 valid original renderings and
eight original error SVGs. The residual file contains 12 valid renderings
and eight original errors; all 20 candidate receipts remain unchanged from
baseline. The fixture generator preserves those baseline and candidate
receipts when regenerating the raw original SVGs.

A compact visible witness is
`\DeclarePairedDelimiter{\over}{[}{]}a\over{\Gamma}+b`.
The original renders `a[Γ]+b`. Before the guard, Go instead enters its builtin
infix-fraction handler and places `a` over `Γ+b`.

Regenerate with
`node --jitless testdata/generate_row_command_priority.cjs PINNED_ASSETS`.
The generator verifies the pinned hashes and uses bounded batches of 24
fresh original VMs. It updates only original SVGs, not the audit inputs or
baseline counts. Run `go test ./... -run TestRowCommandPriorityReferences`.

The full frozen-oracle test suite, race tests, vet, and WebAssembly build pass
on the candidate rebased onto the merged CD baseline. Regenerating both
fixture files from the pinned original assets reproduces them byte for byte.
