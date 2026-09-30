# Paired delimiter priority for tag and reference names

Registered paired delimiters take precedence over ordinary tag/reference commands. The original Mathtools paired map has priority -5, above AMS dynamic operators at -1 and ordinary command maps at 5. A paired definition named `\tag`, `\notag`, `\eqref`, `\label`, `\ref`, `\nonumber` or `\refeq` must reach the paired handler before a tag handler can consume arguments, change tag/label state or emit a diagnostic. These names originate in AMS, Base and Mathtools respectively; they are not all AMS registrations.

The existing `paired` lookup now guards the call to `amsTagCommand`. Its following Mathtools dispatch remains unchanged. Paired expansions still execute in the same parser, use their original argument readers and charge the original macro budget. Ordinary unregistered tag/reference commands retain their existing path.

The active paired option reader also retains the invoked command name in `MissingCloseBracket`, matching the original `GetBrackets` use of `currentCS`. A local adapter handles only that error. The generic bracket reader, `ExtraCloseLooking` precedence and every other diagnostic remain unchanged. This covers ordinary paired names as well as names that override tag commands.

## Complete originals and qualification

`paired_tag_priority_mathjax_3_2_2.json.gz` retains 1,446 distinct TeX/display pairs: the original 1,206 priority inputs plus 240 disjoint ordinary-name diagnostic controls. Its 1,432 complete SVG assertions comprise 1,106 valid renderings and 326 error renderings. All original objects are unchanged, including the six retained discovery cases. The remaining 1,440 inputs were captured during this audit.

The candidate fixes 1,170 complete SVGs: 938 valid renderings and 232 error renderings. It preserves 262 exact controls, with no regression or changed nonexact SVG. The call guard accounts for 966 fixes; the local diagnostic adapter accounts for 204 error-SVG fixes, including 84 original priority cases and 120 ordinary-name cases. The initial failed candidate and all 84 residuals remain preserved in the audit; no expected error text was rewritten to make them pass.

The separate 14 original runtime inputs contain literal NEL and retain complete exceptions beginning `TypeError: Cannot read properties of null (reading '4')`. The test requires the bounded API error `no Unicode range for character U+0085` and empty SVG. Ten replace prior incorrect rendered successes; four preserve existing bounded errors. These outcomes do not count as SVG fixes or claim original exception-message parity. There is no omitted raw SVG partition.

The corpus covers all seven intercepted names and six declaration aliases, call modifiers and malformed arguments, declaration order and redefinition, unregistered commands, lexical/child contexts, tag/label state, array/environment owners, pending recipients, generated helper/XPP bodies, Unicode whitespace and supported macro-budget boundaries. The ordinary-name supplement covers `abs` and `auditpair`, all six aliases, valid/empty/starred options, malformed/grouped/escaped brackets, `ExtraCloseLooking`, and MathFont/fraction child contexts. Family memberships can overlap and are not added to the unique input count.

## Passive observations and literal bindings

`paired_tag_priority_observations.json.gz` preserves 168 passive priority inputs already in the public fixture. It records every selected map, ordered applicable map list, parser source/cursor and current control sequence. Its 182 events select the paired map 154 times, AMS macros 12 times, Base macros 12 times and Mathtools macros four times. The observer records the state, calls original `SubHandler.parse` once and returns the unchanged result. Independent actual `CommandMap.parse` observations confirm that the selected map is the one dispatched.

`paired_tag_priority_bracket_observations.json.gz` preserves 72 passive `GetBrackets` records within the ordinary-name supplement. Each records the reader's argument name, `currentCS`, source and cursor before one original call. Every relevant invocation binds `currentCS` to the invoked paired name. Complete observed outputs equal independently uninstrumented originals.

The two observation sets contain 240 distinct existing public inputs. `paired_tag_priority_literal_bindings.json.gz` preserves 126 complete target-to-source-expansion bindings over 12 distinct literal input pairs already in the same public union. Every target and literal original agrees. Neither observations nor bindings add public inputs or duplicate assertion counts.

## Original-only regeneration

The frozen source is MathJax commit `ad8f5c21cb810236551da8c6512ba733e67357ee`; the public fixture includes the three pinned D2 asset hashes.

```sh
node --jitless testdata/generate_paired_tag_priority.cjs /absolute/path/to/oracle-assets /absolute/path/to/scratch-output
```

Use a separate scratch output directory. No Go process supplies expectations. The generator re-renders all 1,446 inputs and requires identical complete SVG objects. Historical runtime objects remain intact; fresh complete source exceptions are saved separately with the same null-range first-line and literal-NEL contract. JSONL transport escapes line/paragraph separators.

It independently repeats every field of all 240 passive records and requires their complete outputs to equal uninstrumented fresh originals. It rebinds every literal record to the freshly rendered public union. All four gzip files regenerated byte-identically. The inventory records their hashes and the regeneration receipt. The public Go test checks source pin, unique names/input pairs, complete SVG strings, well-formed original XML, exact partitions and the bounded-runtime contract without normalizing SVG assertions.

## Publication overlap and validation

The fresh publication overlap scan covers all 505 tracked testdata JSON/JSON.gz files at actual merged Braket PR178 `b08b35acc9804d816425624a447945164830fa51`. It finds 1,422 first-publication SVG inputs, six promotions from preserved raw GetStar/tag-whitespace originals, and four prior strict whole-SVG hash and AST assertions for bare `\ref`/`\eqref` in both display modes. All 14 runtime inputs are first publication. Each prior full original/hash matches, with no conflicts or ambiguous display modes. The same six raw inputs occur in two older fixtures and count once. Independent consumer review confirms these classifications. Per-fix inventories can overlap and do not claim repository-wide unique coverage.

The source checkpoint `791afca0ca7842fb622e94531257cbe6aa0f7828` was independently compiled and qualified with internal tree `fb1bd62ca6fb47a375856b64c41f4f932ff62f06`. Source and fixture patches rebase byte-identically onto actual merged PR178; rebased source is `71a7911e944b711d115c4b1d8f89d89560becb73` and fixture checkpoint is `a7ef6c81ec6ecfb28d5069dfac61d457364eac16`. Only the two inherited Braket validation documents differ from the preserved preview. The exported 1,446-input public test passed. Full frozen/race/vet/WASM gates on this rebased checkout remain pending.

Independent broad replay retains all 17,549 original objects. The 6,800 published inputs gain 12 complete SVG fixes; the 4,326 upstream inputs have no genuine semantic change, and neither corpus has a regression or changed nonexact result after separate attribute-order qualification. Five serialization controls were repeated 640 times: 512 outputs from four cancel controls equal complete original XML; 128 outputs from an inherited unsupported-enclose control agree only between baseline and candidate and remain nonexact to the original undefined-command error. All 5,621 shared and 802 runtime outcomes remain unchanged. SVG test goldens remain byte-exact.

Protected replay covers 20,396 observations: 20,384 unchanged and 12 improvements representing the same six tag/ref/eqref inputs in two residual corpora. All 830 runtime observations remain unchanged. GetStar strict, Quantity, Eval, HLine, Expectation, KetBra, Braket environment, paired-name strict/residual, public paired, ordinary tag and gather-tag controls retain their prior outcomes. These overlapping checks are not added to the public fixture count.

The final D2 evidence uses two already-captured public inputs: the Huge/boxed `eqref` fraction and an ordinary paired malformed-option diagnostic. Both screenshots were inspected at one shared scale. Fixed and original D2 SVGs are byte-identical, and before differs in each case. All 93 saved D2 witnesses match their original SVGs. Initial `tag` and `notag` before-render attempts were preserved as failures because their percentage-width SVGs cannot be measured by the existing D2 integration; they are not claimed as before-render successes. Their full public SVG parity remains asserted.