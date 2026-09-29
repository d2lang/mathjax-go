// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/core/MmlTree/{MML,MmlNode,MmlFactory}.ts and all 30
// ts/core/MmlTree/MmlNodes/*.ts definitions.

package mml

import (
	"github.com/d2lang/mathjax-go/internal/ordered"
)

// UnboundedArity is the integer representation used by the Go tree for
// MathJax's JavaScript Infinity arity.  Negative one remains reserved for the
// inferred-mrow content model.
const UnboundedArity = int(^uint(0) >> 1)

func propertyMap(entries ...any) *ordered.Map[Property] {
	result := ordered.New[Property]()
	for i := 0; i < len(entries); i += 2 {
		value := entries[i+1]
		if value == "_inherit_" {
			value = Inherit
		}
		result.Set(entries[i].(string), value)
	}
	return result
}

func extendDefaults(base *ordered.Map[Property], entries ...any) *ordered.Map[Property] {
	result := base.Clone()
	for i := 0; i < len(entries); i += 2 {
		value := entries[i+1]
		if value == "_inherit_" {
			value = Inherit
		}
		result.Set(entries[i].(string), value)
	}
	return result
}

// NewMathJaxFactory returns the complete pinned MathJax 3.2.2 registry.
// Each factory owns immutable definitions; Create clones their attribute maps.
func NewMathJaxFactory() *Factory {
	factory := NewFactory()
	base := propertyMap(
		"mathbackground", "_inherit_",
		"mathcolor", "_inherit_",
		"mathsize", "_inherit_",
		"dir", "_inherit_",
	)
	tokenDefaults := extendDefaults(base, "mathvariant", "normal", "mathsize", "_inherit_")
	mathDefaults := extendDefaults(base,
		"mathvariant", "normal",
		"mathsize", "normal",
		"mathcolor", "",
		"mathbackground", "transparent",
		"dir", "ltr",
		"scriptlevel", 0,
		"displaystyle", false,
		"display", "inline",
		"maxwidth", "",
		"overflow", "linebreak",
		"altimg", "",
		"altimg-width", "",
		"altimg-height", "",
		"altimg-valign", "",
		"alttext", "",
		"cdgroup", "",
		"scriptsizemultiplier", 0.7071067811865475,
		"scriptminsize", "8px",
		"infixlinebreakstyle", "before",
		"lineleading", "1ex",
		"linebreakmultchar", "\u2062",
		"indentshift", "auto",
		"indentalign", "auto",
		"indenttarget", "",
		"indentalignfirst", "indentalign",
		"indentshiftfirst", "indentshift",
		"indentalignlast", "indentalign",
		"indentshiftlast", "indentshift",
	)
	mathDefaults.Range(func(name string, value Property) bool {
		factory.Globals().Set(name, value)
		return true
	})

	register := func(kind string, defaults *ordered.Map[Property], flags Flags) {
		factory.Register(kind, Definition{Defaults: defaults, Flags: flags})
	}
	plain := Flags{Arity: UnboundedArity}
	tokenFlags := Flags{Token: true, Arity: UnboundedArity}

	register("math", mathDefaults, Flags{Arity: -1, LinebreakContainer: true})
	register("mi", tokenDefaults, tokenFlags)
	register("mn", tokenDefaults, tokenFlags)
	register("mo", extendDefaults(tokenDefaults,
		"form", "infix", "fence", false, "separator", false,
		"lspace", "thickmathspace", "rspace", "thickmathspace",
		"stretchy", false, "symmetric", false,
		"maxsize", "infinity", "minsize", "0em",
		"largeop", false, "movablelimits", false, "accent", false,
		"linebreak", "auto", "lineleading", "1ex", "linebreakstyle", "before",
		"indentalign", "auto", "indentshift", "0", "indenttarget", "",
		"indentalignfirst", "indentalign", "indentshiftfirst", "indentshift",
		"indentalignlast", "indentalign", "indentshiftlast", "indentshift",
	), Flags{Token: true, Embellished: true, Arity: UnboundedArity})
	register("mtext", tokenDefaults, Flags{Token: true, Spacelike: true, Arity: UnboundedArity})
	register("mspace", extendDefaults(tokenDefaults,
		"width", "0em", "height", "0ex", "depth", "0ex", "linebreak", "auto",
	), Flags{Token: true, Spacelike: true, Arity: 0})
	register("ms", extendDefaults(tokenDefaults, "lquote", "\"", "rquote", "\""), tokenFlags)

	register("mrow", base, plain)
	register("inferredMrow", base, Flags{Arity: UnboundedArity, Spacelike: true, Inferred: true, NotParent: true})
	register("mfrac", extendDefaults(base,
		"linethickness", "medium", "numalign", "center", "denomalign", "center", "bevelled", false,
	), Flags{Arity: 2, LinebreakContainer: true})
	register("msqrt", base, Flags{Arity: -1, LinebreakContainer: true})
	register("mroot", base, Flags{Arity: 2})
	register("mstyle", extendDefaults(base,
		"scriptlevel", "_inherit_", "displaystyle", "_inherit_",
		// V8's 1 / Math.sqrt(2) is one ULP below Go's 1 / math.Sqrt2
		// constant.  The value is observable in MathJax's default layer.
		"scriptsizemultiplier", 0.7071067811865475, "scriptminsize", "8px",
		"mathbackground", "_inherit_", "mathcolor", "_inherit_", "dir", "_inherit_",
		"infixlinebreakstyle", "before",
	), Flags{Arity: -1})
	register("merror", base, Flags{Arity: -1, LinebreakContainer: true})
	register("mpadded", extendDefaults(base,
		"width", "", "height", "", "depth", "", "lspace", 0, "voffset", 0,
	), Flags{Arity: -1})
	register("mphantom", base, Flags{Arity: -1})
	register("mfenced", extendDefaults(base, "open", "(", "close", ")", "separators", ","), plain)
	register("menclose", extendDefaults(base, "notation", "longdiv"), Flags{Arity: -1})
	register("maction", extendDefaults(base, "actiontype", "toggle", "selection", 1), Flags{Arity: 1})

	scriptDefaults := extendDefaults(base, "subscriptshift", "", "superscriptshift", "")
	register("msub", scriptDefaults, Flags{Arity: 2})
	register("msup", scriptDefaults, Flags{Arity: 2})
	register("msubsup", scriptDefaults, Flags{Arity: 3})
	underOverDefaults := extendDefaults(base, "accent", false, "accentunder", false, "align", "center")
	register("munder", underOverDefaults, Flags{Arity: 2, LinebreakContainer: true})
	register("mover", underOverDefaults, Flags{Arity: 2, LinebreakContainer: true})
	register("munderover", underOverDefaults, Flags{Arity: 3, LinebreakContainer: true})
	register("mmultiscripts", scriptDefaults, Flags{Arity: 1})
	register("mprescripts", base, Flags{Arity: 0})
	register("none", base, Flags{Arity: 0})

	register("mtable", extendDefaults(base,
		"align", "axis", "rowalign", "baseline", "columnalign", "center", "groupalign", "{left}",
		"alignmentscope", true, "columnwidth", "auto", "width", "auto",
		"rowspacing", "1ex", "columnspacing", ".8em", "rowlines", "none", "columnlines", "none",
		"frame", "none", "framespacing", "0.4em 0.5ex", "equalrows", false, "equalcolumns", false,
		"displaystyle", false, "side", "right", "minlabelspacing", "0.8em",
	), Flags{Arity: UnboundedArity, LinebreakContainer: true})
	rowDefaults := extendDefaults(base,
		"rowalign", "_inherit_", "columnalign", "_inherit_", "groupalign", "_inherit_",
	)
	register("mlabeledtr", rowDefaults, Flags{Arity: 1, LinebreakContainer: true})
	register("mtr", rowDefaults, Flags{Arity: UnboundedArity, LinebreakContainer: true})
	register("mtd", extendDefaults(base,
		"rowspan", 1, "columnspan", 1, "rowalign", "_inherit_", "columnalign", "_inherit_", "groupalign", "_inherit_",
	), Flags{Arity: -1, LinebreakContainer: true})
	register("maligngroup", extendDefaults(base, "groupalign", "_inherit_"), Flags{Arity: -1, Spacelike: true})
	register("malignmark", extendDefaults(base, "edge", "left"), Flags{Arity: 0, Spacelike: true})

	register("mglyph", extendDefaults(tokenDefaults,
		"alt", "", "src", "", "index", "", "width", "auto", "height", "auto", "valign", "0em",
	), tokenFlags)
	register("semantics", extendDefaults(base, "definitionUrl", nil, "encoding", nil), Flags{Arity: 1, NotParent: true})
	annotationDefaults := extendDefaults(base,
		"definitionUrl", nil, "encoding", nil, "cd", "mathmlkeys", "name", "", "src", nil,
	)
	register("annotation", annotationDefaults, plain)
	register("annotation-xml", annotationDefaults, plain)
	register("TeXAtom", base, Flags{Arity: -1})
	register("MathChoice", base, Flags{Arity: 4, NotParent: true})
	register("text", ordered.New[Property](), Flags{Arity: 0})
	register("XML", ordered.New[Property](), Flags{Arity: 0})
	return factory
}
