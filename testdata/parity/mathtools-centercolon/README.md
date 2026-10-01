# Mathtools centered-colon D2 evidence

Enabling `centercolon` previously rewrote all remaining source, so `\text{Time: 2}` printed the literal `\centercolon` command. Disabling the option later also left the remaining colons centered. The fix handles a colon when it is parsed and uses the current option without rewriting text or command arguments. SetOptions uses the original balanced key/value reader and validates the complete list before applying changes.

`comparison.png` is a Chromium screenshot of the three untouched complete D2 SVGs at one shared scale. The before and after diagrams use the same D2 source and renderer. The original diagram is independently rendered through the frozen, hash-verified MathJax 3.2.2 bundle. The after/original SVGs and PNGs are byte-identical. All images were visually inspected.

The public regression fixture compares 134 complete original SVGs: 120 valid renderings and 14 original rendered errors. It fixes 68 valid renderings and two error SVGs, and preserves 64 controls. Every conversion uses a fresh original VM. The shared key/value reader uses JavaScript whitespace: BOM is trimmed while NEL remains literal. A separate fixture directly observes original `ParseUtil.keyvalOptions` with Empheq's left/right allowlist, covering 16 whitespace and ordinary controls. Regenerate both fixtures with `node testdata/generate_mathtools_options.cjs PINNED_ASSETS`.

To regenerate the diagrams, use D2 at the renderer commit in `manifest.json`. Build it with a replacement for mathjax-go at the before revision or fixed source revision, then run:

```sh
go mod edit -replace=github.com/d2lang/mathjax-go=/absolute/path/to/mathjax-checkout
go build -o /tmp/d2-parity .
/tmp/d2-parity --pad 24 --scale 1 --omit-version witness.d2 output.svg
```

For the independent original, use `testdata/parity/accent-italic-correction/oracle-adapter` as the replacement, with `MATHJAX_GO_NODE`, `MATHJAX_GO_ORACLE_SCRIPT`, and `MATHJAX_GO_ORACLE_DIR` pointing to Node, `testdata/differential/oracle.mjs`, and the pinned assets. The original driver starts a fresh runtime per render. Artifact hashes are in `manifest.json`. This finite regression corpus does not establish universal MathJax parity.
