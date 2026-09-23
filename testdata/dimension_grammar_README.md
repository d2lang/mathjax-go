# TeX dimension grammar

The primary is the unchanged D2 MathJax 3.2.2 bundle at source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. The public generator observes the
complete SVG and explicit/own-property tree. The private generator calls the
actual `TexParser.GetDimen`, `GetArgument`, `GetNext`, `ParseUtil.matchDimen`,
and `BaseMethods.Over` from the pinned bundle; it does not replace them.

D079 changes supported dimension consumers to the source grammar: decimal dot
or comma, all nine units, local ECMAScript whitespace, a complete braced value
or an unbraced prefix with at most one following ASCII space. It preserves
D077's mu conversion. `above` now reads the dimension before its existing
fraction construction, and spreadlines uses the invoking `begin` error label.
D070 now rejects repeated valid infix fractions; registered macro priority is
covered separately.
The generic argument reader and global whitespace scanner are unchanged.

The 60 public cases have 58 raw-primary SVGs and 56 raw-primary complete
explicit/own trees. The two `raise-valid-boundary` cases have primary SVGs and
only the inherited `voffset` spelling `1.5pt` versus primary `+1.5pt`; the test
binds that exact attribute on the exact primary tree and compares every other
field. Two unchanged `rule` cases retain complete accepted-parent SVG/tree receipts.
The two repeated `above` cases use their unchanged primary references after the
D070 row-scope correction. The unsupported `vspace` and `raisebox` commands now
use the same undefined-command errors as pinned MathJax, before their arguments
are read. Their unused handlers and legacy dimension reader are removed;
supported `raise`, `lower` and dimension consumers retain the strict reader.
Twelve extra public controls require complete primary SVGs and trees for valid
and malformed unsupported-command inputs and four supported malformed arguments.
The old alias output fields remain historical receipts, not expected live output.
Registered macro dispatch still precedes the built-in command switch.

The registered fixture contains 106 exact value/error/cursor/suffix calls,
four invoking-control-sequence label cases, and seven primary Over push
observations. A thrown primary call has no returned value (captured null); the
Go error return must have an empty string and the exact error ID/message.
Physical UTF-8 consumption and error-only cursor overrun are separate. A final
ASCII escape inside an unclosed brace advances the primary cursor one past EOF;
all eleven supported public consumer routes must return the error immediately,
and compilation must preserve it without attempting an out-of-bounds slice.

D077's original fixtures and historical boundary receipts remain unchanged.
Its twelve scanner-qualified public rows now require raw primary SVG/trees;
the other 74 outputs stay unchanged. Its 38 private extraction assertions now
use that same fixture's original actual GetDimen observations, while the old
accepted extraction records remain available. Raise/lower and CD-height
qualifications remain exactly as before. Direct MathML length behavior is not
changed by this TeX grammar correction.
