# Physics named-function registrations

`physics_named_functions_mathjax_3_2_2.json` contains 1,000 complete SVGs
from D2's frozen MathJax 3.2.2 bundle at upstream commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. Its three original asset hashes
are recorded in the fixture and checked by the generator. Every input runs
in a fresh VM with `em=16`, `ex=8`, no font cache, and its recorded display
mode. Original SVGs are compared verbatim, without normalization or Go
output replacing original references.

The source's
[`Physics-expressions-macros`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMappings.ts)
registers 29 long command names as `NamedFn` with an explicit identifier.
[`PhysicsMethods.NamedFn`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMethods.ts)
is an alias for
[`BaseMethods.NamedFn`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts),
which creates an `mi` with OP TeX class and pushes an `FnItem`.
The Go function-name dispatcher already implements that handler. The fix
reads these registrations from the retained source map instead of duplicating
the names or expanding them into the short Physics commands.

That distinction is visible in `\sine[2](\frac{1}{x})`: the long name leaves
literal brackets and ordinary parentheses, while `\sin[2](\frac{1}{x})`
uses Physics `Expression` to add an exponent and automatic fences. Bare
long names require no argument. `rank`, the no-argument `NamedFn` source
registration, retains its existing dispatch; unrelated `Expression`,
`OperatorApplication`, and `Macro` registrations are not added here.

The first 580 original references cover every long name in both modes,
including arithmetic, scripts, fonts, pending negation, literal brackets,
and higher-priority paired-delimiter and AMS operator registrations.
420 additional references cover unbraced roots and fractions, signed
scripts, primes, limits, operator spacing, neighboring functions, dots,
positions, font/size/color boundaries, matrices, fences, and internal text
math. Short Physics and existing base functions provide controls.

There are 964 valid original renderings and 36 original error renderings
in the exact set. All 1,000 match. Against baseline
`a828bea8a0ef4fc56aae9d20e28aad363d5b5d5c`, 878 fail and 122 already match.

## Recorded inherited script-argument issue

The full 1,012-input audit also includes 12 cases such as `x^\sine y`.
The original `SubsupItem.checkItem` rejects an unbraced `FnItem` with
“Missing open brace for superscript”; the existing Go script consumer
accepts the function node. `x^\arg y`, `x_\arg y`, `x^\sin y`, and
`x^\rank y` already expose that shared behavior before this registration
change. Braced named-function scripts are supported.

`physics_named_functions_residuals.json` preserves all 12 complete original,
baseline, and candidate outputs, plus eight baseline controls in both
modes. These are a separate script-item ownership follow-up. They are not
counted as passing comparisons, and their original receipts are never
replaced by Go goldens. The generator regenerates their original outputs
as well as the exact set, retaining the recorded baseline and candidate.

A compact D2 witness is `\sine^2(x)+\cosine^2(x)=1`; another is
`\determinant A+\Probability(E)`.

Regenerate original outputs with
`node --jitless testdata/generate_physics_named_functions.cjs PINNED_ASSETS`.
Run `go test ./... -run TestPhysicsNamedFunctionReferences`.
