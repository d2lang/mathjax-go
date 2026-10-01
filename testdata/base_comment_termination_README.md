# Original Base comment termination

The original `BaseMethods.Comment` consumes characters until LF, leaving that
LF for the normal Space handler. A bare CR inside a comment does not terminate
it. Go instead stopped at CR as well, so hidden formula text could reappear in
the output or an unknown control sequence inside the comment could become an
error.

The Go scanner now follows the original LF-only boundary. The D2 witness
contains an actual CR before the fraction and an actual LF before `=C`. Its
escaped JSON spelling is:

```json
{"tex":"A+B% hidden\r+\\frac{1}{0}\n=C"}
```

The original and fixed output is `A+B=C`; Go previously displayed the additional
`+1/0` fraction. D2 preserves the bare CR inside its LaTeX block, so the same
input demonstrates the fix in a D2 render.

## Frozen original verification

`base_comment_termination_mathjax_3_2_2.json.gz` retains 302 complete original API
objects from fresh frozen MathJax 3.2.2 runtimes: 282 valid SVGs and 20
rendered-error SVGs, with no runtime exceptions or excluded cases. Every SVG
matches byte for byte after the fix. Against the preceding Go source `b1a2d61`,
62 valid cases differ and 240 remain exact controls.

The matrix covers LF, CR, CRLF, Unicode line separator and paragraph separator,
hidden operators/fractions/unknown commands/groups/scripts/sums, groups, nested
script styles, box child parsers and matrices. LF and CRLF remain independent
controls; Unicode line and paragraph separators correctly remain comment text.

Source commit: `ad8f5c21cb810236551da8c6512ba733e67357ee`. The original-only
regenerator verifies the pinned D2 asset hashes and fails on unexpected runtime
exceptions rather than discarding inputs:

```sh
python3 testdata/generate_base_comment_termination.py PINNED_ASSETS NODE
```
