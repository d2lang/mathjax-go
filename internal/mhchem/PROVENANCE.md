# mhchem source provenance

This package is a native Go translation of `mhchemparser` 4.1.1, the exact
runtime selected by the MathJax 3.2.2 package lock.

| Artifact | Frozen source | License / integrity |
| --- | --- | --- |
| Parser and texifier | [`mhchem/mhchemParser`](https://github.com/mhchem/mhchemParser), npm `mhchemparser@4.1.1`, `src/mhchemParser.ts` | Apache-2.0; npm integrity `sha512-R75CUN6O6e1t8bgailrF1qPq+HhVeFTM3XQ0uzI+mXTybmphy3b6h4NbLOYhemViQ3lUs+6CKRkC3Ws1TlYREA==` |
| Compatibility vectors | `mhchemparser@4.1.1/test/test.html` | Apache-2.0; all 117 assertions are reproduced in `mhchem_test.go` |
| Initial Go translation structure | [`Luo-Studio/go-tex`](https://github.com/Luo-Studio/go-tex), commit `60635a3613c51e710fd5cb57e1d7eeb7b3163a1f`, `tex/mhchem` | MIT; copyright 2026 luo-studio and erweixin |

`data/machines.json` and `data/patterns.json` are deterministic serializations
of the state-machine definitions and pattern sources in the unmodified
`mhchemparser@4.1.1` distribution. The Go matcher implements the finite
JavaScript regular-expression surface used by those tables without a runtime
JavaScript engine or an external regular-expression dependency.

Files adapted from the initial Go translation carry both Apache-2.0 and MIT
SPDX identifiers. The MIT notice is preserved in `LICENSE-MIT`; the repository
root `LICENSE` contains Apache License 2.0.
