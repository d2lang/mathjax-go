# Cancel option filtering and paint values

The frozen original Cancel methods call `ParseUtil.keyvalOptions` with the
enclosure allowlist and the default `error=false`. Unknown own keys are deleted
after parsing the complete list. Go used a duplicate option parser that rejected
unknown keys, trimmed Go whitespace, and kept the literal boolean values as
strings. Valid cancellation formulas could therefore become error messages or
lose their requested color.

Cancel and Enclose now share the original-shaped option reader, with the
filtering/error flag explicit. Mathtools and Empheq keep `error=true`. The shared
reader preserves balanced braces, escaped separators, JavaScript trimming,
typed boolean values, ordinary-object `__proto__` behavior, and Object.keys
enumeration before filtering.

SVG `handleColor` also tests the original value before string coercion, as the
source's `mathcolor || color` and background fallback expressions do. A typed
`false` does not become the invalid paint string `"false"`. The attribute remains
in the MathML tree; this change applies only when selecting explicit SVG paint.

These D2 inputs are included in both display modes:

```tex
\cancel[unknown=x]{x}+\bcancel[__proto__=true]{y}+\cancelto[unknown=x,mathcolor=red]{0}{z}
\cancel[mathcolor=﻿red﻿]{x}+\cancel[mathcolor=false]{y}
```

The second line contains actual U+FEFF characters around `red`, corresponding to
`mathcolor=\ufeffred\ufeff`.

## Frozen original verification

`cancel_options_mathjax_3_2_2.json.gz` retains all 1,756 complete original API
objects from fresh runtimes: 1,724 valid SVGs and 32 runtime exceptions stored
separately. Every valid SVG matches byte for byte after the fix. Against the
preceding Go source `43f197a`, 1,092 valid cases differed and 632 were exact
controls. Of the differences, 722 were valid original formulas that Go rendered
as errors, and 370 were other valid SVG differences. There are no original
rendered-error SVGs in this matrix.

The matrix covers all four Cancel commands, allowed and unknown options,
numeric/inherited/prototype keys, duplicates, braces/commas/escapes, BOM and NEL,
bare keys and literal booleans, supported enclosure attributes, color/background
fallbacks, nested script/font/color contexts, child text parsers, and matrices.
Independent `mmlToken` color controls retain quoted `false`, `0`, and `red` values.

The 32 original runtime failures are `TypeError: t.trim is not a function` in
`getParameters` for boolean-valued `data-arrowhead` inputs. Tests bind that
specific original failure path and separately verify bounded Go output. These
inputs are retained as original runtime limitations and receive no SVG parity
credit.

`internal/tex/testdata/filtered_keyval_mathjax_3_2_2.json` adds 90 direct original
helper observations: 66 successful option objects and 24 structured errors.
They cover no allowlist, the filtering Cancel allowlist, and the same strict
allowlist, including numeric key order and malformed unknown values. The
existing 194 public and 90 direct keyval own-property references remain strict.

Source commit: `ad8f5c21cb810236551da8c6512ba733e67357ee`. Both original-only
regenerators verify the pinned D2 asset hashes; all SVG, error, and runtime
objects are preserved without constructing expectations from Go output:

```sh
python3 testdata/generate_cancel_options.py PINNED_ASSETS NODE
NODE --jitless testdata/generate_filtered_keyval.cjs PINNED_ASSETS
```
