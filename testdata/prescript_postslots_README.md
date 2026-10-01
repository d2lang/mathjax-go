# Authored right scripts after Mathtools prescript

`prescript_postslots_mathjax_3_2_2.json.gz` retains 1,170 complete, unchanged
original APIs containing SVG: 770 valid renderings and 400 TeX error SVGs.
The separate runtime file retains both complete original exceptions for an
incomplete converted limits family. Those two inputs have no SVG and receive
no parity credit; Go returns a bounded conversion error rather than panicking.

The frozen source creates `mmultiscripts` with absent post-subscript and
post-superscript slots, followed by `mprescripts` and the pre-script pair.
Because this node derives from `msubsup`, authored right scripts and primes
reuse its post slots. Mathtools' final `fixPrescripts` filter fills absent slots
with `none`, removes the post pair only when both were absent, and normalizes
empty pre-script rows. Previously Go normalized the post pair eagerly and
wrapped the whole prescript expression in another script node. Tall or wide
left scripts consequently moved the right scripts away from the base.

The fix retains the source's open slots through parsing, extends the existing
script/prime and Limits family checks to `mmultiscripts`, and runs the source
filter before inheritance. For both-empty prescripts, pushing an inferred
multi-token argument follows the source's flattening behavior, so an authored
right script binds to its last token. No SVG layout code changes.

The inventory covers full/one-sided/empty pre-script pairs, empty-group
arguments, ordinary and operator bases, multi-token and nested bases,
both script orders, primes and prime/script combinations, grouped controls,
missing arguments and repeated-script errors, copied boxes and poor-man's
bold, fonts/colors, nested fractions and arrays, and Limits conversion. On
prerequisite commit `feadb3228901e849dbf0f4daef386a5552dc4ac6`, 336 valid
originals differed and 834 complete originals were already exact. The fixed
renderer matches all 1,170 complete SVG responses byte for byte. The two
suggested D2 witnesses are included in both display modes.

Regenerate from the verified frozen assets:

```sh
python3 testdata/generate_prescript_postslots.py /path/to/assets /path/to/node
```

The generator verifies all three asset hashes, disables JIT, creates a fresh
original VM for every conversion, preserves full runtime stacks separately,
and never invokes Go or filters cases by candidate output.
