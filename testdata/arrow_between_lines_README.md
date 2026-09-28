# Mathtools alignment command ordering

These references use the original MathJax 3.2.2 runtime frozen by D2. The exact commit and SHA-256 hashes of all three runtime assets are recorded in both JSON files. The generator verifies those hashes and obtains each output from a fresh original runtime; it never uses Go output as an expected SVG.

The source implementation follows `MathtoolsMethods.ArrowBetweenLines`, `Aboxed`, `VDotsWithin`, `ShortVDotsWithin`, `FlushSpaceAbove`, and `FlushSpaceBelow`, with `MathtoolsUtil.checkAlignment`, `BaseItems.ArrayItem` and `EqnArrayItem`, and `AmsItems.FlalignItem`. These commands previously used raw row preprocessing. They now execute in parser order, after dynamic command-map lookup and at the actual owning stack item.

`ArrowBetweenLines` closes its generated entry and row immediately. The starred form first supplies two empty entries. Its symbol uses a genuine child parser, and direct EndEntry/EndRow calls retain the current lexical environment. Tags finalize at that same row boundary. By contrast, authored or macro-generated `&` reaches an Entry boundary only after command argument readers have had the opportunity to consume it; a real Entry clears the array environment. The Flalign path uses that same argument ownership while retaining its declared-count check before parsing the next entry. Flalign is not an EqnArray owner for the Mathtools alignment-only methods.

The sibling methods must share this ordering. `Aboxed` reads its two parts through the source GetUpTo rule and inserts the original TeX expansion. VDots and ShortVDots retain their child-parser and direct-method semantics. Flush commands update the cached default row spacing rather than accumulating previous adjustments. This prevents mixed Arrow/box/dots commands from being reordered, omitted, or left in the wrong row. The separately merged boxed macro and GetArgument fixes remain prerequisites rather than duplicated implementations.

The corpus contains 4,450 unique inputs. `arrow_between_lines_mathjax_3_2_2.json` contains **4,018 exact complete SVGs**: 2,380 valid renderings and 1,638 original rendered errors. Against main141 (`dbdf950d9c64d7bb2c1c55784308648a4c28b02b`), 2,320 of these fail and 1,698 already pass. There are no formerly exact regressions in the full corpus. The unit test also checks the source child-parser macro counter boundary, and restoration on success and failure.

Coverage includes all currently supported EqnArray owners, plain/starred and custom-symbol arrows, authored/empty/final rows, multiple entries, mixed sibling ordering, lexical fonts and colors, tags and references, optional arguments and whitespace, genuine groups/fences/arguments, pending function/style/position/prime/not/dots/infix items, shared definitions, dynamic operator/paired overrides, generated entries, comments, Flalign/XalignAt overflow timing, and controls without these commands. The original `eqnarray` environment remains a separate unsupported feature; this fixture does not claim that registration is fixed.

The other **432 original receipts** are preserved in `arrow_between_lines_residuals.json` alongside the unmodified main141 and candidate outputs. They are not accepted goldens or counted as passing. Of these, 254 have valid original SVGs and 46 are original runtime failures. The rest contain original rendered errors. Known families include the separate initial-operator prefix-node gap, l/r gathered alignment policy, nested/raw-row capture and diagnostics, helper whitespace/runtime cases, and safe XML attribute escaping.

The 166 changed valid residuals were reviewed as actual browser-rendered baseline/original/candidate SVGs in both semantic families and representative screenshots. Fifty differ only by empty `mi` prefix nodes addressed by the separately held initial-operator work; four differ only by escaping a tag's ampersand in an XML id. The remaining 112 use the existing l/r gathered table policy. Eighty-eight literal-row expansion controls reproduce the same original SVGs and candidate SVGs on the unchanged prior parser. None of the changed valid inputs becomes a candidate error. The same-glyph comparisons have no worsened aggregate glyph bounds above 0.01 CSS px. Raw originals remain untouched; these comparisons are analysis, not fixture normalization.

A replay of all 2,602 previously published unique residual inputs on main141 finds 149 newly exact results, no exact regressions, and six changed nonexact results. Those are two modes each of the inherited gathered optional-alignment layout, nested-fence ampersand diagnostic, and multline raw-row diagnostic ordering. An independent 140-case source/stack-order review remains completely exact after the main141 rebase.

To regenerate both inventories from verified frozen assets:

```sh
python3 testdata/generate_arrow_between_lines.py /path/to/oracle-assets /path/to/node
```

The generator uses the committed inventory without filtering on Go success. Both files reproduce byte-for-byte. The public reference test compares every original SVG in full, without stripping nodes, attributes, paths, or numbers.
