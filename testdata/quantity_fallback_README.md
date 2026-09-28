# Physics Quantity token fallback

These references exercise the public D2 MathJax 3.2.2 bundle, at source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. The fixture contains complete,
unmodified original SVG strings; the residual report retains full original,
baseline, and candidate observations separately. No Go output is a golden.

The source of this change is [PhysicsMethods.Quantity](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMethods.ts#L72).
It uses `GetNext`/`GetCS` to peek for a supported delimiter, optionally after
`big`, `Big`, `bigg`, or `Bigg`. When the peek finds another control sequence,
it pushes empty `ParseUtil.fenced` delimiters and restores the input cursor.
The following command therefore still belongs to the caller. This fallback
precedes the mandatory-brace check for argument-taking registrations; stars
are consumed only for those registrations. Other unsupported tokens cause an
empty fallback for `qty`/`quantity` and the source missing-argument diagnostic
for argument-taking registrations.

The previous implementation consumed an unsupported token as a child-parser
argument. For example, `\Huge \qty\bf x+y` lost the bold declaration, whereas
the original produces empty parentheses followed by bold `x+y`. The new
fallback also preserves caller macro definitions, pending parser items, and
subsequent spacing commands. The `norm` fallback passes its literal registered
`\|` strings to `ParseUtil.fenced`, as the original does; it does not run the
TeX delimiter command in that branch.

Supported braced and automatic-delimiter arguments still use the existing
handlers. Their remaining limitations (including plain `qty` braced fences,
size-command argument handling, and automatic-delimiter style boundaries)
are preserved as raw controls. This change does not add the separately
missing `Bqty` registration or change the `Eval` handler. Original NEL runtime
failures and existing BOM/other whitespace differences are also retained.

On main143 (`d74de82316658af59f3d0e8a409b4bbeb005fb5d`), 2,584 distinct
original inputs give 2,486 exact references: 1,400 valid expressions and 1,086
original error renderings. The candidate fixes 1,996 baseline differences and
retains 490 exact controls, with no previously exact regressions. The 98 raw
inputs comprise 92 unchanged outputs and six changed outputs. Eight raw
references are original runtime failures (four of those change in Go); the
other two changed outputs combine the repaired bold continuation with an
inherited plain-`qty` braced-fence difference. That braced part remains
unchanged, while the first bold segment now matches the original.

An independent 280-input review supplies 272 exact references and eight
unchanged size-command argument controls, all included in these counts.
The complete frozen oracle suite, `go test -race -p 1 ./...`,
`go vet -p 1 ./...`, and `GOOS=js GOARCH=wasm go build -p 1 ./...` all pass
on main143. Both reference files regenerate byte-identically. A replay of
3,040 previously published residual inputs found four serialized changes in
already documented authored cancel-attribute ordering: two became byte exact
and two ceased to be byte exact in that particular run. All four complete
parsed SVGs, including every numeric geometry value, equal the original;
there are no rendered/source regressions. Complete outputs and the
attribute-order proof are retained, without normalizing fixture goldens.

Regenerate the original references with:

```sh
node --jitless testdata/generate_quantity_fallback.cjs /path/to/pinned-assets
```

The generator verifies SHA-256 hashes for all three frozen assets and uses a
fresh original VM for each expression. It does not invoke Go. Both display
modes are included, together with stars, errors, size-command lookahead,
control-space/EOF, JavaScript whitespace, nested fonts and scripts,
fractions, matrix cells, pending items, and dynamic command overrides.
