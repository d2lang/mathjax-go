# Original JavaScript object behavior in keyval options

MathJax 3.2.2 `ParseUtil.readKeyval` stores primitive option values in an ordinary
JavaScript object. Assigning `__proto__` consumes its value but does not create an
own property. `Object.keys` visits canonical array-index keys before other own
keys. Go previously treated `__proto__` as an invalid option and reported the
first inserted invalid option even when a numeric key should take precedence.

For example, the original accepts this label, while the previous Go version
renders `Invalid option: __proto__`:

```tex
\mathtoolsset{__proto__=true,centercolon=true}\text{Rate: }x:y = \frac{a}{b}
```

The shared collection and validation now preserve these object semantics for
Mathtools and Empheq. The option-value scanner is unchanged. Duplicate keys
retain their first insertion position and use their last value; malformed later
values still take precedence over invalid earlier option names.

## Frozen original fixtures

`keyval_object_properties_mathjax_3_2_2.json.gz` contains 194 complete,
unmodified SVGs from D2's frozen MathJax 3.2.2 bundle. Each expression uses a fresh
runtime. The inventory contains 110 valid original SVGs and 84 original
rendered-error SVGs, with every error case retained. Against the preceding Go
implementation, 134 cases differ: 108 valid originals become renderable and 26
rendered-error diagnostics change. The other 60 cases are unchanged controls.
After the fix, all 194 SVGs match byte for byte.

`internal/tex/testdata/option_keyval_object_mathjax_3_2_2.json` contains 90 direct
observations of the original exported `ParseUtil.keyvalOptions`, preserving
options, own-key enumeration order, and structured errors. Each call uses a
fresh frozen runtime, with either no allowlist or Empheq's `left`/`right`
allowlist. There are 58 successful observations and 32 structured errors. All
match the shared Go helper; the 45 Empheq observations also test its adapter.
The frozen D2 setup does not enable the Empheq package, so these direct helper
observations do not claim Empheq SVG parity.

The captures cover primitive `__proto__` values, braced and BOM-trimmed keys,
other prototype property names, duplicate keys, canonical array-index boundaries,
non-index numeric spellings, and diagnostic precedence. Source-bound regressions
also verify that the ordered map's existing insertion-order API stays unchanged.

Source commit: `ad8f5c21cb810236551da8c6512ba733e67357ee`.

Regenerate using the exact frozen assets; the script checks every asset hash and
fails on unexpected runtime exceptions rather than dropping cases:

```sh
node --jitless testdata/generate_keyval_object_properties.cjs PINNED_ASSETS
```

The generator reads only original MathJax outputs. The public test compares the
complete SVG string, including each original rendered-error SVG.
