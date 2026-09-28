// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/base/BaseMethods.ts, MoveRoot.

package tex

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/d2lang/mathjax-go/internal/jscompat"
)

var rootIntegerPattern = regexp.MustCompile(`-?[0-9]+`)
var rootIntegerPrefix = regexp.MustCompile(`^[+-]?[0-9]+`)

// These handlers push ArrayItem subclasses, whose lexical environment does
// not inherit the enclosing index's inRoot flag. Equation and wrapper
// environments continue to inherit it normally.
func rootIndexArrayEnvironment(name string) bool {
	_, entry, ok := lookupMJSourceEntry(mjSourceEnvironmentMap, name)
	if !ok {
		return false
	}
	handler := entry.Value
	if list, ok := handler.(mjSourceList); ok && len(list) != 0 {
		handler = list[0]
	}
	switch handler {
	case "Array", "AlignedArray", "EqnArray", "AmsEqnArray", "AlignAt", "XalignAt", "FlalignArray",
		"Multline", "MtMultlined", "MtMatrix", "MtSmallMatrix", "Cases", "NumCases", "CD":
		return true
	}
	return false
}

func (p *parser) moveRoot(name string) error {
	if !p.inRoot {
		return texError("MisplacedMoveRoot", "\\%s can appear only within a root", name)
	}
	global := p.ensureStackGlobal()
	offset := &global.upRoot
	if name == "leftroot" {
		offset = &global.leftRoot
	}
	if *offset != "" {
		return texError("MultipleMoveRoot", "Multiple use of \\%s", name)
	}
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	// The original validity check is intentionally unanchored, while
	// parseInt reads only a signed decimal prefix after JavaScript whitespace.
	if !rootIntegerPattern.MatchString(raw) {
		return texError("IntegerArg", "The argument to \\%s must be an integer", name)
	}
	number := math.NaN()
	if prefix := rootIntegerPrefix.FindString(strings.TrimLeftFunc(raw, internalTextSpace)); prefix != "" {
		number, _ = strconv.ParseFloat(prefix, 64)
	}
	value := jscompat.NumberString(number/15) + "em"
	if !strings.HasPrefix(value, "-") {
		value = "+" + value
	}
	*offset = value
	return nil
}
