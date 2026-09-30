# Unicode reversed-C evidence

In `{\Huge aↃb}`, the central reversed C (U+2183) is incorrectly slanted before the fix. The candidate makes it upright, matching original MathJax 3.2.2; the surrounding `a` and `b` remain italic. `after.svg` and `original.svg` are byte-identical. The three-panel screenshot shows the untouched SVGs at one shared scale, with identical 121×93 intrinsic dimensions. Root and an independent reviewer viewed the comparison and found no clipping.

Before is actual merged mathjax-go `6d88373dd7479949265afdeae982b1bd5fcd5613`; its 271 production/dependency blobs match the before binary's source `b25453663d1e4b6f12be0943d367bbb2aab03ae8`. After is candidate source `0158db95a3c1d3bec582320314ad1679fa469a61`. Both use D2 renderer `01bc7ecdbdd04c13d6fe5df1967d2d9aa14ae579`. Original comes from D2 using the unmodified, hash-verified MathJax 3.2.2 bundle. The original and before captures were retained unchanged; the candidate run rendered only after outputs.

The change uses the original ASCII-letter dispatch boundary. Non-ASCII letters proceed through MathJax's `Other` character handling, which selects this glyph's original token and style. The complete focused math SVG for this exact input matches the original in both display modes and is embedded in the complete D2 SVG.

All 98 previously published D2 witnesses also remain byte-identical to their original references with this candidate, for 99 checked D2 witnesses including this one. This visual demonstrates the reversed-C fix; it does not establish complete Unicode or MathJax coverage. Asset and artifact hashes are in `manifest.json`; reproduction instructions are in `REPRODUCE.md`.
