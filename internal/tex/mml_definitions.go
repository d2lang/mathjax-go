// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/core/MmlTree/{MML,MmlNode,MmlFactory}.ts and all 30
// ts/core/MmlTree/MmlNodes/*.ts definitions.

package tex

import (
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/ordered"
)

// mmlUnboundedArity is the integer representation used by the Go tree for
// MathJax's JavaScript Infinity arity.  Negative one remains reserved for the
// inferred-mrow content model.
const mmlUnboundedArity = int(^uint(0) >> 1)

// texMMLFactory is the complete MML registry loaded by MathJax 3.2.2's
// ts/core/MmlTree/MML.ts.  Definitions are immutable after initialization;
// Factory.Create and NewAttributes clone their ordered maps, so the registry
// is safe to share across concurrent Compiler calls.
var texMMLFactory = newTeXMMLFactory()

func propertyMap(entries ...any) *ordered.Map[mml.Property] {
	result := ordered.New[mml.Property]()
	for i := 0; i < len(entries); i += 2 {
		value := entries[i+1]
		if value == "_inherit_" {
			value = mml.Inherit
		}
		result.Set(entries[i].(string), value)
	}
	return result
}

func extendDefaults(base *ordered.Map[mml.Property], entries ...any) *ordered.Map[mml.Property] {
	result := base.Clone()
	for i := 0; i < len(entries); i += 2 {
		value := entries[i+1]
		if value == "_inherit_" {
			value = mml.Inherit
		}
		result.Set(entries[i].(string), value)
	}
	return result
}

func newTeXMMLFactory() *mml.Factory {
	factory := mml.NewFactory()
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
	mathDefaults.Range(func(name string, value mml.Property) bool {
		factory.Globals().Set(name, value)
		return true
	})

	register := func(kind string, defaults *ordered.Map[mml.Property], flags mml.Flags) {
		factory.Register(kind, mml.Definition{Defaults: defaults, Flags: flags})
	}
	plain := mml.Flags{Arity: mmlUnboundedArity}
	tokenFlags := mml.Flags{Token: true, Arity: mmlUnboundedArity}

	register("math", mathDefaults, mml.Flags{Arity: -1, LinebreakContainer: true})
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
	), mml.Flags{Token: true, Embellished: true, Arity: mmlUnboundedArity})
	register("mtext", tokenDefaults, mml.Flags{Token: true, Spacelike: true, Arity: mmlUnboundedArity})
	register("mspace", extendDefaults(tokenDefaults,
		"width", "0em", "height", "0ex", "depth", "0ex", "linebreak", "auto",
	), mml.Flags{Token: true, Spacelike: true, Arity: 0})
	register("ms", extendDefaults(tokenDefaults, "lquote", "\"", "rquote", "\""), tokenFlags)

	register("mrow", base, plain)
	register("inferredMrow", base, mml.Flags{Arity: mmlUnboundedArity, Spacelike: true, Inferred: true, NotParent: true})
	register("mfrac", extendDefaults(base,
		"linethickness", "medium", "numalign", "center", "denomalign", "center", "bevelled", false,
	), mml.Flags{Arity: 2, LinebreakContainer: true})
	register("msqrt", base, mml.Flags{Arity: -1, LinebreakContainer: true})
	register("mroot", base, mml.Flags{Arity: 2})
	register("mstyle", extendDefaults(base,
		"scriptlevel", "_inherit_", "displaystyle", "_inherit_",
		// V8's 1 / Math.sqrt(2) is one ULP below Go's 1 / math.Sqrt2
		// constant.  The value is observable in MathJax's default layer.
		"scriptsizemultiplier", 0.7071067811865475, "scriptminsize", "8px",
		"mathbackground", "_inherit_", "mathcolor", "_inherit_", "dir", "_inherit_",
		"infixlinebreakstyle", "before",
	), mml.Flags{Arity: -1})
	register("merror", base, mml.Flags{Arity: -1, LinebreakContainer: true})
	register("mpadded", extendDefaults(base,
		"width", "", "height", "", "depth", "", "lspace", 0, "voffset", 0,
	), mml.Flags{Arity: -1})
	register("mphantom", base, mml.Flags{Arity: -1})
	register("mfenced", extendDefaults(base, "open", "(", "close", ")", "separators", ","), plain)
	register("menclose", extendDefaults(base, "notation", "longdiv"), mml.Flags{Arity: -1})
	register("maction", extendDefaults(base, "actiontype", "toggle", "selection", 1), mml.Flags{Arity: 1})

	scriptDefaults := extendDefaults(base, "subscriptshift", "", "superscriptshift", "")
	register("msub", scriptDefaults, mml.Flags{Arity: 2})
	register("msup", scriptDefaults, mml.Flags{Arity: 2})
	register("msubsup", scriptDefaults, mml.Flags{Arity: 3})
	underOverDefaults := extendDefaults(base, "accent", false, "accentunder", false, "align", "center")
	register("munder", underOverDefaults, mml.Flags{Arity: 2, LinebreakContainer: true})
	register("mover", underOverDefaults, mml.Flags{Arity: 2, LinebreakContainer: true})
	register("munderover", underOverDefaults, mml.Flags{Arity: 3, LinebreakContainer: true})
	register("mmultiscripts", scriptDefaults, mml.Flags{Arity: 1})
	register("mprescripts", base, mml.Flags{Arity: 0})
	register("none", base, mml.Flags{Arity: 0})

	register("mtable", extendDefaults(base,
		"align", "axis", "rowalign", "baseline", "columnalign", "center", "groupalign", "{left}",
		"alignmentscope", true, "columnwidth", "auto", "width", "auto",
		"rowspacing", "1ex", "columnspacing", ".8em", "rowlines", "none", "columnlines", "none",
		"frame", "none", "framespacing", "0.4em 0.5ex", "equalrows", false, "equalcolumns", false,
		"displaystyle", false, "side", "right", "minlabelspacing", "0.8em",
	), mml.Flags{Arity: mmlUnboundedArity, LinebreakContainer: true})
	rowDefaults := extendDefaults(base,
		"rowalign", "_inherit_", "columnalign", "_inherit_", "groupalign", "_inherit_",
	)
	register("mlabeledtr", rowDefaults, mml.Flags{Arity: 1, LinebreakContainer: true})
	register("mtr", rowDefaults, mml.Flags{Arity: mmlUnboundedArity, LinebreakContainer: true})
	register("mtd", extendDefaults(base,
		"rowspan", 1, "columnspan", 1, "rowalign", "_inherit_", "columnalign", "_inherit_", "groupalign", "_inherit_",
	), mml.Flags{Arity: -1, LinebreakContainer: true})
	register("maligngroup", extendDefaults(base, "groupalign", "_inherit_"), mml.Flags{Arity: -1, Spacelike: true})
	register("malignmark", extendDefaults(base, "edge", "left"), mml.Flags{Arity: 0, Spacelike: true})

	register("mglyph", extendDefaults(tokenDefaults,
		"alt", "", "src", "", "index", "", "width", "auto", "height", "auto", "valign", "0em",
	), tokenFlags)
	register("semantics", extendDefaults(base, "definitionUrl", nil, "encoding", nil), mml.Flags{Arity: 1, NotParent: true})
	annotationDefaults := extendDefaults(base,
		"definitionUrl", nil, "encoding", nil, "cd", "mathmlkeys", "name", "", "src", nil,
	)
	register("annotation", annotationDefaults, plain)
	register("annotation-xml", annotationDefaults, plain)
	register("TeXAtom", base, mml.Flags{Arity: -1})
	register("MathChoice", base, mml.Flags{Arity: 4, NotParent: true})
	register("text", ordered.New[mml.Property](), mml.Flags{Arity: 0})
	register("XML", ordered.New[mml.Property](), mml.Flags{Arity: 0})
	return factory
}
