# AMS tag whitespace references

The source change restores two caller-specific rules from the frozen MathJax
source at `ad8f5c21cb810236551da8c6512ba733e67357ee`:

- `AmsMethods.HandleTag` uses `ParseUtil.trimSpaces`, including its special
  preservation of one terminal ASCII control-space. JavaScript whitespace
  includes BOM and excludes NEL.
- `Tags.formatId` replaces JavaScript whitespace in equation IDs. NEL remains
  part of the ID instead of being replaced with an underscore.

The source scope does not change GetStar, generic argument readers, paired
declaration priority, or ordinary math tokenization. Groups, font declarations,
styles, active operator overrides, references, equation ownership, and real
998/999/1000-invocation boundaries are included in the original inventory.

## Complete original observations

There are 496 unique `(tex, display)` inputs. The strict fixture contains 474
complete, unmodified original SVGs: 446 valid renders and 28 error renders. The
isolated tag candidate on main162 matches all 474; 197 differ from main162 and
277 are unchanged controls. Tests compare the entire SVG and reject duplicate
names or inputs.

The separate residual file retains all 22 other original observations and both
main162/candidate outputs. All 22 candidate outputs are unchanged from main162:
six paired-declaration dispatch controls and sixteen GetStar BOM/NEL controls.
They are not accepted Go goldens or counted as passing original references.

The provenance inventory scans 413 JSON files from actual main162. Of the 474
strict inputs, 468 are first-publication inputs, four were previously recorded
only with an original SVG hash, and two promote earlier complete raw originals.
The four earlier hashes exactly match the new complete original SVGs. The 22
residual inputs are also first-publication observations, outside the passing
count. Thus 474 strict assertions must not be described as 474 new inputs.

## Broader source qualification

The isolated main162 probe was compared with 5,654 retained published residual
inputs and 4,326 upstream inputs. Two published terminal-control-space tag
inputs became original-exact and are included in the strict fixture. No other
genuine source changes or regressions were found. Two apparent cancel fixes
and one changed enclose result were attribute-order-only observations: 64
renders per binary for each input were preserved and parsed as complete XML.
The enclose package is unavailable in the frozen original configuration, so
its XML equivalence only establishes unchanged Go behavior, not original
parity. Those serialization observations are excluded from fix counts.

The source and fixtures were subsequently rebased onto actual main163
(`c09d0460db08b943c49003fbe2409402f9c2fca4`). A fresh isolated candidate build
and the strict 474-reference test passed. Both baseline and candidate outputs
for all 496 saved inputs are byte-identical to their respective main162
outputs, retaining 197 fixes and no regressions or changed residuals. The
broader replay and provenance scan above remain explicitly main162 receipts.
Full Go gates have not yet been run for this fixture package.

## Regeneration

Run `node --jitless testdata/generate_ams_tag_whitespace.cjs /path/to/assets`.
The generator verifies all three frozen asset SHA-256 values, then regenerates
the strict and residual original fields in fresh oracle processes. It never
calls Go or selects cases by candidate output. Original runtime errors, when
present in an inventory, are retained as observations rather than SVG goldens.

The original inputs were captured independently of candidate outcomes. The
source capture, complete comparisons, overlap receipt, and serialization
repeats are retained in the task's `work/ams-tag-whitespace-audit` directory.
