# Binomial commands and macro precedence

The oracle is unmodified D2-pinned MathJax 3.2.2, source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. Both generators hash-check the three
original assets and create a fresh VM per record. Public commands are evaluated
with the unchanged package set and complete SVG renderer. The registered corpus
uses the existing `ams-declare-ops` map and original `Macro`/`BaseMethods.Macro`
handler to install named declarations; it does not claim support for an authored
newcommand package or replace a primary implementation.

```
node testdata/generate_binomial_commands.cjs /path/to/pinned/d2latex [evidence-directory]
node testdata/generate_binomial_macros.cjs /path/to/pinned/d2latex [evidence-directory]
```

The 48 public records cover all three binomial commands, ambient and forced
styles, scripts, tall and nested arguments, surrounding spacing/fences, two
malformed inputs in both modes, literal text, and unchanged fraction controls.
All 48 complete SVG hashes and explicit-attribute trees are primary exact. All
44 unqualified own-property trees are exact too. Four precise preexisting
metadata receipts remain in `binomial_commands_boundaries.json`: two genfrac
rows omit open/close/texClass (entire accepted tree unchanged); two prime
operators omit primary's own pseudoscript=false (the complete prime subtree is
accepted-before exact despite repaired binomial ancestors). No SVG or explicit
attribute qualification is used. Each recorded path/key/value adjustment must
reconstruct the full expected tree, with every other own property retained.

Accepted parent `46e9024262e80131193abaa119a965789ec1c9ac` fails the 30 affected
public records. The other 18 complete SVGs and raw compiler trees are unchanged.
The six formerly qualified binomial records in the inherited D063 corpus now
require untouched raw-primary SVG and full primary trees; its other 42 actual
outputs remain unchanged, including the eight exact metadata qualifications.

The 24 registered records compare complete SVG and all explicit/own tree fields.
Exact-name overrides of binom/dbinom/tbinom continue to win; overriding binom
cannot redirect the independently registered dbinom/tbinom handlers; overriding
genfrac cannot intercept any built-in binomial; argument and nested macro
expansion still use the normal path. Every conversion has fresh declaration
state. The Go test uses that existing state table and the same successful
compiler finalization operations without adding a production registration API.
Six direct-handler checks require the whole fixed-fence expression to be inside
the optional style owner, preserve argument attributes/parent links, retain
string zero and withDelims, and defer fixed-palette selection to inherited style.

Production scope is only three incorrect built-in macro entries and the existing
binomial handler. It uses the accepted fixed-fence palette and ORD row contract.
Generic genfrac, general fences, renderer, prime collection, error timing, other
macro registrations and the original macro-first dispatch order are unchanged.
