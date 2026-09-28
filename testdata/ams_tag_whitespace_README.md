# AMS tag whitespace references

This change restores two rules from frozen MathJax source
`ad8f5c21cb810236551da8c6512ba733e67357ee`:

- `AmsMethods.HandleTag` uses `ParseUtil.trimSpaces`, including its preservation
  of one terminal ASCII control-space. JavaScript whitespace includes BOM and
  excludes NEL.
- `Tags.formatId` replaces JavaScript whitespace in equation IDs. NEL remains
  part of the ID instead of becoming an underscore.

The source scope is equation tags and their IDs. The original inventory covers
starred tags, groups, font declarations, styles, active operator overrides,
references, equation ownership, and 998/999/1000-invocation boundaries.

## Complete original observations

There are 496 unique `(tex, display)` inputs. The strict fixture contains 474
complete, unmodified original SVGs: 446 valid renders and 28 error renders.
Against main168 (`f0c55a3c63d1bb7ffd777f3a300dc7d250832b18`), all 474 match:
197 fixes and 277 unchanged controls. Tests compare the entire SVG and reject
duplicate names or inputs. Both baseline and candidate outputs for all 496
inputs are unchanged from the earlier main167 composition.

The residual file retains the 22 other original observations and historical
main162/candidate outputs. Current outputs remain unchanged: six paired-
declaration dispatch controls and sixteen GetStar BOM/NEL controls. These
observations are not accepted Go goldens or passing SVG references.

A scan of 444 JSON files on main168 confirms that 468 strict inputs are first
publications, four had earlier original SVG hashes, and two promote earlier
complete raw originals. The earlier hashes match the complete originals; no
stored original differs. All 22 residual inputs are first observations, outside
the passing count. Thus 474 strict assertions are not 474 new inputs.

## Validation

The current candidate was replayed against 6,416 saved published residual inputs
and 4,326 upstream MathJax-Tests TeX/display inputs under D2's frozen package
configuration. The published replay has two genuine tag fixes, already included
in the strict fixture, and no genuine regressions. The upstream replay has no
changes. Six cancel/cancelto serialization changes, including apparent fixes
and regressions, were excluded after 64 renders per binary per input (768 total)
matched the complete original XML trees, including attributes, text, tails and
ordered children. Original strings remain unmodified.

Focused tests, the full frozen-oracle suite, race tests, vet and WebAssembly
build pass on the publication base. The D2 comparison uses a literal U+0085 NEL
tag: the previous implementation removes it; the fixed output preserves its
parenthesized label. All 76 D2 witnesses match original MathJax byte-for-byte.

## Regeneration

Run `node --jitless testdata/generate_ams_tag_whitespace.cjs /path/to/assets`.
The generator verifies all three frozen asset SHA-256 values, then regenerates
the strict and residual original fields in fresh oracle processes. It never
calls Go or selects cases by candidate output. Regeneration is byte-identical.
Original runtime errors, when present in an inventory, remain observations
rather than SVG goldens.

The original inputs were captured independently of candidate outcomes. The
committed inventory preserves the full observations; the companion D2 evidence
includes raw SVGs, screenshots, hashes and reproduction details.
