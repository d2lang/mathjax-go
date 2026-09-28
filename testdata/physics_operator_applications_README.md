# Physics residue and operator applications

`physics_operator_applications_mathjax_3_2_2.json` contains 2,658 complete
SVG responses from D2's frozen MathJax 3.2.2 component at source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`: 1,878 valid expressions and
780 original error SVGs. Each expression runs in a fresh VM with `em=16`,
`ex=8`, no font cache, and its recorded display mode. All original SVGs
are compared verbatim, without normalization or Go-generated goldens.
The three asset hashes are recorded and checked by the generator.

All 2,658 references match. Against baseline
`3f802e1cd86b638ac11221a2b7ed3ed0fa5fca45`, which includes the preceding
shared function/script ownership repair, 2,308 fail and 350 already match.

## Source registrations and behavior

The retained
[`Physics-expressions-macros` table](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMappings.ts#L133)
defines the ordinary `Residue` macro as `\mathrm{Res}` and six
`OperatorApplication` registrations. They were previously unimplemented;
`Re` and `Im` incorrectly fell through to the older blackletter symbols.

| Commands | Parsed operator expression | Automatically enlarged opening fences |
| --- | --- | --- |
| `Res` | `\Residue` | `(`, `[`, `{` |
| `principalvalue`, `pv` | `{\cal P}` | none |
| `PV` | `{\rm P.V.}` | none |
| `Re`, `Im` | `\mathrm{Re}`, `\mathrm{Im}` | `{` |

[`vectorApplication`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMethods.ts#L305)
parses the operator through a fresh child parser and pushes a pending
function item before inspecting the caller's next token. The Go handler
preserves that order through an after-node action. For example,
`x^\Res{` rejects the unbraced function script before scanning the missing
closing brace. An error inside an overridden `Residue` operator still
occurs first, in its genuine child parser.

An operand is optional. A braced argument is consumed and reinserted with
surrounding spaces into the caller's source; only allowed braces receive
generated `\left\{` / `\right\}` delimiters. Principal-value braces
therefore disappear as a lexical group, allowing fonts and argument-taking
macros to continue into the caller. For instance, `\pv{\frac}xy` is valid.
This is not equivalent to parsing every argument in a separate group.

Allowed raw parentheses/brackets use
[`AutoOpen`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsItems.ts#L32)
with the configured pair. The existing derivative item now carries that
pair while retaining its default parentheses. Matching raw closing events,
nested opening-token counts, and stack barriers remain observable. An
unallowed fence, including a vertical bar, remains ordinary following input.

Dynamic paired-delimiter and AMS operator registrations retain their
existing precedence. The corpus covers both definition orders, overrides
inside the parsed operator expression, all fence choices, empty and missing
arguments, ECMAScript whitespace/comments, nested roots/fractions/matrices,
font/color/style continuation, functions/negation/dots/positions, scripts,
primes, infix fractions, and error precedence across child parsers.

## Preserved inherited controls

The full audit contains 2,836 comparisons before duplicate inputs are
removed. `physics_operator_applications_residuals.json` separately retains
136 complete original/baseline/candidate receipts: 72 existing vector-helper
controls, 12 calligraphic-P script rounding cases, eight original JavaScript
exceptions on NEL, 14 existing misplaced-`cr` diagnostics, 14 original error
SVGs with an unescaped ampersand attribute, and 16 existing primitive controls.
The vector controls include 20 already exact cases. The other vector cases
still expose the old eager argument/parenthesis helper. This change does not
rewrite gradient, divergence, curl, or laplacian.

The calligraphic rounding difference is also present with plain
`\mathcal{P}^2` and `{\cal P}'`; the same misplaced-`cr`, ampersand,
and NEL behavior occurs without these new commands. These receipts are
not passing references or weakened comparisons. Their original outputs
regenerate independently; historical Go receipts remain unchanged.

Regenerate with
`python3 testdata/generate_physics_operator_applications.py PINNED_ASSETS NODE`.
Run `go test ./... -run TestPhysicsOperatorApplicationReferences`.

Compact visual examples are `\Res[\frac{1}{z-a}]`,
`\Re{1+i}+\Im{1+i}`, and
`\pv\int_{-\infty}^{\infty}\frac{f(x)}{x-a}\,dx`.
