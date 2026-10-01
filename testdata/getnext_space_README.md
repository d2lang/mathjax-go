# Active GetNext whitespace boundaries

The frozen original `TexParser.GetNext` advances while `nextIsSpace` matches
JavaScript `\s`. That set includes U+FEFF (BOM) and excludes U+0085 (NEL),
U+180E (MVS), and U+200B (zero-width space). Six active command families still
used Go's whitespace set at these boundaries: Base Matrix, Braket, Physics
Derivative, Differential, Expression, and DiagonalMatrix. BOM could hide a
matrix body, a derivative argument, or an automatically sized parenthesis.

The narrow shared helper now uses the same original-shaped predicate as the
already corrected argument/bracket readers. The six call sites change; the
generic Go-only reader and GetDelimiter remain outside this change. Matrix's
previous special handling for displaylines is retained by the shared predicate.

`getnext_space_mathjax_3_2_2.json.gz` contains 2,256 complete valid original SVGs
and 104 original runtime exceptions stored separately. Every valid SVG matches
byte for byte. The prior combined renderer matches 2,158 valid controls and
differs on 98 valid BOM inputs. All 25 JavaScript whitespace characters and the
three exclusion controls are covered across 36 registered aliases, both display
modes, optional powers, multi-argument derivatives, scripts, and font/color
contexts. The D2 formulas use literal BOM and are included in these references.

The original-only generator also records active handler identity, registration
arguments, the selected unmodified functions, and 28 original GetNext cursor
observations. This confirms actual D2 package precedence before crediting any
command. The direct cursor test compares UTF-16 source positions with Go's
UTF-8 byte positions.

The 104 NEL inputs throw the original `TypeError` in `SymbolMap.item`. They are
retained to identify that original runtime boundary and receive no SVG parity
credit. An independent expanded review also retained diagnostic-only boundaries
in malformed matrix-generator inputs; those are separate from valid-original
whitespace repairs. Legacy Matrix spacing after a preceding math text token is
being verified in a separate focused change. These finite references do not
establish universal MathJax parity.

Source commit: `ad8f5c21cb810236551da8c6512ba733e67357ee`. Both original-only
generators verify the frozen D2 asset hashes and never read Go output:

```sh
python3 testdata/generate_getnext_space.py PINNED_ASSETS NODE
NODE --jitless testdata/generate_getnext_bindings.cjs PINNED_ASSETS
```
