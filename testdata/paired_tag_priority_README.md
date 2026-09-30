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

The preview overlap scan covers all 505 tracked testdata JSON/JSON.gz files at pending Braket checkpoint `41623db5f6d3c6840f193e69586865f5127989fe`. It finds 1,422 first-publication SVG inputs, six promotions from preserved raw GetStar/tag-whitespace originals, and four prior strict whole-SVG hash and AST assertions for bare `\ref`/`\eqref` in both display modes. All 14 runtime inputs are first publication. Each prior full original/hash matches, and there are no conflicts or ambiguous display modes. The same six raw inputs occur in two older fixtures and are counted once. These are preview counts; actual merged main178 must still be bound before publication. Per-fix inventories may overlap prior inventories and do not claim repository-wide unique coverage.

The source checkpoint is `791afca0ca7842fb622e94531257cbe6aa0f7828`, internal tree `fb1bd62ca6fb47a375856b64c41f4f932ff62f06`, compiled on pending Braket preview `41623db5f6d3c6840f193e69586865f5127989fe`. Independent own-corpus qualification passes for every original, and focused source/compatibility tests passed. The exported 1,446-input public test passed. Actual-base rebase and full frozen/race/vet/WASM gates remain pending root validation.

Independent broad replay retains all 17,549 original objects. The 6,800 published inputs gain 12 complete SVG fixes; the 4,326 upstream inputs have no genuine semantic change, and neither corpus has a regression or changed nonexact result after separate attribute-order qualification. Five serialization controls were repeated 640 times: 512 outputs from four cancel controls equal complete original XML; 128 outputs from an inherited unsupported-enclose control agree only between baseline and candidate and remain nonexact to the original undefined-command error. All 5,621 shared and 802 runtime outcomes remain unchanged. SVG test goldens remain byte-exact.

Protected replay covers 20,396 observations: 20,384 unchanged and 12 improvements representing the same six tag/ref/eqref inputs in two residual corpora. All 830 runtime observations remain unchanged. GetStar strict, Quantity, Eval, HLine, Expectation, KetBra, Braket environment, paired-name strict/residual, public paired, ordinary tag and gather-tag controls retain their prior outcomes. These overlapping checks are not added to the public fixture count.

The captured Huge/boxed `tag` and `eqref` fraction inputs are already in the public union. Final D2 before/after/original screenshots and raw SVG verification remain pending root evidence publication.
