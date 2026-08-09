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
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// jsRegexp is the small compatibility layer needed by mhchemparser's anchored
// JavaScript regular expressions.  Most expressions are mechanically lowered
// to Go's RE2 syntax.  The finite set that uses look-ahead is implemented by
// native anchored matchers below, avoiding a backtracking-regexp dependency.
type jsRegexp struct {
	re         *regexp.Regexp
	captureMap []int
	custom     func(string) (string, bool)
}

type jsGroup struct {
	text   string
	Length int
}

func (g jsGroup) String() string { return g.text }

type jsMatch struct {
	Index  int
	Length int
	text   string
	groups []jsGroup
}

// ECMAScript's \s set is fixed and differs from both RE2's ASCII \s and
// unicode.IsSpace (which also accepts characters such as U+0085).
const jsWhitespaceClassBody = `\x{0009}-\x{000D}\x{0020}\x{00A0}\x{1680}\x{2000}-\x{200A}\x{2028}\x{2029}\x{202F}\x{205F}\x{3000}\x{FEFF}`

func isJSWhitespace(r rune) bool {
	switch r {
	case '\u0009', '\u000A', '\u000B', '\u000C', '\u000D', '\u0020', '\u00A0', '\u1680', '\u2028', '\u2029', '\u202F', '\u205F', '\u3000', '\uFEFF':
		return true
	}
	return r >= '\u2000' && r <= '\u200A'
}

func (m *jsMatch) String() string    { return m.text }
func (m *jsMatch) Groups() []jsGroup { return m.groups }

func (r *jsRegexp) FindStringMatch(input string) (*jsMatch, error) {
	if r.custom != nil {
		matched, ok := r.custom(input)
		if !ok {
			return nil, nil
		}
		return &jsMatch{Length: len(matched), text: matched, groups: []jsGroup{{text: matched, Length: len(matched)}}}, nil
	}
	indices := r.re.FindStringSubmatchIndex(input)
	if indices == nil {
		return nil, nil
	}
	full := input[indices[0]:indices[1]]
	groups := []jsGroup{{text: full, Length: len(full)}}
	for _, translatedIndex := range r.captureMap {
		pair := translatedIndex * 2
		text := ""
		if pair+1 < len(indices) && indices[pair] >= 0 {
			text = input[indices[pair]:indices[pair+1]]
		}
		groups = append(groups, jsGroup{text: text, Length: len(text)})
	}
	return &jsMatch{Index: indices[0], Length: indices[1] - indices[0], text: full, groups: groups}, nil
}

func compileJSRegexp(source string) (*jsRegexp, error) {
	if custom := customLookaheadMatcher(source); custom != nil {
		return &jsRegexp{custom: custom}, nil
	}
	translated, captures, err := lowerJSRegexp(source)
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(translated)
	if err != nil {
		return nil, err
	}
	return &jsRegexp{re: re, captureMap: captures}, nil
}

// lowerJSRegexp converts JavaScript Unicode/identity escapes and
// non-capturing groups.  captureMap retains only the groups that JavaScript
// would expose, despite the extra Go groups used for (?:...).
func lowerJSRegexp(source string) (string, []int, error) {
	var out strings.Builder
	var captures []int
	groups := 0
	inClass := false
	for i := 0; i < len(source); {
		if source[i] == '[' {
			inClass = true
			out.WriteByte(source[i])
			i++
			continue
		}
		if source[i] == ']' && inClass {
			inClass = false
			out.WriteByte(source[i])
			i++
			continue
		}
		if source[i] == '\\' {
			if i+1 >= len(source) {
				return "", nil, fmt.Errorf("dangling escape")
			}
			if i+6 <= len(source) && source[i+1] == 'u' {
				if _, err := strconv.ParseUint(source[i+2:i+6], 16, 16); err == nil {
					out.WriteString(`\x{`)
					out.WriteString(source[i+2 : i+6])
					out.WriteByte('}')
					i += 6
					continue
				}
			}
			next := source[i+1]
			switch next {
			case 's':
				if inClass {
					out.WriteString(jsWhitespaceClassBody)
				} else {
					out.WriteByte('[')
					out.WriteString(jsWhitespaceClassBody)
					out.WriteByte(']')
				}
			case '/', ',', '_', ' ', ';', ':':
				out.WriteByte(next)
			case '-':
				out.WriteString(`\x{2d}`)
			default:
				out.WriteByte('\\')
				out.WriteByte(next)
			}
			i += 2
			continue
		}
		if source[i] == '(' && !inClass {
			if strings.HasPrefix(source[i:], "(?:") {
				groups++
				out.WriteByte('(')
				i += 3
				continue
			}
			if strings.HasPrefix(source[i:], "(?=") || strings.HasPrefix(source[i:], "(?!") {
				return "", nil, fmt.Errorf("unsupported look-ahead in %q", source)
			}
			groups++
			captures = append(captures, groups)
		}
		out.WriteByte(source[i])
		i++
	}
	return out.String(), captures, nil
}

var (
	amountNumber = regexp.MustCompile(`^(\([+-]?[0-9]+/[0-9]+\)|[+-]?([0-9]+|\$[a-z]\$|[a-z])/[0-9]+|[+-]?[0-9]+[.,][0-9]+|[+-]?\.[0-9]+|[+-]?[0-9]+)`)
	parenLower   = regexp.MustCompile(`^\([a-z]+\)`)
)

var greekCommands = []string{
	"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta", "iota", "kappa", "lambda", "mu", "nu", "xi", "omicron", "pi", "rho", "sigma", "tau", "upsilon", "phi", "chi", "psi", "omega",
	"Gamma", "Delta", "Theta", "Lambda", "Xi", "Pi", "Sigma", "Upsilon", "Phi", "Psi", "Omega",
}

var lowerGreekCommands = greekCommands[:24]

func isASCIILetter(b byte) bool { return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' }

func commandMatch(input string, names []string) (int, bool) {
	if !strings.HasPrefix(input, `\`) {
		return 0, false
	}
	for _, name := range names {
		prefix := `\` + name
		if !strings.HasPrefix(input, prefix) {
			continue
		}
		i := len(prefix)
		if i == len(input) {
			return i, true
		}
		if strings.HasPrefix(input[i:], "{}") {
			return i + 2, true
		}
		r, width := utf8.DecodeRuneInString(input[i:])
		if isJSWhitespace(r) {
			i += width
			for i < len(input) {
				r, width = utf8.DecodeRuneInString(input[i:])
				if !isJSWhitespace(r) {
					break
				}
				i += width
			}
			return i, true
		}
		if !isASCIILetter(input[i]) {
			return i, true
		}
	}
	return 0, false
}

func lettersMatch(input string) int {
	i := 0
	for i < len(input) {
		r, width := utf8.DecodeRuneInString(input[i:])
		if isASCIILetter(input[i]) || r >= 'α' && r <= 'ω' || r >= 'Α' && r <= 'Ω' || r == '?' || r == '@' {
			i += width
			continue
		}
		if n, ok := commandMatch(input[i:], greekCommands); ok {
			i += n
			continue
		}
		break
	}
	return i
}

func lowerGreekFullMatch(input string) bool {
	i := 0
	if i < len(input) && input[i] == '$' {
		i++
	}
	if i >= len(input) {
		return false
	}
	r, width := utf8.DecodeRuneInString(input[i:])
	if r >= 'α' && r <= 'ω' {
		i += width
	} else {
		found := false
		for _, name := range lowerGreekCommands {
			prefix := `\` + name
			if strings.HasPrefix(input[i:], prefix) {
				i += len(prefix)
				found = true
				for i < len(input) {
					r, width = utf8.DecodeRuneInString(input[i:])
					if !isJSWhitespace(r) {
						break
					}
					i += width
				}
				break
			}
		}
		if !found {
			return false
		}
	}
	if i < len(input) && input[i] == '$' {
		i++
	}
	if i == len(input) {
		return true
	}
	if strings.HasPrefix(input[i:], "{}") {
		i += 2
	} else {
		for i < len(input) {
			r, width = utf8.DecodeRuneInString(input[i:])
			if !isJSWhitespace(r) {
				return false
			}
			i += width
		}
	}
	return i == len(input)
}

func nextNonSpaceUpper(input string) bool {
	for len(input) > 0 {
		r, width := utf8.DecodeRuneInString(input)
		if !isJSWhitespace(r) {
			return r >= 'A' && r <= 'Z'
		}
		input = input[width:]
	}
	return false
}

func matchAmountMain(input string) (string, bool) {
	if loc := amountNumber.FindStringIndex(input); loc != nil && loc[0] == 0 {
		end := loc[1]
		if end < len(input) && input[end] >= 'a' && input[end] <= 'z' && nextNonSpaceUpper(input[end+1:]) {
			end++
		}
		return input[:end], true
	}
	i := 0
	if i < len(input) && (input[i] == '+' || input[i] == '-') {
		i++
	}
	if i < len(input) && input[i] >= 'a' && input[i] <= 'z' && nextNonSpaceUpper(input[i+1:]) {
		return input[:i+1], true
	}
	if strings.HasPrefix(input, "+") {
		if len(input) == 1 {
			return "+", true
		}
		r, _ := utf8.DecodeRuneInString(input[1:])
		if !isJSWhitespace(r) {
			return "+", true
		}
	}
	return "", false
}

func boundary(input string, end int) bool { return end == len(input) || !isASCIILetter(input[end]) }

func customLookaheadMatcher(source string) func(string) (string, bool) {
	switch {
	case source == `^.`:
		return func(input string) (string, bool) {
			if input == "" {
				return "", false
			}
			r, width := utf8.DecodeRuneInString(input)
			if r == '\n' || r == '\r' || r == '\u2028' || r == '\u2029' {
				return "", false
			}
			return input[:width], true
		}
	case source == `^\([a-z]{1,3}(?=[\),])`:
		return func(input string) (string, bool) {
			if len(input) < 3 || input[0] != '(' {
				return "", false
			}
			i := 1
			for i < len(input) && i <= 3 && input[i] >= 'a' && input[i] <= 'z' {
				i++
			}
			if i == 1 || i >= len(input) || input[i] != ',' && input[i] != ')' {
				return "", false
			}
			return input[:i], true
		}
	case source == `^(?=\{)`:
		return func(input string) (string, bool) { return "", strings.HasPrefix(input, "{") }
	case strings.Contains(source, `[a-z](?=\s*[A-Z])`) && strings.Contains(source, `\+(?!\s)`):
		return matchAmountMain
	case source == `^(?:[A-Z][a-z]{0,2}|i)(?=,)`:
		return func(input string) (string, bool) {
			if strings.HasPrefix(input, "i,") {
				return "i", true
			}
			if input == "" || input[0] < 'A' || input[0] > 'Z' {
				return "", false
			}
			i := 1
			for i < len(input) && i < 3 && input[i] >= 'a' && input[i] <= 'z' {
				i++
			}
			return input[:i], i < len(input) && input[i] == ','
		}
	case strings.HasPrefix(source, `^-(?=(?:[spd]|sp)`):
		return func(input string) (string, bool) {
			if !strings.HasPrefix(input, "-") {
				return "", false
			}
			rest := input[1:]
			for _, orbital := range []string{"sp", "s", "p", "d"} {
				if strings.HasPrefix(rest, orbital) {
					tail := rest[len(orbital):]
					if tail == "" {
						return "-", true
					}
					r, _ := utf8.DecodeRuneInString(tail)
					if isJSWhitespace(r) || strings.ContainsRune(",;)]}", r) {
						return "-", true
					}
				}
			}
			return "", false
		}
	case strings.HasPrefix(source, `^-(?=[\s_},;`) && strings.Contains(source, `\([a-z]+\))`):
		return func(input string) (string, bool) {
			if !strings.HasPrefix(input, "-") {
				return "", false
			}
			rest := input[1:]
			if rest == "" || strings.ContainsRune("_},;]/", rune(rest[0])) || parenLower.MatchString(rest) {
				return "-", true
			}
			r, _ := utf8.DecodeRuneInString(rest)
			return "-", isJSWhitespace(r)
		}
	case source == `^-(?=[0-9])`:
		return func(input string) (string, bool) {
			return "-", len(input) > 1 && input[0] == '-' && input[1] >= '0' && input[1] <= '9'
		}
	case source == `^\.\.\.(?=$|[^.])`:
		return func(input string) (string, bool) {
			return "...", strings.HasPrefix(input, "...") && (len(input) == 3 || input[3] != '.')
		}
	case source == `^[CMT](?=\[)`:
		return func(input string) (string, bool) {
			if len(input) < 2 || !strings.ContainsRune("CMT", rune(input[0])) || input[1] != '[' {
				return "", false
			}
			return input[:1], true
		}
	case strings.HasPrefix(source, `^\\ca(?:\s+|`):
		return func(input string) (string, bool) {
			if !strings.HasPrefix(input, `\ca`) {
				return "", false
			}
			i := 3
			for i < len(input) {
				r, width := utf8.DecodeRuneInString(input[i:])
				if !isJSWhitespace(r) {
					break
				}
				i += width
			}
			return input[:i], i > 3 || i == len(input) || !isASCIILetter(input[i])
		}
	case strings.HasPrefix(source, `^\\(?:alpha|beta`) && strings.Contains(source, `(?![a-zA-Z])`):
		return func(input string) (string, bool) {
			n, ok := commandMatch(input, greekCommands)
			if !ok {
				return "", false
			}
			return input[:n], true
		}
	case source == `^(?:\^(?=_)|\_(?=\^)|[\^_]$)`:
		return func(input string) (string, bool) {
			if strings.HasPrefix(input, "^_") || input == "^" {
				return "^", true
			}
			if strings.HasPrefix(input, "_^") || input == "_" {
				return "_", true
			}
			return "", false
		}
	case strings.HasPrefix(source, `^(?:v|\(v\)|\^|\(\^\))(?=`):
		return func(input string) (string, bool) {
			for _, op := range []string{"(v)", "(^)", "v", "^"} {
				if strings.HasPrefix(input, op) {
					tail := input[len(op):]
					if tail == "" {
						return op, true
					}
					r, _ := utf8.DecodeRuneInString(tail)
					if isJSWhitespace(r) || strings.ContainsRune(",;)]}", r) {
						return op, true
					}
				}
			}
			return "", false
		}
	case strings.HasPrefix(source, `^(?:[a-zA-Z\u03B1-`):
		return func(input string) (string, bool) {
			n := lettersMatch(input)
			return input[:n], n > 0
		}
	case strings.HasPrefix(source, `^(?:\$?[\u03B1-`) && strings.HasSuffix(source, `$`):
		return func(input string) (string, bool) { return input, lowerGreekFullMatch(input) }
	case strings.HasPrefix(source, `^(?:\+|(?:[\-=<>]`):
		return func(input string) (string, bool) {
			if strings.HasPrefix(input, "+") {
				return "+", true
			}
			for _, op := range []string{`$\approx$`, `\approx`, "<<", ">>", "-", "=", "<", ">"} {
				if strings.HasPrefix(input, op) {
					tail := input[len(op):]
					if tail == "" {
						return op, true
					}
					r, _ := utf8.DecodeRuneInString(tail)
					if isJSWhitespace(r) || tail[0] >= '0' && tail[0] <= '9' || len(tail) > 1 && tail[0] == '-' && tail[1] >= '0' && tail[1] <= '9' {
						return op, true
					}
				}
			}
			return "", false
		}
	case strings.HasPrefix(source, `^(?:[0-9]{1,2}[spdfgh]`):
		return func(input string) (string, bool) {
			digits := 0
			for digits < len(input) && digits < 2 && input[digits] >= '0' && input[digits] <= '9' {
				digits++
			}
			if strings.HasPrefix(input[digits:], "sp") && boundary(input, digits+2) {
				return input[:digits+2], true
			}
			if digits > 0 && digits < len(input) && strings.ContainsRune("spdfgh", rune(input[digits])) && boundary(input, digits+1) {
				return input[:digits+1], true
			}
			return "", false
		}
	case source == `^\s(?=[A-Z\\$])`:
		return func(input string) (string, bool) {
			if input == "" {
				return "", false
			}
			r, width := utf8.DecodeRuneInString(input)
			return input[:width], isJSWhitespace(r) && width < len(input) && (input[width] >= 'A' && input[width] <= 'Z' || input[width] == '\\' || input[width] == '$')
		}
	case strings.HasPrefix(source, `^(?:pH|pOH|pC|pK|iPr|iBu)(?=`):
		return func(input string) (string, bool) {
			for _, entity := range []string{"pOH", "pH", "pC", "pK", "iPr", "iBu"} {
				if strings.HasPrefix(input, entity) && boundary(input, len(entity)) {
					return entity, true
				}
			}
			return "", false
		}
	case source == `^\{\}(?=\^)`:
		return func(input string) (string, bool) { return "{}", strings.HasPrefix(input, "{}^") }
	}
	return nil
}
