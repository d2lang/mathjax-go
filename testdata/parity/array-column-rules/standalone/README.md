# Standalone column-rule evidence

`standalone-comparison-focused.jpg` is the visually inspected comparison. `index.html` recreates it from the six unchanged raw SVG files. `standalone-comparison.jpg` retains the uncropped browser capture. Every panel uses the same 52px font size and the same original MathJax stylesheet, with no per-panel scaling or transformations.

The first row shows `\begin{array}{|c:c|}a&c\\b&d\end{array}`: the old output omits both outside frames and uses a solid middle rule. The fixed output restores both outside frames and the dashed middle rule. The second row shows `\begin{array}{c|cc}a&b&c\\d&e&f\end{array}`: the old output repeats a rule in the unmarked second gap. The fixed output leaves that gap empty. Both fixed SVGs are byte-identical to their original MathJax SVGs.

## Original CSS provenance

`extract-original.cjs` loads only the three SHA256-verified frozen MathJax assets, renders with the normal D2 oracle configuration, and calls the original output jax API:

```js
const styleNode = html.outputJax.styleSheet(html);
const css = adaptor.textContent(styleNode);
```

It writes `original-mathjax-svg.css` without modifications and also preserves the original serialized style element. `original-css-extraction.json` records all input asset hashes, the renderer commit, exact API method source, render options, wrapper attributes, and CSS/output hashes. The extracted CSS is 2,065 bytes with SHA256 `92e9394efc59fcac2ea1ba3883ab9ab1bd71450c2bd7e9df6b503760678f3bb5`.

The original stylesheet supplies `stroke-width: 70px` to table `line[data-line]` nodes and `stroke-dasharray: 140` to `.mjx-dashed`. The separate `comparison-layout.css` controls only the page layout, labels and uniform math font/scale. It does not add or replace math rule styling. `browser-verification.json` records computed styles: all six containers are 52px with no transforms; solid/dashed internal lines use 70px and dashed lines use a 140px dash pattern in SVG coordinates.

## D2 limitation

This is supplemental standalone browser evidence. D2 embeds the renderer SVG but does not carry the original MathJax stylesheet, so internal table rules are much thinner in its output. The D2 evidence separately demonstrates the restored outside frames. This comparison does not claim D2 includes the original stylesheet.

## Reproduction and bindings

Both scripts accept explicit input/output paths and use only standard Node.js/Python libraries. The original asset directory must contain `polyfills.js`, `mathjax.js`, and `setup.js` with the pinned SHA256 values embedded in the extractor. A baseline/candidate probe must read JSONL objects with `tex`, `display`, and `options.Display`, and write one JSONL object containing `svg` (or an `error`) for each input. The included `probe/main.go` uses only the public `github.com/d2lang/mathjax-go` API. Build this same source from each module checkout as shown below; the reproduction scripts do not depend on a particular checkout layout. Explicit top-level `display` takes precedence over `options.Display`; when both are omitted, the public API defaults are used. The probe accepts one request per line and returns `svg` plus an optional `error` per line. TeX syntax errors remain ordinary MathJax error SVGs, as specified by the public API.

Choose a separate output directory. The scripts refuse to overwrite this preserved evidence directory. The extractor writes fresh original SVGs/CSS and its API receipt there. The builder invokes the supplied probes, verifies every raw SVG/CSS byte against this evidence package, checks that fixed/original SVGs are identical, and recreates the three HTML pages. It writes a new `reproduction-verification.json`; it never rewrites this package's historical `manifest.json` or extraction receipt. Rebuilt probe hashes are recorded as current inputs rather than asserted to equal publication binary hashes.

Build both probes using the same absolute source file path and distinct module checkouts. Use the baseline commit `42932ee918ca4db7f3edf9da30e59b0c9cd61f86` and fixed source commit `e6656bbd1429d8f402ed08c57698ae1f1cb68bd6` (or the published PR commit with the same production source). The `-C` directory selects which checkout supplies the imported public package; no module replacement is needed.

```sh
go -C /path/to/baseline-checkout build -p=1 \
  -o /path/to/baseline-jsonl-probe \
  /path/to/standalone-evidence/probe/main.go

go -C /path/to/candidate-checkout build -p=1 \
  -o /path/to/candidate-jsonl-probe \
  /path/to/standalone-evidence/probe/main.go
```

Portable invocation (replace each absolute path with your local input location):

```sh
node --jitless /path/to/standalone-evidence/extract-original.cjs \
  --asset-dir /path/to/frozen-original-assets \
  --output-dir /path/to/new-reproduction

python3 /path/to/standalone-evidence/build-evidence.py \
  --baseline /path/to/baseline-jsonl-probe \
  --candidate /path/to/candidate-jsonl-probe \
  --reference-dir /path/to/standalone-evidence \
  --output-dir /path/to/new-reproduction
```

`--reference-dir` defaults to the directory containing `build-evidence.py`. The optional `--original-corpus /path/to/corpus.json.gz` adds a check against an independent original-only corpus with `cases[].tex`, `display`, and `original.svg`; it is not needed to reproduce the self-contained evidence package. Serve the new output directory with an ordinary local HTTP server and open `index.html`. Screenshots require a browser and are not regenerated by these scripts.

Exact locally verified invocation with the previously preserved probes (the fresh builds below also reproduce the same bytes):

```sh
/Users/aw/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/bin/node --jitless \
  /Users/aw/Documents/Codex/2026-09-27/we-x20/work/array-rules-audit/standalone-evidence/extract-original.cjs \
  --asset-dir /Users/aw/Documents/Codex/2026-09-27/we-x20/work/oracle-assets \
  --output-dir /Users/aw/Documents/Codex/2026-09-27/we-x20/work/array-rules-audit/portable-reproduction-157

python3 /Users/aw/Documents/Codex/2026-09-27/we-x20/work/array-rules-audit/standalone-evidence/build-evidence.py \
  --baseline /Users/aw/Documents/Codex/2026-09-27/we-x20/work/under-over-mapping-audit/candidate156-final-newline \
  --candidate /Users/aw/Documents/Codex/2026-09-27/we-x20/work/array-rules-audit/candidate154-prerequisites \
  --reference-dir /Users/aw/Documents/Codex/2026-09-27/we-x20/work/array-rules-audit/standalone-evidence \
  --output-dir /Users/aw/Documents/Codex/2026-09-27/we-x20/work/array-rules-audit/portable-reproduction-157 \
  --original-corpus /Users/aw/Documents/Codex/2026-09-27/we-x20/work/array-rules-audit/prepared-originals156.json.gz
```

`manifest.json` binds the original before/after binaries and commits, six raw SVGs, original CSS receipt, HTML, layout CSS, screenshots and browser observation. Fresh originals also matched the previously captured independent originals exactly. The focused screenshot is a browser capture cropped only to the complete comparison page; all six panels and the explanatory footer remain visible. No image content was edited.

## Fresh public-API probe verification

The following two lightweight builds used the actual baseline and final candidate checkouts without changing either checkout or a module file:

```sh
/Users/aw/.local/share/go/1.27.0/bin/go \
  -C /Users/aw/Documents/Codex/2026-09-27/we-x20/work/mathjax-go build -p=1 \
  -o /Users/aw/Documents/Codex/2026-09-27/we-x20/work/array-rules-audit/portable-probe157-baseline \
  /Users/aw/Documents/Codex/2026-09-27/we-x20/work/array-rules-audit/standalone-evidence/probe/main.go

/Users/aw/.local/share/go/1.27.0/bin/go \
  -C /Users/aw/Documents/Codex/2026-09-27/we-x20/work/mathjax-reversible build -p=1 \
  -o /Users/aw/Documents/Codex/2026-09-27/we-x20/work/array-rules-audit/portable-probe157-candidate \
  /Users/aw/Documents/Codex/2026-09-27/we-x20/work/array-rules-audit/standalone-evidence/probe/main.go
```

The extraction/build invocation above was then repeated with `--output-dir .../array-rules-audit/portable-public-probe-reproduction-157`, `--baseline .../array-rules-audit/portable-probe157-baseline`, and `--candidate .../array-rules-audit/portable-probe157-candidate`. All six raw SVGs, both CSS files, the original style element and receipt, and all three HTML pages reproduced byte-for-byte.
