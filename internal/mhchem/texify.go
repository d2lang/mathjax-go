// Copyright (c) 2015-2021 Martin Hensel
// Copyright (c) 2026 luo-studio
// Copyright (c) 2026 erweixin (upstream RaTeX)
// SPDX-License-Identifier: Apache-2.0 AND MIT
//
// Go translation and modification of mhchemparser 4.1.1. Initial Go
// translation structure adapted from github.com/Luo-Studio/go-tex (MIT).

package mhchem

import (
	"fmt"
	"strings"
)

// texify renders the parser output (a flat []Value of strings and tagged
// maps) into a TeX fragment. Mirrors upstream texify::go.
func texify(input []Value, addOuterBraces bool) (string, error) {
	var sb strings.Builder
	cee := false
	for _, v := range input {
		if s, ok := v.(string); ok {
			sb.WriteString(s)
			continue
		}
		if typeStr(v) == "1st-level escape" {
			cee = true
		}
		out, err := texify2(v)
		if err != nil {
			return "", err
		}
		sb.WriteString(out)
	}
	res := sb.String()
	if addOuterBraces && !cee && res != "" {
		return "{" + res + "}", nil
	}
	return res, nil
}

func typeStr(v Value) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	s, _ := m["type_"].(string)
	return s
}

func mapStr(v Value, key string) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	s, _ := m[key].(string)
	return s
}

func arrInner(v Value, key string) (string, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return "", nil
	}
	raw, ok := m[key]
	if !ok {
		return "", nil
	}
	arr, ok := raw.([]Value)
	if !ok {
		// May come in as []any after json round-trip; coerce.
		if as, ok2 := raw.([]any); ok2 {
			cv := make([]Value, len(as))
			for i := range as {
				cv[i] = as[i]
			}
			arr = cv
		} else {
			return "", nil
		}
	}
	return texify(arr, false)
}

func texify2(buf Value) (string, error) {
	t := typeStr(buf)
	if t == "" {
		return "", parserError("MhchemBugT", "mhchem bug T. Please report.")
	}
	switch t {
	case "chemfive":
		return texifyChemfive(buf)
	case "rm":
		return fmt.Sprintf("\\mathrm{%s}", mapStr(buf, "p1")), nil
	case "text":
		p1 := mapStr(buf, "p1")
		if strings.ContainsAny(p1, "^_") {
			p1 = strings.Replace(p1, " ", "~", 1)
			p1 = strings.Replace(p1, "-", "\\text{-}", 1)
			return fmt.Sprintf("\\mathrm{%s}", p1), nil
		}
		return fmt.Sprintf("\\text{%s}", p1), nil
	case "roman numeral":
		return fmt.Sprintf("\\mathrm{%s}", mapStr(buf, "p1")), nil
	case "state of aggregation":
		i, err := arrInner(buf, "p1")
		if err != nil {
			return "", err
		}
		return "\\mskip2mu " + i, nil
	case "state of aggregation subscript":
		i, err := arrInner(buf, "p1")
		if err != nil {
			return "", err
		}
		return "\\mskip1mu " + i, nil
	case "bond":
		k := mapStr(buf, "kind_")
		out, ok := getBond(k)
		if !ok {
			return "", parserError("MhchemErrorBond", "mhchem Error. Unknown bond type ("+k+")")
		}
		return out, nil
	case "frac":
		p1 := mapStr(buf, "p1")
		p2 := mapStr(buf, "p2")
		c := fmt.Sprintf("\\frac{%s}{%s}", p1, p2)
		return fmt.Sprintf("\\mathchoice{\\textstyle%s}{%s}{%s}{%s}", c, c, c, c), nil
	case "pu-frac":
		p1, err := arrInner(buf, "p1")
		if err != nil {
			return "", err
		}
		p2, err := arrInner(buf, "p2")
		if err != nil {
			return "", err
		}
		d := fmt.Sprintf("\\frac{%s}{%s}", p1, p2)
		return fmt.Sprintf("\\mathchoice{\\textstyle%s}{%s}{%s}{%s}", d, d, d, d), nil
	case "tex-math":
		return mapStr(buf, "p1") + " ", nil
	case "frac-ce":
		p1, err := arrInner(buf, "p1")
		if err != nil {
			return "", err
		}
		p2, err := arrInner(buf, "p2")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("\\frac{%s}{%s}", p1, p2), nil
	case "overset":
		p1, err := arrInner(buf, "p1")
		if err != nil {
			return "", err
		}
		p2, err := arrInner(buf, "p2")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("\\overset{%s}{%s}", p1, p2), nil
	case "underset":
		p1, err := arrInner(buf, "p1")
		if err != nil {
			return "", err
		}
		p2, err := arrInner(buf, "p2")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("\\underset{%s}{%s}", p1, p2), nil
	case "underbrace":
		p1, err := arrInner(buf, "p1")
		if err != nil {
			return "", err
		}
		p2, err := arrInner(buf, "p2")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("\\underbrace{%s}_{%s}", p1, p2), nil
	case "color":
		c1 := mapStr(buf, "color1")
		c2, err := arrInner(buf, "color2")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("{\\color{%s}{%s}}", c1, c2), nil
	case "color0":
		c := mapStr(buf, "color")
		return fmt.Sprintf("\\color{%s}", c), nil
	case "arrow":
		rd, err := arrInner(buf, "rd")
		if err != nil {
			return "", err
		}
		rq, err := arrInner(buf, "rq")
		if err != nil {
			return "", err
		}
		r := mapStr(buf, "r")
		arrow, err := getArrow(r)
		if err != nil {
			return "", err
		}
		if rd != "" || rq != "" {
			switch r {
			case "<=>", "<=>>", "<<=>", "<-->":
				arrow = "\\long" + arrow
				if rd != "" {
					arrow = "\\overset{" + rd + "}{" + arrow + "}"
				}
				if rq != "" {
					lower := "\\lower6mu{"
					if r == "<-->" {
						lower = "\\lower2mu{"
					}
					arrow = "\\underset{" + lower + rq + "}}{" + arrow + "}"
				}
				arrow = " {}\\mathrel{" + arrow + "}{} "
			default:
				if rq != "" {
					arrow += "[{" + rq + "}]"
				}
				arrow += "{" + rd + "}"
				arrow = " {}\\mathrel{\\x" + arrow + "}{} "
			}
		} else {
			arrow = " {}\\mathrel{\\long" + arrow + "}{} "
		}
		return arrow, nil
	case "operator":
		k := mapStr(buf, "kind_")
		out, ok := getOperator(k)
		if !ok {
			return "", parserError("MhchemBugT", "mhchem bug T. Please report.")
		}
		return out, nil
	case "1st-level escape":
		return mapStr(buf, "p1") + " ", nil
	case "space":
		return " ", nil
	case "tinySkip":
		return "\\mkern2mu", nil
	case "entitySkip":
		return "~", nil
	case "pu-space-1":
		return "~", nil
	case "pu-space-2":
		return "\\mkern3mu ", nil
	case "1000 separator":
		return "\\mkern2mu ", nil
	case "commaDecimal":
		return "{,}", nil
	case "comma enumeration L":
		return fmt.Sprintf("{%s}\\mkern6mu ", mapStr(buf, "p1")), nil
	case "comma enumeration M":
		return fmt.Sprintf("{%s}\\mkern3mu ", mapStr(buf, "p1")), nil
	case "comma enumeration S":
		return fmt.Sprintf("{%s}\\mkern1mu ", mapStr(buf, "p1")), nil
	case "hyphen":
		return "\\text{-}", nil
	case "addition compound":
		return "\\,{\\cdot}\\,", nil
	case "electron dot":
		return "\\mkern1mu \\bullet\\mkern1mu ", nil
	case "KV x":
		return "{\\times}", nil
	case "prime":
		return "\\prime ", nil
	case "cdot":
		return "\\cdot ", nil
	case "tight cdot":
		return "\\mkern1mu{\\cdot}\\mkern1mu ", nil
	case "times":
		return "\\times ", nil
	case "circa":
		return "{\\sim}", nil
	case "^":
		return "uparrow", nil
	case "v":
		return "downarrow", nil
	case "ellipsis":
		return "\\ldots ", nil
	case "/":
		return "/", nil
	case " / ":
		return "\\,/\\,", nil
	}
	return "", parserError("MhchemBugT", "mhchem bug T. Please report.")
}

func texifyChemfive(buf Value) (string, error) {
	b5a, err := arrInner(buf, "a")
	if err != nil {
		return "", err
	}
	b5b, err := arrInner(buf, "b")
	if err != nil {
		return "", err
	}
	b5p, err := arrInner(buf, "p")
	if err != nil {
		return "", err
	}
	b5o, err := arrInner(buf, "o")
	if err != nil {
		return "", err
	}
	b5q, err := arrInner(buf, "q")
	if err != nil {
		return "", err
	}
	b5d, err := arrInner(buf, "d")
	if err != nil {
		return "", err
	}
	var res strings.Builder
	if b5a != "" {
		a := b5a
		if strings.HasPrefix(a, "+") || strings.HasPrefix(a, "-") {
			a = "{" + a + "}"
		}
		res.WriteString(a)
		res.WriteString("\\,")
	}
	if b5b != "" || b5p != "" {
		res.WriteString("{\\vphantom{A}}")
		fmt.Fprintf(&res, "^{\\hphantom{%s}}_{\\hphantom{%s}}", b5b, b5p)
		res.WriteString("\\mkern-1.5mu")
		res.WriteString("{\\vphantom{A}}")
		fmt.Fprintf(&res, "^{\\smash[t]{\\vphantom{2}}\\llap{%s}}", b5b)
		fmt.Fprintf(&res, "_{\\vphantom{2}\\llap{\\smash[t]{%s}}}", b5p)
	}
	if b5o != "" {
		o := b5o
		if strings.HasPrefix(o, "+") || strings.HasPrefix(o, "-") {
			o = "{" + o + "}"
		}
		res.WriteString(o)
	}
	dTy := mapStr(buf, "dType")
	switch dTy {
	case "kv":
		if b5d != "" || b5q != "" {
			res.WriteString("{\\vphantom{A}}")
		}
		if b5d != "" {
			fmt.Fprintf(&res, "^{%s}", b5d)
		}
		if b5q != "" {
			fmt.Fprintf(&res, "_{\\smash[t]{%s}}", b5q)
		}
	case "oxidation":
		if b5d != "" {
			res.WriteString("{\\vphantom{A}}")
			fmt.Fprintf(&res, "^{%s}", b5d)
		}
		if b5q != "" {
			res.WriteString("{\\vphantom{A}}")
			fmt.Fprintf(&res, "_{\\smash[t]{%s}}", b5q)
		}
	default:
		if b5q != "" {
			res.WriteString("{\\vphantom{A}}")
			fmt.Fprintf(&res, "_{\\smash[t]{%s}}", b5q)
		}
		if b5d != "" {
			res.WriteString("{\\vphantom{A}}")
			fmt.Fprintf(&res, "^{%s}", b5d)
		}
	}
	return res.String(), nil
}

func getArrow(a string) (string, error) {
	switch a {
	case "->", "→", "⟶":
		return "rightarrow", nil
	case "<-":
		return "leftarrow", nil
	case "<->":
		return "leftrightarrow", nil
	case "<-->":
		return "leftrightarrows", nil
	case "<=>", "⇌":
		return "rightleftharpoons", nil
	case "<=>>":
		return "Rightleftharpoons", nil
	case "<<=>":
		return "Leftrightharpoons", nil
	}
	return "", parserError("MhchemBugT", "mhchem bug T. Please report.")
}

func getBond(a string) (string, bool) {
	switch a {
	case "-", "1":
		return "{-}", true
	case "=", "2":
		return "{=}", true
	case "#", "3":
		return "{\\equiv}", true
	case "~":
		return "{\\tripledash}", true
	case "~-":
		return "{\\rlap{\\lower.1em{-}}\\raise.1em{\\tripledash}}", true
	case "~=", "~--":
		return "{\\rlap{\\lower.2em{-}}\\rlap{\\raise.2em{\\tripledash}}-}", true
	case "-~-":
		return "{\\rlap{\\lower.2em{-}}\\rlap{\\raise.2em{-}}\\tripledash}", true
	case "...":
		return "{{\\cdot}{\\cdot}{\\cdot}}", true
	case "....":
		return "{{\\cdot}{\\cdot}{\\cdot}{\\cdot}}", true
	case "->":
		return "{\\rightarrow}", true
	case "<-":
		return "{\\leftarrow}", true
	case "<":
		return "{<}", true
	case ">":
		return "{>}", true
	}
	return "", false
}

func getOperator(a string) (string, bool) {
	switch a {
	case "+":
		return " {}+{} ", true
	case "-":
		return " {}-{} ", true
	case "=":
		return " {}={} ", true
	case "<":
		return " {}<{} ", true
	case ">":
		return " {}>{} ", true
	case "<<":
		return " {}\\ll{} ", true
	case ">>":
		return " {}\\gg{} ", true
	case "\\pm":
		return " {}\\pm{} ", true
	case "\\approx", "$\\approx$":
		return " {}\\approx{} ", true
	case "v", "(v)":
		return " \\downarrow{} ", true
	case "^", "(^)":
		return " \\uparrow{} ", true
	}
	return "", false
}
