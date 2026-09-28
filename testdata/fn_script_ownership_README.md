# Function items and script ownership

`fn_script_ownership_mathjax_3_2_2.json` contains 1,204 complete, unmodified
SVG responses from D2's frozen MathJax 3.2.2 component at source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`: 418 valid expressions and 786
original error SVGs. Every expression uses a fresh original VM, `em=16`,
`ex=8`, no font cache, and the recorded display mode. The fixture records
the three verified asset hashes. All 1,204 match; 682 differ from baseline
`89120d7480b50c05245e2efaa0bdb249da7e953d`.

[`SubsupItem.checkItem`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseItems.ts#L208)
accepts a completed MML item or an opening group/fence. A pending `FnItem`
cannot itself serve as an unbraced script. Thus `x^\arg y` reports
“Missing open brace for superscript”, while `x^{\arg y}` remains valid.
Rejection occurs before later malformed input is read. The script consumer
now preserves that distinction rather than accepting the function node
stored inside the pending item.

Two existing command classifications also needed correction:
[`NamedOp`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts#L379)
emits completed MML for `gcd`, while
[`TeXAtom`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts#L683)
emits `FnItem` for `mathop`. Correct classification preserves valid
unbraced `gcd` scripts and rejects unbraced `mathop` scripts. AMS
`operatorname` and `DeclareMathOperator` continue to emit completed MML.

The corpus covers all 29 explicit-ID Physics NamedFn aliases plus existing
base functions, both script positions, braced and signed scripts, primes,
font/style barriers, pending negation/dots/positions, malformed suffixes,
and dynamic operator/paired-delimiter overrides. It also includes all
previously recorded NamedFn script residuals and controls. Their historical
receipt file remains unchanged; the failures are now exact regression
references here.

`fn_script_ownership_residuals.json` separately preserves ten raw inherited
controls with complete original, baseline, and candidate responses. These
cover named-operator spacing beside a prime (also present for lim/max/min/inf
and operatorname), the existing `mathop{\rm ...}` optimized tree shape, and
the existing Physics Expression eager-parenthesis/bracket diagnostic path.
They are not counted as passing and do not weaken any equality comparison.
Their original outputs can be regenerated; their historical Go receipts
are retained.

Regenerate with `python3 testdata/generate_fn_script_ownership.py PINNED_ASSETS NODE`.
The generator verifies hashes and uses bounded subprocess batches, with a
fresh original runtime for each expression. Run
`go test ./... -run TestFunctionScriptOwnershipReferences`.
