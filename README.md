# mathjax-go

`mathjax-go` is a native Go port of the MathJax 3.2.2 TeX-to-SVG pipeline used
by [D2](https://github.com/d2lang/d2). Its production packages do not embed,
load, or execute JavaScript.

The compatibility target is D2's frozen custom component, not current MathJax.
That component enables the `base`, `mathtools`, `ams`, `amscd`, `braket`,
`cancel`, `cases`, `color`, `gensymb`, `mhchem`, and `physics` TeX packages and
uses SVG output with `fontCache: 'none'`.

```go
svg, err := mathjax.Render(`\frac{1}{2}`)
width, height, err := mathjax.Measure(`\frac{1}{2}`)
```

`Render` uses the same conversion settings as D2: display mode, a 16-pixel em,
an 8-pixel ex, and inline SVG glyph paths. `RenderWithOptions` exposes those
settings explicitly. Both functions return the bare `<svg>` element that D2
previously obtained from the inner HTML of MathJax's `<mjx-container>`.
`Measure` preserves D2's pixel sizing contract by rounding the SVG's ex-based
dimensions up with the configured default of 8 pixels per ex.

As in MathJax, malformed or unsupported TeX is normally rendered as an
`merror` SVG. A Go error instead indicates invalid conversion options or an
internal pipeline failure.

## Compatibility tests

The regular suite is pure Go:

```sh
go test ./...
go test -race ./...
go vet ./...
```

An optional test-only differential runner compares the port with D2's
cryptographically pinned MathJax bundle. Node.js and the JavaScript assets are
used only by this oracle and are not linked into the Go library:

```sh
MATHJAX_GO_ORACLE_DIR=/absolute/path/to/d2renderers/d2latex \
MATHJAX_GO_ORACLE_FULL=1 \
go test ./...
```

The directory must contain `mathjax.js`, `polyfills.js`, and `setup.js` with
the exact SHA-256 values recorded in [PROVENANCE.md](PROVENANCE.md).

## License

Apache License 2.0, with identified MIT-derived portions. See
[LICENSE](LICENSE), [LICENSES](LICENSES), [NOTICE](NOTICE), and
[PROVENANCE.md](PROVENANCE.md).

### TeX mu dimension references

The TeX parser converts valid, fully extracted `mu` values with the pinned
`ParseUtil.muReplace/Em` precision before creating MathML. The existing dimension
scanner and direct MathML length conversion are separate and unchanged.

The dimension corpus keeps 86 full SVG/explicit/own-tree records: 68 raw-primary
results (including eight direct-MathML controls), 12 unchanged grammar boundaries,
two CD-height renderer boundaries with exact primary trees, and four raise/lower
consumer controls. The last four retain inherited structure and sign handling;
only their three dimension strings change. Their SVG hashes are candidate
regression captures, not claims of primary renderer parity. The untouched primary
records are retained beside every qualification.

Registered parser fixtures also bind 44 full-token conversion controls, three
large-number formatting cases, and 38 accepted extraction/cursor/error controls.
The extraction records explicitly preserve scanner differences rather than
claiming complete `GetDimen` grammar compatibility. Regenerate the primary files
with `testdata/generate_tex_mu_dimensions.cjs` and
`internal/tex/testdata/generate_mu_dimension.cjs`, passing the pinned asset
directory and output directory/file respectively.


### Math inside text boxes

The supported `text`, text-font, `hbox` and `mbox` commands parse embedded
`$...$` and `\(...\)` math with MathJax's empty inner lexical environment and
shared macro/tag configuration. Literal chunks preserve internal whitespace
and tilde; only the source's four literal escapes and edge whitespace rules
apply. `hbox` and `mbox` retain their level-zero wrapper, including empty input.

`TestInternalTextPinnedReferences` checks70 complete unmodified-primary SVGs
and explicit/own-property trees, plus20 unchanged shared-caller controls.
FBox/colorbox embedded math, AMS tag text, the extra `textsl` alias and missing
`textup` dispatch remain separately recorded boundaries. The28 actual-method
references also bind delimiter errors, cardinality, child ownership, font and
macro-counter isolation. Two private pre-postfilter script nodes retain the
existing Go `msub` versus primary two-child `msubsup` representation; public
compiler trees have no such normalization.

Regenerate the references with the pinned D2 MathJax3.2.2 assets:

```
node testdata/generate_internal_text.cjs PINNED_ASSETS [EVIDENCE_DIRECTORY]
node internal/tex/testdata/generate_internal_text_method.cjs PINNED_ASSETS
```
