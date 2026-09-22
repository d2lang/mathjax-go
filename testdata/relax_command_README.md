# Relax command dispatch references

`generate_relax_command.cjs` runs the untouched pinned D2 MathJax 3.2.2 assets (hashes checked in the script), with a fresh VM for each of 70 public source/mode cases. It observes the complete explicit/own-property tree before the unchanged SVG renderer and stores the complete SVG hash and integer display dimensions. Error references retain the primary error text, tree, and SVG.

The D062 correction removes only the built-in `relax` no-op. The existing undefined-command fallback now handles math occurrences, including scripts, primes, fonts, groups and the `stackrel` substitution macro. Registered macro definitions retain earlier precedence. Literal `text`/`mmlToken` and comments are not scanned for forbidden strings. The internal four-case test exercises defined `relax` and other macros expanding to an undefined `relax` through the actual row parser.

64 cases require raw primary SVG, complete explicit tree and own-property tree equality; 40 primary cases are errors. Six named controls retain exact prior-Go output from D061 head `03c34cde20d3c98f326a7cb3961290e0264375ab`: inline/display `pmod` encounters an earlier undefined `pod`, `mbox` lacks its original `mstyle` wrapper, and dollar-delimited math inside `text` remains literal. Their complete primary references are untouched. `relax_command_boundaries.json` separately binds the prior SVG and tree for each exact source/mode. These six outputs are unchanged by D062; they are not claimed as primary parity or fixed here. Separate observation evidence records the remaining scope.

Thirty-four whole SVG/explicit/own-tree cases fail on the prior dispatcher and match the original after the correction. The other 36 full SVG and raw compiler-tree outputs remain identical, including the six qualified controls. Existing no-op/tag/style/limits and inherited prime/annotation fixtures remain untouched. Other generalized-fraction, binomial, argument and whitespace discrepancies are outside this change.

Regenerate with:

```
node testdata/generate_relax_command.cjs /path/to/pinned/d2latex /path/to/evidence
```

The optional evidence output includes every original SVG, the explicit/own-property observation and the full raw MML tree. Public tests compare raw JSON values after serializing the Go tree, preserving types and all recorded fields without a normalization/projection exception.
