// Copyright (c) 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

const (
	mjOperatorPackedFile = "operator_dictionary_packed_gen.go"
	mjOperatorPackedEnv  = "MATHJAX_GO_WRITE_OPERATOR_DICTIONARY"
)

// mjOperatorReferenceForms returns the source-shaped dictionaries in the
// order of the packed forms.
func mjOperatorReferenceForms() [3][]mjOperatorDictionaryEntry {
	return [3][]mjOperatorDictionaryEntry{
		mjOperatorPrefix:  mjOperatorDictionaryPrefix,
		mjOperatorPostfix: mjOperatorDictionaryPostfix,
		mjOperatorInfix:   mjOperatorDictionaryInfix,
	}
}

// renderMJOperatorPackedFile writes operator_dictionary_packed_gen.go from
// the source-shaped tables. inheritance.go describes the layout.
func renderMJOperatorPackedFile() ([]byte, error) {
	var (
		starts        = []int{0}
		text          strings.Builder
		offsets       = []int{0}
		indexes       []int
		words         []int
		wordNumber    = map[int]int{}
		lists         = []mjOperatorProperties{nil}
		listNumber    = map[string]int{"": 0}
		textLines     []string
		operatorCount int
	)
	for form, entries := range mjOperatorReferenceForms() {
		sorted := append([]mjOperatorDictionaryEntry(nil), entries...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Operator < sorted[j].Operator })
		line := ""
		for i, entry := range sorted {
			// The infix dictionary has an entry for the empty operator.
			if !utf8.ValidString(entry.Operator) {
				return nil, fmt.Errorf("form %d: operator %q is not UTF-8", form, entry.Operator)
			}
			if i > 0 && sorted[i-1].Operator == entry.Operator {
				return nil, fmt.Errorf("form %d: operator %q occurs twice", form, entry.Operator)
			}
			definition := entry.Definition
			key := ""
			if len(definition.Properties) != 0 {
				key = fmt.Sprintf("%#v", definition.Properties)
			} else if definition.Properties != nil {
				return nil, fmt.Errorf("operator %q has an empty property list that is not nil", entry.Operator)
			}
			list, ok := listNumber[key]
			if !ok {
				list = len(lists)
				listNumber[key] = list
				lists = append(lists, definition.Properties)
			}
			if definition.Lspace < 0 || definition.Lspace >= 1<<mjOperatorSpaceBits ||
				definition.Rspace < 0 || definition.Rspace >= 1<<mjOperatorSpaceBits ||
				definition.TexClass < -1 || definition.TexClass+1 >= 1<<mjOperatorTexClassBits ||
				list >= 1<<(16-2*mjOperatorSpaceBits-mjOperatorTexClassBits) {
				return nil, fmt.Errorf("operator %q: definition %#v does not fit a word", entry.Operator, definition)
			}
			word := definition.Lspace | definition.Rspace<<mjOperatorSpaceBits |
				(definition.TexClass+1)<<(2*mjOperatorSpaceBits) |
				list<<(2*mjOperatorSpaceBits+mjOperatorTexClassBits)
			number, ok := wordNumber[word]
			if !ok {
				number = len(words)
				wordNumber[word] = number
				words = append(words, word)
			}
			indexes = append(indexes, number)
			text.WriteString(entry.Operator)
			offsets = append(offsets, text.Len())
			line += entry.Operator
			if (i+1)%16 == 0 || i == len(sorted)-1 {
				textLines = append(textLines, strconv.QuoteToASCII(line))
				line = ""
			}
		}
		operatorCount += len(sorted)
		starts = append(starts, operatorCount)
	}
	if text.Len() >= 1<<16 || len(words) > 1<<8 {
		return nil, fmt.Errorf("%d bytes of operators or %d definitions do not fit the tables", text.Len(), len(words))
	}

	var out bytes.Buffer
	numbers := func(values []int, format string) {
		for i, value := range values {
			if i%16 == 0 {
				out.WriteString("\n\t")
			}
			fmt.Fprintf(&out, format+", ", value)
		}
		out.WriteString("\n}\n\n")
	}
	out.WriteString(`// Copyright (c) 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Code generated from the tables of operator_dictionary_reference_test.go
// (MathJax 3.2.2 ts/core/MmlTree/OperatorDictionary.ts) by
//
//	` + mjOperatorPackedEnv + `=1 go test -run TestMJOperatorDictionaryPacked ./internal/tex
//
// DO NOT EDIT.

package tex

// The operators of the prefix, postfix and infix dictionaries are entries
// mjOperatorFormStart[form] up to mjOperatorFormStart[form+1] of the tables
// below, sorted by their bytes within a form.
var mjOperatorFormStart = [...]uint16{`)
	for _, start := range starts {
		fmt.Fprintf(&out, "%d, ", start)
	}
	out.WriteString("}\n\n// Operator i is mjOperatorText[mjOperatorOffsets[i]:mjOperatorOffsets[i+1]].\nconst mjOperatorText = \"\" +\n\t")
	out.WriteString(strings.Join(textLines, " +\n\t"))
	out.WriteString("\n\nvar mjOperatorOffsets = [...]uint16{")
	numbers(offsets, "%d")
	out.WriteString("// mjOperatorDefinitionIndex gives, for every operator, the position of its\n// definition in mjOperatorDefinitionWords.\nvar mjOperatorDefinitionIndex = [...]uint8{")
	numbers(indexes, "%d")
	out.WriteString("// mjOperatorDefinitionWords holds the distinct definitions; see\n// operatorDefinitionAt.\nvar mjOperatorDefinitionWords = [...]uint16{")
	numbers(words, "%#04x")
	out.WriteString("// mjOperatorPropertyLists holds the distinct property lists of the\n// definitions. The first is that of a definition without properties.\nvar mjOperatorPropertyLists = [...]mjOperatorProperties{\n\tnil,\n")
	for _, list := range lists[1:] {
		out.WriteString("\t{")
		for i, property := range list {
			if i > 0 {
				out.WriteString(", ")
			}
			switch value := property.Value.(type) {
			case bool:
				fmt.Fprintf(&out, "{Name: %q, Value: %t}", property.Name, value)
			case string:
				fmt.Fprintf(&out, "{Name: %q, Value: %q}", property.Name, value)
			default:
				return nil, fmt.Errorf("property %q has unsupported type %T", property.Name, property.Value)
			}
		}
		out.WriteString("},\n")
	}
	out.WriteString("}\n")
	return format.Source(out.Bytes())
}

// TestMJOperatorDictionaryPacked keeps the packed dictionary that the parser
// reads identical to what the source-shaped tables produce.
func TestMJOperatorDictionaryPacked(t *testing.T) {
	want, err := renderMJOperatorPackedFile()
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv(mjOperatorPackedEnv) != "" {
		if err := os.WriteFile(mjOperatorPackedFile, want, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s; run the tests again without %s", mjOperatorPackedFile, mjOperatorPackedEnv)
		return
	}
	got, err := os.ReadFile(mjOperatorPackedFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s is not what the reference tables produce; regenerate it with %s=1", mjOperatorPackedFile, mjOperatorPackedEnv)
	}

	forms := mjOperatorReferenceForms()
	if len(mjOperatorFormStart) != len(forms)+1 || mjOperatorFormStart[0] != 0 {
		t.Fatalf("mjOperatorFormStart = %v", mjOperatorFormStart)
	}
	count := int(mjOperatorFormStart[len(forms)])
	if len(mjOperatorOffsets) != count+1 || len(mjOperatorDefinitionIndex) != count ||
		mjOperatorOffsets[0] != 0 || int(mjOperatorOffsets[count]) != len(mjOperatorText) {
		t.Fatalf("tables do not describe %d operators in %d bytes", count, len(mjOperatorText))
	}
	operator := func(i int) string { return mjOperatorText[mjOperatorOffsets[i]:mjOperatorOffsets[i+1]] }
	for form, entries := range forms {
		start, end := int(mjOperatorFormStart[form]), int(mjOperatorFormStart[form+1])
		if end-start != len(entries) {
			t.Fatalf("form %d has %d packed operators, the source %d", form, end-start, len(entries))
		}
		for i := start; i < end; i++ {
			if mjOperatorOffsets[i] > mjOperatorOffsets[i+1] {
				t.Fatalf("the offsets of operator %d decrease", i)
			}
			if i > start && operator(i-1) >= operator(i) {
				t.Fatalf("form %d: %q and %q are not in increasing order", form, operator(i-1), operator(i))
			}
			if int(mjOperatorDefinitionIndex[i]) >= len(mjOperatorDefinitionWords) {
				t.Fatalf("operator %q: definition %d is out of range", operator(i), mjOperatorDefinitionIndex[i])
			}
		}
	}
}

// mjOperatorReferenceLookup is the lookup over the source-shaped tables that
// the packed dictionary replaced: the first entry with the operator, in the
// first form that has one.
func mjOperatorReferenceLookup(text string, forms []string) (mjOperatorDefinition, bool) {
	for _, form := range forms {
		var dictionary []mjOperatorDictionaryEntry
		switch form {
		case "prefix":
			dictionary = mjOperatorDictionaryPrefix
		case "postfix":
			dictionary = mjOperatorDictionaryPostfix
		default:
			dictionary = mjOperatorDictionaryInfix
		}
		for _, entry := range dictionary {
			if entry.Operator == text {
				return entry.Definition, true
			}
		}
	}
	return mjOperatorDefinition{}, false
}

// TestMJOperatorLookupMatchesReference asks for every operator of the
// dictionary and for strings near them, in every order of the forms.
func TestMJOperatorLookupMatchesReference(t *testing.T) {
	probes := map[string]bool{"": true, "x": true, "D058-test": true, "\x00": true, "\xff": true, "\U0010FFFF": true}
	for _, entries := range mjOperatorReferenceForms() {
		for _, entry := range entries {
			operator := entry.Operator
			probes[operator] = true
			probes[operator+"x"] = true
			probes[operator+"\x00"] = true
			probes[" "+operator] = true
			probes[operator+operator] = true
			for i := 1; i < len(operator); i++ {
				probes[operator[:i]] = true
				probes[operator[i:]] = true
			}
			if _, size := utf8.DecodeRuneInString(operator); size < len(operator) {
				probes[operator[:size]] = true
			}
			if operator != "" {
				last := []byte(operator)
				last[len(last)-1]++
				probes[string(last)] = true
				last[len(last)-1] -= 2
				probes[string(last)] = true
			}
		}
	}
	formLists := [][]string{
		nil,
		{"prefix"}, {"postfix"}, {"infix"}, {"other"}, {""},
		{"prefix", "infix", "postfix"}, {"postfix", "infix", "prefix"}, {"infix", "prefix", "postfix"},
		{"prefix", "postfix", "infix"}, {"postfix", "prefix", "infix"}, {"infix", "postfix", "prefix"},
		{"other", "prefix"}, {"postfix", "postfix", "prefix"},
	}
	hits := 0
	for text := range probes {
		for _, forms := range formLists {
			want, wantOK := mjOperatorReferenceLookup(text, forms)
			got, gotOK := lookupOperatorDefinition(text, forms)
			if gotOK != wantOK || !reflect.DeepEqual(got, want) {
				t.Fatalf("lookupOperatorDefinition(%q, %q) = %#v, %v; the source tables give %#v, %v", text, forms, got, gotOK, want, wantOK)
			}
			if gotOK {
				hits++
			}
		}
	}
	if len(probes) < 4000 || hits < 9000 {
		t.Fatalf("only %d probes and %d hits", len(probes), hits)
	}
}
