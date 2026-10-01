# Original Base and AMS command registrations

The frozen original registers `stackbin` as the two-argument macro
`\mathbin{\mathop{#2}\limits^{#1}}`, next to `stackrel`. It also maps the escaped
nonbreaking-space control symbol (backslash followed by U+00A0) to the Base
`Tilde` handler, which emits a nonbreaking-space `mtext` token. Go omitted those
registrations and rendered unknown-command errors for both valid inputs.

The implementation adds those exact entries to the existing macro and command
dispatch. The frozen mappings are already present in the translated source
tables, and the existing macro argument and `space` handlers supply the behavior.

These display and inline D2 witnesses are explicit reference cases:

```tex
A\stackbin{k}{+}B = C
A\ B\ C = D
```

The second line contains actual U+00A0 characters after each backslash. Its
escaped spelling for inspection is `A\\\u00a0B\\\u00a0C = D`.

## Frozen original verification

`base_command_registration_mathjax_3_2_2.json.gz` preserves 992 complete original
API objects from fresh frozen MathJax 3.2.2 runtimes: 824 valid SVGs and 168
rendered-error SVGs. There are no runtime exceptions or excluded cases. All 992
SVGs match byte for byte after the fix. Against the preceding Go source
`521f3bb`, 254 cases differ: 246 valid renderings and eight rendered-error
diagnostics; 738 cases are unchanged exact controls.

The matrix covers binary/relation spacing, fraction and empty labels, grouped
and own scripts, nested script styles, fonts/colors, text/box child parsers,
matrices, missing/malformed macro arguments, and higher-priority paired/operator
registrations. `stackrel`, `space`, raw tilde, and raw nonbreaking-space are
independent controls. The source-unregistered `nobreakspace` name is preserved as
an original rendered-error control.

Source commit: `ad8f5c21cb810236551da8c6512ba733e67357ee`. The original-only
regenerator verifies the pinned D2 asset hashes and fails on unexpected runtime
exceptions rather than discarding inputs:

```sh
python3 testdata/generate_base_command_registration.py PINNED_ASSETS NODE
```
