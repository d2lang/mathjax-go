# Mandatory argument whitespace

The pinned MathJax 3.2.2
[GetArgument method](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/TexParser.ts#L299)
starts with GetNext, whose whitespace test is JavaScript `/\s/`.
That includes U+FEFF (BOM) and excludes U+0085 (NEL). The previous Go
`readArgument` used `unicode.IsSpace`, which makes the opposite choices for
those two characters. For example, BOM between `\frac` and `{a}{b}` became
the numerator instead of being skipped.

The change uses the existing JavaScript whitespace predicate only at the
start of `readArgument`. Generic math tokenization, `skipSpaces`,
`readArgumentAtCursor`, and caller-specific lookahead remain unchanged.
Source-specific GetBrackets or GetStar lookahead can independently consume
whitespace before GetArgument; their behavior is not claimed fixed here.

## Original observations and exact assertions

The public inventory has 1,024 distinct TeX/display pairs:

- 706 complete, unmodified original SVG assertions, including 30 original
  error renderings. Against main159 `c19a00c2202542652ffa46ce0a93d17c0bdd8108`,
  156 become byte-exact and no previously exact case becomes nonexact.
- 316 unresolved observations: 90 original-valid SVGs, 40 original error
  SVGs, and 186 original runtime exceptions. All original, baseline and
  candidate outputs are retained, including unchanged controls.
- Two additional cancel serialization controls are kept outside the exact
  assertions and the fix count. Both binaries produced both attribute orders
  in 64 repeats each; all 256 parsed XML trees equal the originals, including
  tags, attributes and exact values, child order, text, and tails.

A recursive scan of all 396 tracked JSON files under root and internal
`testdata` directories on main159 finds 50 asserted input pairs already
published: 48 previous raw observations promoted to complete SVG assertions
and two existing full-SVG controls. The other 656 asserted pairs are first
published here. These are per-fix input counts, not globally unique MathJax
coverage totals.

The independent method fixture under `internal/tex/testdata` retains all
308 original GetArgument observations. Of these, 302 assert the returned
value or exact error, consumed UTF-8 byte cursor, and remaining source. The
original UTF-16 cursor is also preserved. Six observations of an escaped LF,
CR, or U+2028 retain a pre-existing GetCS value mismatch: the original returns
control-space, while Go returns the escaped line character. Their source path
and baseline value/cursor are bound separately; these are not normalized into
passing assertions and their input begins with a backslash, so the changed
leading-whitespace loop is never entered.

Public controls include both fraction arguments, font and text readers,
phantoms, decorations, macros, environment names and column specifications,
paired commands, the newly registered matrixdeterminant/mdet/smdet aliases
and their vmqty/svmqty helpers, EOF and closing-brace boundaries, Unicode space and nonspace
characters, and plain-token/group/dimension controls. `\enclose` is unavailable
in the frozen D2 component; its 24 inputs remain explicitly labeled unknown-
command controls, not evidence that an Enclose handler runs.

## Changed residuals

Ten changed original-valid residuals have strict complete literal bindings in
`getargument_whitespace_qualification.json`: each original target equals its
original literal control, and each candidate target equals the unchanged
main159 rendering of that control. The residuals remain unasserted:

- Two `\dfrac` BOM inputs inherit the ordinary dfrac wrapper difference.
  Their complete outlined geometry already equals the original.
- Four NEL text arguments followed by BOM now consume the correct argument
  and restore the original root dimensions. The remaining ordinary-math BOM
  font variant differs from the original, exactly as in the explicit braced
  NEL control. Complete outlined equality is not claimed for these inputs.
- Four `\qty`/`\order` BOM inputs inherit the legacy literal handler's fence
  choice or token-font/function-application differences. The source-correct
  argument boundary exposes the same unchanged literal result.

The other changed unresolved observations are 150 original runtime exceptions
and four original error renderings (two unknown-command controls and two
unknown-environment diagnostics). Runtime exceptions are not valid-rendering
parity claims. No original output is replaced with candidate output.

The separate current published-residual replay covers 5,364 inputs: 48 genuine
fixes, zero exact regressions, and 44 changed unresolved observations (40
original runtime exceptions plus the four literal-bound qty/order cases).
Three apparent cancel matches are excluded through repeated serialization
proofs. The 4,326-input upstream MathJax-Tests replay has no genuine change;
its sole changed enclose result is attribute ordering in an already divergent
Go handler (the frozen original returns an unknown-command error).

`getargument_whitespace_replay_qualification.json` preserves 640 repeated
renders across five controls. All candidate and current-baseline XML trees
agree, retaining tags, attribute values, ordered children, text and tails.
The cancel controls also equal their complete original trees; the unavailable
enclose control remains explicitly divergent from the original. These proofs
qualify counts only; no exact SVG assertion normalizes attribute ordering.

The main159 composition retains all 840 prior inputs and adds 184 distinct
input pairs from unmodified determinant residuals and fresh caller controls.
All 20 previously published BOM determinant/helper observations become exact.
Direct, script, MathFont and Position controls also cover BOM, NEL, line
separator and ASCII whitespace with the newly registered aliases. Original
NEL runtime failures remain raw; none is presented as a rendering fix.

## Regeneration

```sh
node --jitless testdata/generate_getargument_whitespace.cjs PINNED_ASSETS
node --jitless internal/tex/testdata/generate_getargument_whitespace.cjs PINNED_ASSETS
```

Both generators verify all three frozen asset hashes and invoke original
MathJax only. Public rendering uses a fresh VM per input in bounded batches.
It regenerates original SVG fields and does not select an assertion subset or
invoke Go. Runtime-only records retain their complete captured stack strings;
a fresh runtime must still raise the same first-line error message. The
original caller stack is not rewritten merely because a checkout path differs.
The SVG test is `TestGetArgumentWhitespaceReferences`; the method test is
`TestGetArgumentWhitespacePrimaryMethod`.
