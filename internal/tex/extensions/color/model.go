// Copyright (c) 2018-2022 Omar Al-Ithawi and The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go translation and modification of MathJax 3.2.2 ColorUtil.ts.

package color

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	decimalPattern = regexp.MustCompile(`^(\d+(\.\d*)?|\.\d+)$`)
	integerPattern = regexp.MustCompile(`^\d+$`)
)

// Error is the color extension's source-compatible TeX error. ID is the
// localization key used by MathJax 3.2.2 and Message is its formatted English
// fallback.
type Error struct {
	ID      string
	Message string
}

func (e *Error) Error() string { return e.Message }

func texError(id, format string, args ...any) error {
	return &Error{ID: id, Message: fmt.Sprintf(format, args...)}
}

// Model stores user-defined colors for one parser. A Model must not be shared
// between parser instances; this preserves MathJax's packageData isolation.
// The zero value is ready for use.
type Model struct {
	userColors map[string]string
}

// NewModel creates an empty, parser-local color model.
func NewModel() *Model { return &Model{userColors: make(map[string]string)} }

// GetColor resolves def according to model. An empty model has MathJax's
// special named-color semantics.
func (m *Model) GetColor(model, def string) (string, error) {
	if model == "" || model == "named" {
		return m.getColorByName(def), nil
	}
	return normalizeColor(model, def)
}

// DefineColor creates or replaces one parser-local named color.
func (m *Model) DefineColor(model, name, def string) error {
	normalized, err := normalizeColor(model, def)
	if err != nil {
		return err
	}
	if m.userColors == nil {
		m.userColors = make(map[string]string)
	}
	m.userColors[name] = normalized
	return nil
}

func (m *Model) getColorByName(name string) string {
	if m != nil && m.userColors != nil {
		if value, ok := m.userColors[name]; ok {
			return value
		}
	}
	if value, ok := NamedColors[name]; ok {
		return value
	}
	// MathJax v2 compatibility: unknown values pass through to CSS. This also
	// permits CSS literals such as #ff0 and browser-known color names.
	return name
}

func normalizeColor(model, def string) (string, error) {
	if model == "" || model == "named" {
		return def, nil
	}
	switch model {
	case "rgb":
		return normalizeRGBDecimal(def)
	case "RGB":
		return normalizeRGBInteger(def)
	case "gray":
		return normalizeGray(def)
	default:
		return "", texError("UndefinedColorModel", "Color model '%s' not defined", model)
	}
}

func normalizeRGBDecimal(rgb string) (string, error) {
	parts := splitComponents(rgb)
	if len(parts) != 3 {
		return "", texError("ModelArg1", "Color values for the %s model require 3 numbers", "rgb")
	}
	var result strings.Builder
	result.WriteByte('#')
	for _, part := range parts {
		if !decimalPattern.MatchString(part) {
			return "", texError("InvalidDecimalNumber", "Invalid decimal number")
		}
		n, _ := strconv.ParseFloat(part, 64)
		if n < 0 || n > 1 {
			return "", texError("ModelArg2", "Color values for the %s model must be between %s and %s", "rgb", "0", "1")
		}
		writeHexByte(&result, int(math.Floor(n*255)))
	}
	return result.String(), nil
}

func normalizeRGBInteger(rgb string) (string, error) {
	parts := splitComponents(rgb)
	if len(parts) != 3 {
		return "", texError("ModelArg1", "Color values for the %s model require 3 numbers", "RGB")
	}
	var result strings.Builder
	result.WriteByte('#')
	for _, part := range parts {
		if !integerPattern.MatchString(part) {
			return "", texError("InvalidNumber", "Invalid number")
		}
		// Avoid machine-integer overflow while retaining parseInt's behavior for
		// arbitrarily many leading zeroes.
		digits := strings.TrimLeft(part, "0")
		if digits == "" {
			digits = "0"
		}
		if len(digits) > 3 {
			return "", texError("ModelArg2", "Color values for the %s model must be between %s and %s", "RGB", "0", "255")
		}
		n, _ := strconv.Atoi(digits)
		if n > 255 {
			return "", texError("ModelArg2", "Color values for the %s model must be between %s and %s", "RGB", "0", "255")
		}
		writeHexByte(&result, n)
	}
	return result.String(), nil
}

func normalizeGray(gray string) (string, error) {
	gray = trimECMAScriptSpace(gray)
	if !decimalPattern.MatchString(gray) {
		return "", texError("InvalidDecimalNumber", "Invalid decimal number")
	}
	n, _ := strconv.ParseFloat(gray, 64)
	if n < 0 || n > 1 {
		return "", texError("ModelArg2", "Color values for the %s model must be between %s and %s", "gray", "0", "1")
	}
	pn := int(math.Floor(n * 255))
	return fmt.Sprintf("#%02x%02x%02x", pn, pn, pn), nil
}

func splitComponents(value string) []string {
	parts := strings.Split(trimECMAScriptSpace(value), ",")
	for i := range parts {
		parts[i] = trimECMAScriptSpace(parts[i])
	}
	return parts
}

// trimECMAScriptSpace matches the WhiteSpace and LineTerminator code points
// recognized by JavaScript's String.trim() and regular-expression \s in the
// MathJax 3.2.2 implementation. In particular, Go's strings.TrimSpace does not
// include U+FEFF.
func trimECMAScriptSpace(value string) string {
	return strings.TrimFunc(value, func(r rune) bool {
		switch r {
		case '\u0009', '\u000A', '\u000B', '\u000C', '\u000D', '\u0020',
			'\u00A0', '\u1680', '\u2028', '\u2029', '\u202F', '\u205F',
			'\u3000', '\uFEFF':
			return true
		default:
			return r >= '\u2000' && r <= '\u200A'
		}
	})
}

func writeHexByte(builder *strings.Builder, value int) {
	fmt.Fprintf(builder, "%02x", value)
}
