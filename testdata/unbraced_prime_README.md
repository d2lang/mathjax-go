# Pending script arguments and prime events

These references use the unchanged D2-pinned MathJax 3.2.2 assets at
`ad8f5c21cb810236551da8c6512ba733e67357ee`. The generator checks all three asset
hashes and captures complete SVG plus the explicit-attribute/own-property tree.
The 120 rows contain 100 public fixed-package inputs and 20 separately labelled
registered-macro inputs. The latter are actual `Macro` registrations, not
additional public TeX commands.

The script reader keeps a Subsup argument pending across commands that emit no
item, including macro expansion and font declarations. It applies the existing
attachment ownership/occupied-slot decision before reading the argument. An
unbraced prime runs the primary Prime checks and collector, then receives the
pending Subsup item's typed error. Actual grouped MathML and left/right groups
still supply arguments. Font changes follow the containing parse-row lifetime.
Other commands' GetArgument routes and generic macro concatenation are unchanged.

The companion private fixture calls the actual pinned Superscript, Subscript,
Prime and SubsupItem methods for 264 constructed records. It binds the prepared
family, selected slot, wrapper reuse, original child identities, explicit/own
maps, error stage, typed error and UTF-16 cursor. The read-only Go decision does
not reparent an input before a real argument exists. Existing attachment fill
tests continue to bind successful ownership and the eager-subtype distinction.
The 62 captured whole-input error records also bind error ID/message and suffix;
56 retain exact primary source/cursor coordinates. Six registered expansions
retain the existing Go consumed-prefix representation and explicitly bind that
prefix instead of claiming identical absolute primary positions.

Of 120 outputs, 114 complete SVGs and explicit trees match raw primary; 96 also
match every own property. Eighteen rows retain only precisely named inherited
`pseudoscript` boolean omissions at the recorded node paths. Sixteen retain the
same accepted-before path; two registered grouped-prime rows move the unchanged
prime token into the corrected argument, with the earlier token path recorded.
Every other own property and the complete resulting tree remain strict.

Six rows retain exact accepted-before output/tree boundaries: protected
`operatorname` curly-prime handling (two modes), a registered font macro's
existing control-word concatenation (two), and unbraced numeric `x^12` (two).
The last is a separate numeric-argument observation; no numeric scanner changes
are part of this fix. `unbraced_prime_boundaries.json` binds these to accepted
PR36 merge `e93f5eb32042cd5a772f5b8fa3c6aabe540701da` with complete saved trees
and hashes. No new output qualification is used for a changed-but-wrong result.

For percentage-width table output, the complete SVG remains strict and the
fixture explicitly records that the existing ex-only public `Measure` contract
returns `SVG dimensions not found`; no synthetic pixel width is substituted.

Regenerate the two fixture sets with normal Node and the pinned assets:

```
node testdata/generate_unbraced_prime.cjs PINNED_ASSETS OUTPUT_DIRECTORY
node internal/tex/testdata/generate_script_argument_items.cjs PINNED_ASSETS OUTPUT_JSON
```

The final candidate inherits the accepted prime collector whitespace policy;
initial prime dispatch, prime grouping and other argument readers are separate.
