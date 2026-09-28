# Plain TeX cases command

Source: `BaseMappings.cases`, `BaseMethods.Matrix`, and `BaseMethods.Entry` in MathJax 3.2.2. `\cases` uses a left brace, left-aligned columns, and `.1em` row spacing. Its second column is scanned as text before normal TeX parsing: literal percent signs are retained, embedded `$...$` and `\(...\)` math is supported, extra alignment tabs are rejected, and a complete `\text{...}` cell retains the original compatibility behavior.

The fixture stores 246 unmodified full SVGs from D2's frozen original MathJax at `ad8f5c21cb810236551da8c6512ba733e67357ee`. The generator validates the three asset hashes and uses Node `--jitless`. Rebuild with `node --jitless testdata/generate_matrix_cases.cjs /path/to/pinned/assets`.

Coverage includes implicit text and embedded math, exact `\text` recognition, groups and escapes, literal percent signs versus first-cell comments, Unicode, references, nested matrices/cases/environments, macros, font/color/root scopes, display/inline/script sizing, row gaps, extra tabs, incomplete math, missing arguments/braces, and first-cell error precedence. Final rows containing only comments or font declarations are omitted, as `ArrayItem.EndTable` requires; ordinary matrix and numbered-alignment controls cover that shared boundary.

238 SVGs are byte-exact. The other eight contain the original diagnostic `Misplaced &`; the original serializer leaves a bare ampersand in its `data-mjx-error` attribute. The test requires Go to retain safe XML escaping for that attribute alone, with identical error text, rendered text, geometry, and remaining SVG. The stored oracle is unchanged. The preceding main commit `729d74d` differed on 228 cases.

D2 witness math: `f(x) = \cases{x^2 & if $x\ge0$ \\ -x & otherwise}`.

Separate existing array-parser followups: macro expansions that introduce row/cell separators are not recognized by the shared raw-source table splitter, and literal math-mode `$` outside the cases text cell lacks the original TeXAtom wrapper. Neither boundary changes this command's second-column text semantics.

Review followup: leading/trailing U+2003 or U+FEFF around a complete `\text{...}` cell takes the ordinary math parser path, where existing Unicode whitespace/token behavior differs from the original. This is distinct from implicit cases text trimming.
