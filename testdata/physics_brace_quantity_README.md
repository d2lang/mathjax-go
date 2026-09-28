# Physics brace quantities

The references come from the original D2 MathJax 3.2.2 bundle, pinned to
`ad8f5c21cb810236551da8c6512ba733e67357ee`. Exact tests compare the entire
unmodified original SVG. The separate residual file preserves original,
baseline, and candidate SVGs without normalizing attributes or geometry.

The source registration is
[`Bqty: ['Quantity', '\\{', '\\}', true]`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMappings.ts).
[`PhysicsMethods.Quantity`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMethods.ts#L72)
requires a braced operand and builds TeX using the registered delimiter
commands. Ordinary arguments use `\left\{ ... \right\}`; stars use fixed
`\{ ... \}`; the four literal size commands `big`, `Big`, `bigg`, and `Bigg`
select their corresponding left/right forms. A star still consumes an
optional size command, then uses the fixed fences.

The expansion runs in a new child parser, preserving the original copied
environment and shared configuration. Its inferred-row children are pushed
individually. This matters for a following script and for an unbraced script
operand: the first final item satisfies the script, while remaining items
return to the caller through the already implemented script-delivery path.

Unsupported control sequences use the existing source-defined Quantity
fallback. That branch creates empty fences from the literal registration
strings and restores the token cursor. It does not parse the delimiter
strings as TeX. Missing arguments and malformed input retain the original
error-rendering controls.

The inventory includes both display modes; all four sizes and stars;
fractions, roots, scripts, primes, pending functions and positions; font and
color scopes; matrices, CD, Eqnarray, and MathLap compositions; JavaScript
whitespace and cursor fallback; and operator/paired-delimiter overrides of
the command and expansion helpers. Existing Quantity registrations and
explicit literal expansions are retained as independent controls.

## Reproduce the original references

```sh
node --jitless testdata/generate_physics_brace_quantity.cjs /path/to/pinned-assets
```

The generator checks all three asset hashes and uses a fresh original VM
for every expression, in bounded subprocess batches. It regenerates the
original SVG fields in both exact and raw inventories. It never executes Go
or selects which references are asserted. The accompanying public-overlap
receipt counts distinct input/display pairs already published at the tested
baseline, rather than claiming that per-PR fixture totals are globally unique
MathJax coverage.

Final baseline counts, residual qualification, and gate receipts are recorded
after composition with the preceding Eqnarray and MathLap fixes.
