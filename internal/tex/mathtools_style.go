// Copyright (c) 2020-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/mathtools/{MathtoolsMethods,MathtoolsUtil}.ts.

package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// Mathtools reads a literal optional style name, without expanding it. Its
// accepted keys differ from the numeric styles used by the fraction handlers.
func setMathtoolsDisplayLevel(styled *mml.Node, raw string) {
	display, level := false, 0
	switch strings.TrimFunc(raw, internalTextSpace) {
	case "\\displaystyle":
		display = true
	case "\\textstyle":
	case "\\scriptstyle":
		level = 1
	case "\\scriptscriptstyle":
		level = 2
	default:
		return
	}
	styled.Attributes.Set("displaystyle", display)
	styled.Attributes.Set("scriptlevel", level)
}
