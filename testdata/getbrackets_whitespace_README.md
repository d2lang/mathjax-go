# GetBrackets whitespace

The source chain is `TexParser.GetBrackets` → `GetNext` → `nextIsSpace`; the last method uses JavaScript `/\s/`. The Go optional-argument reader now uses that exact whitespace set at entry: BOM U+FEFF is skipped, NEL U+0085 is retained. Ordinary math tokenization and other argument readers keep their existing, separately scoped behavior.

`internal/tex/testdata/getbrackets_whitespace_mathjax_3_2_2.json` records 198 direct calls to the frozen original method, including its original method bodies. These check the returned argument/default, presence, consumed byte cursor, and remaining source for all 25 JavaScript whitespace code points, four excluded controls, empty/default arguments, nested/escaped contents, and mixed whitespace. The private generator calls the actual original methods without substituting implementations.

The public fixture contains 438 complete, unmodified original SVGs across sqrt, Physics differential, continued fractions, extensible arrows, color, MmlToken, Physics expression, and Mathtools bracket consumers. It covers all 25 accepted whitespace characters, optional/absent/empty arguments, malformed brackets, and ordinary controls. All 438 match byte for byte; 46 differed on baseline `cbe99b2`. Public and private generators validate all three frozen D2 assets at source commit `ad8f5c21cb810236551da8c6512ba733e67357ee` and run with Node `--jitless`.

Run `node --jitless testdata/generate_getbrackets_whitespace.cjs /path/to/assets` for public SVGs and `node --jitless internal/tex/testdata/generate_getbrackets_whitespace.cjs /path/to/assets` for method records. The JSONL generator escapes U+2028/U+2029 in transport; the original runtime still receives the authored code points.

Separate existing followups: non-whitespace characters left to `GetArgument` or the ordinary math tokenizer can still differ from original parsing; an explicitly empty lower label in `\xrightarrow[]{a}` still creates an extra empty script. The shared GetBrackets results for those inputs are covered by direct original-method assertions, without changing the caller behavior in this fix.

D2 witness: `\sqrt<BOM>[3]{x} + \dd<BOM>[2]{x}`, replacing each `<BOM>` with an actual U+FEFF character immediately after the command.
