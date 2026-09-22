// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/input/tex/base/BaseMappings.ts,
// ts/input/tex/ams/AmsMappings.ts, and
// ts/input/tex/braket/BraketMappings.ts.

package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// symbolDef mirrors the CharacterMap entries in BaseMappings.ts and
// AmsMappings.ts.  Maps are lookup-only; source ordering is retained in the
// command dispatcher where precedence can affect parsing.
type symbolDef struct {
	char  string
	kind  string
	class mml.TeXClass
	attrs map[string]any
}

var identifierSymbols = map[string]symbolDef{
	"alpha": {char: "α"}, "beta": {char: "β"}, "gamma": {char: "γ"},
	"delta": {char: "δ"}, "epsilon": {char: "ϵ"}, "zeta": {char: "ζ"},
	"eta": {char: "η"}, "theta": {char: "θ"}, "iota": {char: "ι"},
	"kappa": {char: "κ"}, "lambda": {char: "λ"}, "mu": {char: "μ"},
	"nu": {char: "ν"}, "xi": {char: "ξ"}, "omicron": {char: "ο"},
	"pi": {char: "π"}, "rho": {char: "ρ"}, "sigma": {char: "σ"},
	"tau": {char: "τ"}, "upsilon": {char: "υ"}, "phi": {char: "ϕ"},
	"chi": {char: "χ"}, "psi": {char: "ψ"}, "omega": {char: "ω"},
	"varepsilon": {char: "ε"}, "vartheta": {char: "ϑ"}, "varpi": {char: "ϖ"},
	"varrho": {char: "ϱ"}, "varsigma": {char: "ς"}, "varphi": {char: "φ"},

	"Gamma": {char: "Γ"}, "Delta": {char: "Δ"}, "Theta": {char: "Θ"},
	"Lambda": {char: "Λ"}, "Xi": {char: "Ξ"}, "Pi": {char: "Π"},
	"Sigma": {char: "Σ"}, "Upsilon": {char: "Υ"}, "Phi": {char: "Φ"},
	"Psi": {char: "Ψ"}, "Omega": {char: "Ω"},

	"aleph": {char: "ℵ"}, "beth": {char: "ℶ"}, "gimel": {char: "ℷ"},
	"daleth": {char: "ℸ"}, "hbar": {char: "ℏ"}, "imath": {char: "ı"},
	"jmath": {char: "ȷ"}, "ell": {char: "ℓ"}, "wp": {char: "℘"},
	"Re": {char: "ℜ"}, "Im": {char: "ℑ"}, "partial": {char: "∂"},
	"infty": {char: "∞"}, "emptyset": {char: "∅"}, "varnothing": {char: "∅"},
	"nabla": {char: "∇"}, "top": {char: "⊤"}, "bot": {char: "⊥"},
	"angle": {char: "∠"}, "measuredangle": {char: "∡"}, "sphericalangle": {char: "∢"},
	"triangle": {char: "△"}, "forall": {char: "∀"}, "exists": {char: "∃"},
	"nexists": {char: "∄"}, "neg": {char: "¬"}, "lnot": {char: "¬"},
	"flat": {char: "♭"}, "natural": {char: "♮"}, "sharp": {char: "♯"},
	"clubsuit": {char: "♣"}, "diamondsuit": {char: "♢"},
	"heartsuit": {char: "♡"}, "spadesuit": {char: "♠"},
}

func op(char string, class mml.TeXClass) symbolDef {
	return symbolDef{char: char, kind: "mo", class: class}
}

func bigOp(char string, movable bool) symbolDef {
	return symbolDef{char: char, kind: "mo", class: mml.TeXClassOp,
		attrs: map[string]any{"largeop": true, "movablelimits": movable, "movesupsub": true}}
}

var operatorSymbols = map[string]symbolDef{
	"sum": bigOp("∑", true), "prod": bigOp("∏", true), "coprod": bigOp("∐", true),
	"bigcup": bigOp("⋃", true), "bigcap": bigOp("⋂", true),
	"bigvee": bigOp("⋁", true), "bigwedge": bigOp("⋀", true),
	"bigoplus": bigOp("⨁", true), "bigotimes": bigOp("⨂", true),
	"bigodot": bigOp("⨀", true), "biguplus": bigOp("⨄", true),
	"int": bigOp("∫", false), "intop": bigOp("∫", true),
	"iint": bigOp("∬", false), "iiint": bigOp("∭", false), "oint": bigOp("∮", false),

	"pm": op("±", mml.TeXClassBin), "mp": op("∓", mml.TeXClassBin),
	"times": op("×", mml.TeXClassBin), "div": op("÷", mml.TeXClassBin),
	"cdot": op("⋅", mml.TeXClassBin), "ast": op("∗", mml.TeXClassBin),
	"star": op("⋆", mml.TeXClassBin), "circ": op("∘", mml.TeXClassBin),
	"bullet": op("∙", mml.TeXClassBin), "diamond": op("⋄", mml.TeXClassBin),
	"cap": op("∩", mml.TeXClassBin), "cup": op("∪", mml.TeXClassBin),
	"sqcap": op("⊓", mml.TeXClassBin), "sqcup": op("⊔", mml.TeXClassBin),
	"uplus": op("⊎", mml.TeXClassBin), "wedge": op("∧", mml.TeXClassBin),
	"land": op("∧", mml.TeXClassBin), "vee": op("∨", mml.TeXClassBin),
	"lor": op("∨", mml.TeXClassBin), "oplus": op("⊕", mml.TeXClassBin),
	"ominus": op("⊖", mml.TeXClassBin), "otimes": op("⊗", mml.TeXClassBin),
	"oslash": op("⊘", mml.TeXClassBin), "odot": op("⊙", mml.TeXClassBin),
	"setminus": op("∖", mml.TeXClassBin), "backslash": op("∖", mml.TeXClassOrd),

	"le": op("≤", mml.TeXClassRel), "leq": op("≤", mml.TeXClassRel),
	"ge": op("≥", mml.TeXClassRel), "geq": op("≥", mml.TeXClassRel),
	"ne": op("≠", mml.TeXClassRel), "neq": op("≠", mml.TeXClassRel),
	"equiv": op("≡", mml.TeXClassRel), "sim": op("∼", mml.TeXClassRel),
	"simeq": op("≃", mml.TeXClassRel), "approx": op("≈", mml.TeXClassRel),
	"cong": op("≅", mml.TeXClassRel), "propto": op("∝", mml.TeXClassRel),
	"in": op("∈", mml.TeXClassRel), "notin": op("∉", mml.TeXClassRel),
	"ni": op("∋", mml.TeXClassRel), "owns": op("∋", mml.TeXClassRel),
	"subset": op("⊂", mml.TeXClassRel), "supset": op("⊃", mml.TeXClassRel),
	"subseteq": op("⊆", mml.TeXClassRel), "supseteq": op("⊇", mml.TeXClassRel),
	"prec": op("≺", mml.TeXClassRel), "succ": op("≻", mml.TeXClassRel),
	"preceq": op("⪯", mml.TeXClassRel), "succeq": op("⪰", mml.TeXClassRel),
	"parallel": op("∥", mml.TeXClassRel), "mid": op("∣", mml.TeXClassRel),
	"perp": op("⊥", mml.TeXClassRel), "models": op("⊨", mml.TeXClassRel),
	"vdash": op("⊢", mml.TeXClassRel), "dashv": op("⊣", mml.TeXClassRel),

	"leftarrow": op("←", mml.TeXClassRel), "gets": op("←", mml.TeXClassRel),
	"rightarrow": op("→", mml.TeXClassRel), "to": op("→", mml.TeXClassRel),
	"leftrightarrow": op("↔", mml.TeXClassRel), "mapsto": op("↦", mml.TeXClassRel),
	"Leftarrow": op("⇐", mml.TeXClassRel), "Rightarrow": op("⇒", mml.TeXClassRel),
	"Leftrightarrow": op("⇔", mml.TeXClassRel), "iff": op("⟺", mml.TeXClassRel),
	"longleftarrow": op("⟵", mml.TeXClassRel), "longrightarrow": op("⟶", mml.TeXClassRel),
	"longleftrightarrow": op("⟷", mml.TeXClassRel),
	"uparrow":            op("↑", mml.TeXClassRel), "downarrow": op("↓", mml.TeXClassRel),
	"updownarrow": op("↕", mml.TeXClassRel),

	"ldots": op("…", mml.TeXClassInner), "cdots": op("⋯", mml.TeXClassInner),
	"vdots": op("⋮", mml.TeXClassOrd), "ddots": op("⋱", mml.TeXClassOrd),
	"colon": op(":", mml.TeXClassPunct),
}

var delimiterSymbols = map[string]string{
	"(": "(", ")": ")", "[": "[", "]": "]", "{": "{", "}": "}", ".": "",
	"lparen": "(", "rparen": ")", "lbrack": "[", "rbrack": "]",
	"lbrace": "{", "rbrace": "}",
	"vert": "|", "lvert": "|", "rvert": "|", "|": "‖",
	"Vert": "‖", "lVert": "‖", "rVert": "‖",
	"langle": "⟨", "rangle": "⟩", "lceil": "⌈", "rceil": "⌉",
	"lfloor": "⌊", "rfloor": "⌋", "backslash": "∖",
	"uparrow": "↑", "downarrow": "↓", "updownarrow": "↕",
	"Uparrow": "⇑", "Downarrow": "⇓", "Updownarrow": "⇕",
}

var functionNames = map[string]string{
	"arccos": "arccos", "arcsin": "arcsin", "arctan": "arctan", "arg": "arg",
	"cos": "cos", "cosh": "cosh", "cot": "cot", "coth": "coth", "csc": "csc",
	"deg": "deg", "det": "det", "dim": "dim", "exp": "exp", "gcd": "gcd",
	"hom": "hom", "inf": "inf", "ker": "ker", "lg": "lg", "lim": "lim",
	"liminf": "lim inf", "limsup": "lim sup", "ln": "ln", "log": "log",
	"max": "max", "min": "min", "Pr": "Pr", "sec": "sec", "sin": "sin",
	"sinh": "sinh", "sup": "sup", "tan": "tan", "tanh": "tanh",
	"injlim": "inj lim", "projlim": "proj lim", "varliminf": "lim inf", "varlimsup": "lim sup",
}

var simpleMacros = map[string]string{
	"stackrel":  "\\mathrel{\\mathop{#2}\\limits^{#1}}",
	"dfrac":     "\\displaystyle\\frac{#1}{#2}",
	"tfrac":     "\\textstyle\\frac{#1}{#2}",
	"binom":     "\\left(\\frac{#1}{#2}\\right)",
	"dbinom":    "\\displaystyle\\binom{#1}{#2}",
	"tbinom":    "\\textstyle\\binom{#1}{#2}",
	"pmod":      "\\pod{{\\rm mod}\\ #1}",
	"mod":       "\\mathop{\\rm mod}\\nolimits\\ ",
	"bmod":      "\\mathbin{\\rm mod}",
	"iff":       "\\;\\Longleftrightarrow\\;",
	"implies":   "\\;\\Longrightarrow\\;",
	"impliedby": "\\;\\Longleftarrow\\;",
	"not":       "\\mathrel{\\rlap{#1}\\mkern2mu{/}}",
	"substack":  "\\begin{subarray}{c}#1\\end{subarray}",
	"Bra":       "\\left\\langle #1\\right|",
	"Ket":       "\\left|#1\\right\\rangle",
	"Braket":    "\\left\\langle #1\\right\\rangle",
	"Set":       "\\left\\{#1\\right\\}",
}
