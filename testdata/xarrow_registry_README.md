# Active xArrow mappings and optional lower labels

D2's frozen MathJax 3.2.2 setup loads Mhchem after AMS and Mathtools. Its command
map overrides `xleftrightarrow` padding to 6/6 mu and `xrightleftharpoons` to 5/7
mu, and registers `xRightleftharpoons` and `xLeftrightharpoons` with the same 5/7
mu padding. Go previously used the earlier Mathtools 10/10 mu values and did not
dispatch the two additional spellings.

The xArrow registry and dispatch now use the existing translated command maps
in their source order, including those final Mhchem entries. The existing
xArrow construction also follows the source's truthiness check for the optional
lower-label string: an explicit empty `[]` creates no lower label. Missing-close
bracket errors name the invoked arrow, matching `GetBrackets`.

These display and inline D2 witnesses are included in the reference corpus:

```tex
A\xRightleftharpoons[k_2]{k_1}B\xLeftrightharpoons[k_4]{k_3}C
A\xleftrightarrow{k}B\xrightleftharpoons{k}C
```

## Frozen original verification

`xarrow_registry_mathjax_3_2_2.json.gz` preserves 1,262 complete original API
objects from fresh frozen MathJax 3.2.2 runtimes: 1,126 valid SVGs and 136
rendered-error SVGs. There are no original runtime failures and no omitted cases.
All 1,262 SVGs match byte for byte after the fix. Against the preceding Go source
`521f3bb`, 324 cases differ: 278 valid renderings and 46 rendered-error diagnostics;
the other 938 cases are unchanged exact controls.

The matrix covers all 17 active xArrow command names, absent/empty/authored lower
labels, fraction labels, empty labels, own and grouped scripts, nested script
styles, fonts/colors, text/box child parsers, matrices, malformed/missing
arguments, and higher-priority paired/operator registrations. The unaffected
arrow mappings serve as independent registry and geometry controls.

Source commit: `ad8f5c21cb810236551da8c6512ba733e67357ee`. The original-only
regenerator verifies the pinned D2 asset hashes before conversion and fails on
unexpected runtime exceptions rather than discarding an input:

```sh
python3 testdata/generate_xarrow_registry.py PINNED_ASSETS NODE
```
