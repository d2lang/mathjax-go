# Physics vector applications

The Go handlers for `gradient`, `grad`, `laplacian`, `divergence`, `div`, and
`curl` used fixed operator tokens and eagerly read a single argument. They
added parentheses around braced and unbraced operands, converted square
brackets to parentheses, and isolated argument declarations from the caller.
They also bypassed redefinitions of commands inside the operator.

The pinned MathJax 3.2.2
[PhysicsMappings.ts](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMappings.ts#L76)
registers these six names through two methods sharing
[`vectorApplication`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMethods.ts#L307).
This change reads those registrations from the retained source table and
reuses the existing translation of that helper:

| Commands | Parsed operator | Source stack item |
| --- | --- | --- |
| `gradient`, `grad` | `\vnabla` | `FnItem` |
| `laplacian` | `\nabla^2` | `FnItem` |
| `divergence`, `div` | `\vnabla\vdot` | completed MML |
| `curl` | `\vnabla\crossproduct` | completed MML |

The operator is parsed by a child parser with the inherited lexical
environment and shared definitions. Its source item is delivered before the
caller argument is inspected. That distinction matters in scripts: a script
recipient rejects a pending function item but accepts completed MML. Brace
arguments are reinserted into the continuing caller without retaining their
outer group. Parentheses and square brackets use the source `AutoOpen` item;
the original fence shape and event order are preserved. No argument is required.

A visible witness, whose entire fixed SVG equals the original, is:

```tex
\Huge\grad x+\div{\vb F}+\curl[\vb F]+\laplacian f
```

## Frozen original references

`physics_vector_applications_mathjax_3_2_2.json` contains 2,122 complete,
unmodified SVGs from D2's frozen MathJax 3.2.2 bundle: 1,476 valid renderings
and 646 original error SVGs. The fixture records source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee` and all three verified bundle hashes.
Each reference uses a fresh VM, font cache none, `em=16`, `ex=8`, and the
recorded display mode.

Against main126 (`3f16eb233d573625bc180b2807fb9e31370e2fd9`), 1,562 of these
references failed and now match exactly. The 2,188-case audit has no formerly
exact regressions. All 72 vector controls preserved with the preceding
OperatorApplication change now match, including its 52 mismatches. Their
historical receipts remain unchanged in the earlier residual file.

The inventory covers all six registrations; no-argument, braced, unbraced,
and both automatic fence forms; declaration continuation and inner group
isolation; operator/paired overrides of both commands and operator internals;
fonts, colors, sizes and styles; scripts, primes, limits, positions, negation,
dots and pending functions; nested applications; fractions, roots, boxes,
Braket, Left/Middle, arrays and CD cells; and invalid syntax/error precedence.

## Raw inherited differences

`physics_vector_applications_residuals.json` retains 66 complete original,
baseline, and candidate responses. None was exact at baseline. They are not
asserted as correct or substituted with Go-generated expected output:

- 12 function-successor cases expose existing omission of ApplyFunction before
  a following `\mathop{+}` function item.
- 18 explicit `\bigl`/`\bigr` fence-sizing cases retain an existing renderer
  difference. Included `\Res[\bigl[x\bigr]]`, larger `\Bigl`/`\Bigr`, and
  primitive `\left[\bigl[x\bigr]\right]` controls have identical baseline and
  candidate SVGs and independently demonstrate that behavior.
- 12 `\cr` cases retain an existing undefined-versus-misplaced diagnostic.
- 12 unmatched `\end{matrix}` cases retain an existing diagnostic difference.
- 12 misplaced ampersands retain the original's unescaped error attribute
  alongside Go's escaped XML output.

Regenerate both files using
`python3 testdata/generate_physics_vector_applications.py PINNED_ASSETS NODE`.
The generator checks the hashes and updates only fresh original responses,
preserving the raw historical baseline/candidate receipts. It runs bounded
batches of 24 fresh VMs. Run `go test ./... -run TestPhysicsVectorApplicationReferences`
for the complete exact-reference check.
