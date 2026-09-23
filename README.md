# mathjax-go

`mathjax-go` is a native Go port of the MathJax 3.2.2 TeX-to-SVG pipeline used
by [D2](https://github.com/d2lang/d2). Its production packages do not embed,
load, or execute JavaScript.

The compatibility target is D2's frozen custom component, not current MathJax.
That component enables the `base`, `mathtools`, `ams`, `amscd`, `braket`,
`cancel`, `cases`, `color`, `gensymb`, `mhchem`, and `physics` TeX packages and
uses SVG output with `fontCache: 'none'`.

```go
svg, err := mathjax.Render(`\frac{1}{2}`)
width, height, err := mathjax.Measure(`\frac{1}{2}`)
```

`Render` uses the same conversion settings as D2: display mode, a 16-pixel em,
an 8-pixel ex, and inline SVG glyph paths. `RenderWithOptions` exposes those
settings explicitly. Both functions return the bare `<svg>` element that D2
previously obtained from the inner HTML of MathJax's `<mjx-container>`.
`Measure` preserves D2's pixel sizing contract by rounding the SVG's ex-based
dimensions up with the configured default of 8 pixels per ex.

As in MathJax, malformed or unsupported TeX is normally rendered as an
`merror` SVG. A Go error instead indicates invalid conversion options or an
internal pipeline failure.

## Compatibility tests

The regular suite is pure Go:

```sh
go test ./...
go test -race ./...
go vet ./...
```

An optional test-only differential runner compares the port with D2's
cryptographically pinned MathJax bundle. Node.js and the JavaScript assets are
used only by this oracle and are not linked into the Go library:

```sh
MATHJAX_GO_ORACLE_DIR=/absolute/path/to/d2renderers/d2latex \
MATHJAX_GO_ORACLE_FULL=1 \
go test ./...
```

The directory must contain `mathjax.js`, `polyfills.js`, and `setup.js` with
the exact SHA-256 values recorded in [PROVENANCE.md](PROVENANCE.md).

## License

Apache License 2.0, with identified MIT-derived portions. See
[LICENSE](LICENSE), [LICENSES](LICENSES), [NOTICE](NOTICE), and
[PROVENANCE.md](PROVENANCE.md).

### TeX mu dimension references

The TeX parser converts valid, fully extracted `mu` values with the pinned
`ParseUtil.muReplace/Em` precision before creating MathML. The existing dimension
scanner and direct MathML length conversion are separate and unchanged.

The dimension corpus keeps 86 full SVG/explicit/own-tree records: 68 raw-primary
results (including eight direct-MathML controls), 12 unchanged grammar boundaries,
two CD-height renderer boundaries with exact primary trees, and four raise/lower
consumer controls. The last four retain inherited structure and sign handling;
only their three dimension strings change. Their SVG hashes are candidate
regression captures, not claims of primary renderer parity. The untouched primary
records are retained beside every qualification.

Registered parser fixtures also bind 44 full-token conversion controls, three
large-number formatting cases, and 38 accepted extraction/cursor/error controls.
The extraction records explicitly preserve scanner differences rather than
claiming complete `GetDimen` grammar compatibility. Regenerate the primary files
with `testdata/generate_tex_mu_dimensions.cjs` and
`internal/tex/testdata/generate_mu_dimension.cjs`, passing the pinned asset
directory and output directory/file respectively.


### Math inside text boxes

The supported `text`, text-font, `hbox` and `mbox` commands parse embedded
`$...$` and `\(...\)` math with MathJax's empty inner lexical environment and
shared macro/tag configuration. Literal chunks preserve internal whitespace
and tilde; only the source's four literal escapes and edge whitespace rules
apply. `hbox` and `mbox` retain their level-zero wrapper, including empty input.

`TestInternalTextPinnedReferences` checks 78 complete unmodified-primary SVGs
and explicit/own-property trees, plus 12 unchanged shared-caller controls.
`textup` uses the same normal-font HBox mapping as `textrm`; `textsl` retains
the pinned undefined-command behavior. AMS tag text remains a separately
recorded boundary. The28 actual-method
references also bind delimiter errors, cardinality, child ownership, font and
macro-counter isolation. Two private pre-postfilter script nodes retain the
existing Go `msub` versus primary two-child `msubsup` representation; public
compiler trees have no such normalization.

Regenerate the references with the pinned D2 MathJax3.2.2 assets:

```
node testdata/generate_internal_text.cjs PINNED_ASSETS [EVIDENCE_DIRECTORY]
node internal/tex/testdata/generate_internal_text_method.cjs PINNED_ASSETS
```


`fbox`, `colorbox` and `fcolorbox` use the same internal text/math parser. Content
is parsed before resolving the outer background or frame color, so a color
created by embedded math is available to its box. The existing notation,
5px padding and 2px border are unchanged. A content error prevents the outer
box from being returned; an earlier shared color definition is not rolled back.

`TestFramedInternalTextPinnedReferences` binds 64 complete primary SVGs and
explicit/own-property trees across both modes, including twelve required error
outputs and unchanged `boxed` controls. The two `macro-inside` cases retain
the fixed package set's unsupported `def` error; successful registered macros
are exercised by the direct handler records instead. Twelve direct primary handler records
also bind argument consumption, node ownership, font/macro lifetime and color
lookup timing. Six private inner `mi(x)` nodes carry exactly three true Go
scope markers before compiler cleanup; the direct test checks that exact
triple and count, then compares every remaining explicit attribute and own
property. Public trees are compared without this qualification. Complete raw
map-layer snapshots remain diagnostic evidence, not a claim that the internal
attribute-storage representations are identical.

Regenerate the framed caller references with:

```
node testdata/generate_framed_internal_text.cjs PINNED_ASSETS [EVIDENCE_DIRECTORY]
node internal/tex/testdata/generate_framed_text_method.cjs PINNED_ASSETS
```


### Parenthesized modulo commands

`pod` and `pmod` use the original one-argument macro definitions, including
style-dependent spacing, fixed parentheses, and forwarding through a registered
`pod` override. All52 targeted public SVGs and complete explicit/own-property
trees match the pinned MathJax3.2.2 renderer. Eight surrounding controls also
match; four `bmod` cases remain exact accepted outputs for a separate fix.
Twenty registered-macro cases cover forwarding, argument counts, missing
arguments and fresh parser state. All twenty match the primary, including its
original recursive-macro diagnostic wording.

Regenerate the primary references with:

```
node testdata/generate_pod_commands.cjs PINNED_ASSETS [EVIDENCE_DIRECTORY]
node testdata/generate_pod_macros.cjs PINNED_ASSETS [EVIDENCE_DIRECTORY]
```

### Initial numeric script arguments

The pinned Superscript/Subscript handlers call GetNext once, then insert a space
following an immediate ASCII digit before checking the existing script slot.
Thus `x^12` scripts `1` and leaves `2`; a macro, font declaration, or comment
encountered first retains the ordinary number scanner. The same initial handler
rule applies after a pending prime and before a nested-marker error. This
script-local lookahead uses the primary JavaScript whitespace set, without
changing ordinary row parsing or number tokenization.

The dedicated 120-case matrix binds complete SVG and explicit/own-property
MathML, including 12 registered macro cases. It has 116 raw primary cases and four
exact-SVG prime cases retaining the precise inherited `pseudoscript` metadata
boundary. The ordinary/Unicode scanner and comma-tail cases now use their
unchanged primary references. Another 48 actual registered-handler records and 108 delegated
parser traces bind source insertion, cursor/error order and base ownership.
Four BOM cases match primary; four NEL cases preserve accepted Go outputs while
the pinned primary throws an internal exception, so they are not parity claims.

The two earlier D066 `digit-tail` whole-output qualifications now compare to
their unchanged primary references. Macro-boundary handling also promotes the
registered font-macro pair and the six retained-prefix cursor records; the
remaining D066 whole-output and precise prime-property boundaries stay intact.

### Ordinary numbers

Ordinary numeric tokens use the pinned ASCII scanner: `1,234` has a comma
operator, while `1{,}234` forms one number. Decimal and source-cursor behavior
follows that same pattern. Supported non-ASCII decimal digits use the pinned
range token kind and variant individually, without changing script lookahead.
The dedicated corpus covers 80 public expressions and 15 parser observations;
92 SVGs are primary, with four precise inherited metadata boundaries and
three explicitly unsupported-primary inputs retained as accepted Go controls.
See `testdata/ordinary_number_README.md` for provenance and regeneration.

### Registered macros using infix names

Registered definitions of `over`, `atop`, `above`, `choose`, `brace` and `brack`
use normal macro dispatch before the built-in fraction handler. For example,
`\DeclareMathOperator{\choose}{pick}x\choose y` renders the declared operator.
An empty registered macro also retains the existing pending-prime behavior.

The 82-case fixture covers all six names, public declarations, argument errors,
empty definitions, grouping and aliases that expand into genuine infix commands.
All 82 complete SVGs match pinned MathJax, including the joined infix aliases.
Sixty ordered full attribute/property trees match primary. The other 22 are bound to
complete accepted records or equivalent ordinary controls, including property
insertion order, atop's existing string zero and plain-prime metadata. No fields
are omitted from these comparisons.

Regenerate the unmodified primary references with:

```
node testdata/generate_infix_macro_priority.cjs PINNED_ASSETS [EVIDENCE_DIRECTORY]
```

### Modulo argument and style spacing

The default `mod` macro consumes one argument and uses the original mathchoice
spacing for display, text and both script styles. The44-case pinned corpus
covers unbraced, grouped, nested and missing arguments, script placement, and a
user-defined override. All42 scoped/control SVGs and explicit/own trees match
primary; the two separate `bmod` row-spacing cases remain unchanged.

```
node testdata/generate_mod_command.cjs PINNED_ASSETS [EVIDENCE_DIRECTORY]
```

### Infix fraction row scope

A second supported infix fraction command in the same logical row reports the
incoming command's `AmbiguousUseOf` error. Denominator, style, size, font and
color continuations retain that row; braces, left/right groups, script arguments
and separate math parsers start independent rows. Incoming `above` dimensions
are read before ambiguity is checked, and registered macro dispatch stays first.

The 108-case pinned corpus has 103 complete primary SVG, ordered explicit/own-tree
and structured-error matches. Five exact accepted-output boundaries retain
unsupported `overwithdelims`, one inline nested fraction, and two nested-style
outputs. No changed output is qualified: all 50 original scope fixes, the two
parsed-group diagnostic upgrades and four registered macro alias upgrades match
primary. Six parser re-entry/error-unwind
controls check independent subsequent rows and restored font state.

Only the unchanged inline nested-fraction control `separate-args-inline` permits
two exact whole-SVG states, captured from accepted code on ARM64 and AMD64.
Its complete raw tree, structured error and parser state remain fixed; no other
case gains an alternative SVG and no numerical tolerance is used.

```
node testdata/generate_infix_scope.cjs PINNED_ASSETS EVIDENCE_DIRECTORY
```

The two earlier dimension-parser repeated-`above` cases and six registered-infix
ambiguity cases now use their untouched primary references; all other inherited
qualifications remain unchanged.

### Tokens with multiple text children

The SVG renderer gives each nonempty text child a positionable group when its
parent has multiple children. This preserves each child and its measured
advance instead of painting successive children at the same origin. Single
children, empty text and the separate explicit-font path retain their source
behavior.

Thirty-two constructed-token cases compare complete SVG and all explicit
attributes and own properties before and after rendering, while checking
original text-node identity. They cover `mo`, `mi`, `mn`, `mtext`, `ms` quotes,
empty children, multiple advances, a bold variant and explicit fonts in both
modes, including repeated rendering. This is an internal
renderer contract prerequisite for the separate adjacent-relation filter;
it is not a claim that accepted TeX parsing already produces every input.

Regenerate the unmodified primary references with:

```
node testdata/generate_multi_text_renderer.cjs PINNED_ASSETS [EVIDENCE_DIRECTORY]
```

### Parsed-group end-of-input diagnostics

An unfinished literal group reports `ExtraOpenMissingClose` ("Extra open brace
or missing close brace"). Argument readers retain their separate
`MissingCloseBrace` diagnostic. The 64-case pinned corpus covers ordinary,
nested, fractional, script, font, style, color and size groups, earlier errors,
and valid controls in both modes. Twenty-four parsed-group errors now match
complete primary SVGs, ordered explicit/own trees, errors and source cursors.
Four missing-right controls now match the primary `ExtraLeftMissingRight`
identity and "Extra \left or missing \right" message, complete SVG and ordered
explicit/own tree. All 40 historical controls keep the same source/cursor state;
two hash diagnostic controls now also match their complete primary outputs,
including the quoted diagnostic. The other 34 complete accepted outputs stay
unchanged. Eight argument-reader and two hash cursor differences remain bound.
The untouched primary and historical accepted references are preserved.
No general error or cursor normalization is used.

```
node testdata/generate_group_eof.cjs PINNED_ASSETS EVIDENCE_DIRECTORY
```

### Macro expansion boundaries

Macro substitution preserves a control word at each literal/argument/tail join,
uses the original escaped-hash rules, and counts UTF-16 code units against the
5120-unit expansion buffer. It installs only expansion plus unparsed input and
resets the cursor before checking the macro count. Argument, substitution and
buffer errors retain the original precedence. The same rule covers the existing
parameter-template macro route; environments and paired delimiters retain their
separate implementations. The operator-name token consumer carries rewritten
source and cursor together, preserving its caller and already-emitted nodes.

The 85 complete-output references comprise the original 72 cases and 13 bounded
default-buffer/UTF-16/prime/style/font controls. All 40 changed SVGs match primary;
83 SVGs and 77 full attribute/own-property trees match without qualifications.
The four hash-diagnostic outputs now use their untouched primary outputs; their
historical accepted records remain in the fixture. Two supplemental
newcommand-environment outputs remain exact accepted boundaries. Two vector and
four prime records have raw primary SVGs with precisely retained own-property
differences; the prime boundary
is independently present in the direct accepted-before control. Registered
handlers bind 22 actual entry/return/error source, cursor and count receipts;
28 actual helper references cover separators, escaping, UTF-16 and error order.
Custom primary maxBuffer observations are not presented as a configurable Go API.

Fourteen additional operator-name records preserve the complete primary trees
and exact accepted property arrays: six SVGs match primary, while eight retain
the separately tracked D103 constructor-spacing residual. The repaired macro
composites equal the accepted direct authored route in full SVG and ordered
trees; they are not counted as primary passes. All literal attributes and every
own property remain asserted, including the existing insertion-order boundary.
The child-error control also requires the caller source, cursor and shared state
to remain intact.

The original fixed-D2 package corpus remains distinct from the official-3.2.2
newcommand augmented observations. Generate raw unchanged-primary captures with:

```
node testdata/generate_macro_boundary.cjs PINNED_ASSETS NEWCOMMAND_3_2_2_JS NEW_EVIDENCE_DIRECTORY
```

Responsive display math now applies the final minimum width of the last eligible
constructed top table. Six labelled-equation cases with surrounding material
match complete pinned-primary SVGs. Constructor identity, reverse measurement,
zero-width/inline behavior and the private math alignment policy have focused
controls; existing table layout consumers are unchanged. The fixture notes in
`testdata/responsive_minwidth_README.md` distinguish complete primary outputs
from unchanged accepted controls and unresolved equation/table composites.
