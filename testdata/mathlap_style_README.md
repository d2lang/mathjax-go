# MathLap optional style and argument references

Pinned original source: MathJax 3.2.2 commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`.
[MathtoolsMethods.MathLap](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/mathtools/MathtoolsMethods.ts#L212)
reads an optional literal style, parses one argument in a child TexParser,
creates its zero-width padded box inside mstyle, and applies
[MathtoolsUtil.setDisplayLevel](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/mathtools/MathtoolsUtil.ts#L47).
The six registrations are `mathllap`, `mathrlap`, `mathclap`, and their
`cramped` variants.

Go previously treated the opening `[` as the command's argument. The rest
of the option leaked into the caller, and an explicit style command there
changed the surrounding expression. The math variants now consume the
option first and apply only the four literal TeX style names accepted by
the original. Unknown options are consumed without being executed. The
existing Mathtools style helper provides the source JavaScript trim behavior.
The unprefixed `llap`, `rlap`, and `clap` paths remain separate.

The local argument boundary also follows TexParser.GetArgument's JavaScript
whitespace and TexParser.ParseArg's independent child parser. This retains
the copied font/root environment and shared command registrations while
giving the child its own macro counter. Missing brackets retain the original
active-command diagnostic. Existing argument capture, layout, and public
command-override paths are reused.

The current 1,788 unique captured observations contain 1,726 complete exact
original SVGs (1,446 valid and 280 error renderings), with 1,434 fixes against
main149 `e3d4006b093073ad46555f92e94df4c85dd46331`. They cover all six
commands, default/empty/recognized/unknown
options, JavaScript whitespace, malformed brackets, terminal backslash,
fonts, nested wrappers, roots, scripts, Nonscript, pending items, dynamic
registrations, arrays/CD, and caller/child macro budgets at and beyond the
1,000-expansion boundary. All 112 intersections with the first-final-script-item
repair are exact,
and the prior 1,676 candidate outputs are unchanged after that composition.

Four already-published observations from `cramped_style_residuals.json` are
promoted by a separate test. The historical receipt is unchanged, and those
four references are not counted as new observations.

All 62 nonexact observations are retained in `mathlap_style_residuals.json`:

- Twenty-four NEL inputs cause original MathJax to throw a JavaScript runtime
  exception in the captured child argument. The Go output changes because
  it now captures NEL according to the source boundary, but remains a
  nonpanic rendering. Original runtime stacks and both Go outputs are kept;
  these are not counted as source-exact results.
- Two unprefixed `clap` text/math controls retain the existing internal-text
  parsing gap. Their before and after SVGs are byte-identical.

Twelve optional Quantity/infix-over expressions retain an inherited
wrong-acceptance gap: original MathJax returns an error while Go renders.
Twelve corresponding optionless controls prove the original error is
identical and the control Go output is unchanged. The full proof is in
`mathlap_style_quantity_over_proof.json`; all 24 observations remain raw.

An independent review supplied 180 distinct original inputs: 168 exact,
108 fixes, and these 12 inherited Quantity/infix-over cases. Its source
review found no scoped issue. All of its references and 12 optionless
controls are included in the deduplicated counts above.

Twelve child-array error controls retain the original `Misplaced &`
diagnostic. Original MathJax leaves a bare ampersand in the XML attribute;
Go escapes it. The entire remainder is identical, and the Go output is
unchanged. `mathlap_style_ampersand_proof.json` preserves the raw strings
and exact substitution proof; these remain outside the exact count.

No valid changed nonexact rendering was found in this fresh inventory.
Replaying 4,306 published main149 residual inputs fixes the four promoted
lap observations. Three unrelated cancel serializations change attribute
order only (one became byte-exact and two ceased to be byte-exact in this run);
the complete parsed XML trees are identical before, after, and
original. The raw strings and proof remain separate from exact assertions.

The per-fix count is deduplicated by TeX and display mode. A scan of main149
JSON references with explicit `tex`, `display`, and complete `svg` or
`original.svg` fields found no complete-SVG overlap among these 1,726
assertions. The report records this scan scope rather than claiming a
global project coverage total. Four promotions remain separate.

Regenerate the captured original inventories with:

```sh
node --jitless testdata/generate_mathlap_style.cjs /path/to/pinned-assets
```

The generator checks all three frozen D2 asset hashes, creates a fresh
original VM for each expression, and never invokes Go. It preserves full
original SVGs and runtime failures. Both inventories regenerate
byte-identically. Final base and gate results will be recorded before handoff.
