// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"unicode/utf8"
)

func TestUnicodePrimeRegisteredCollector(t *testing.T) {
	testPrimeCollectorReferences(t, "testdata/unicode_prime_collector_mathjax_3_2_2.json", 40)
}

func TestPrimeWhitespaceRegisteredCollector(t *testing.T) {
	testPrimeCollectorReferences(t, "testdata/prime_whitespace_collector_mathjax_3_2_2.json", 148)
}

func testPrimeCollectorReferences(t *testing.T, fixture string, count int) {
	t.Helper()
	var f struct {
		Rows []struct {
			Source                 string
			Reject                 bool
			Error                  *string
			CursorUTF16            int
			Remaining              string
			BaseUnchanged          bool
			BaseIdentity           *bool
			Token                  *string
			Attributes, Properties map[string]any
		}
	}
	b, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Rows) != count {
		t.Fatal("missing registered collector records")
	}

	for _, c := range f.Rows {
		t.Run(c.Source+map[bool]string{false: "/collect", true: "/reject"}[c.Reject], func(t *testing.T) {
			core := limitToken("mi", "x")
			base := core
			if c.Reject {
				base = node("msubsup", core, limitToken("mi", "i"), limitToken("mi", "n"))
			}
			before := base.Clone()
			_, size := utf8.DecodeRuneInString(c.Source)
			p := &parser{source: c.Source, pos: size, state: newParseState()}
			pending, err := p.startPrime(base)
			if c.Error != nil {
				if !c.Reject || c.CursorUTF16 != 1 || c.Remaining != c.Source[size:] || c.Token != nil || c.BaseIdentity != nil || err == nil || err.Error() != *c.Error || p.pos != size || !reflect.DeepEqual(base, before) || !c.BaseUnchanged {
					t.Fatalf("rejected collector changed cursor/base or error: %v", err)
				}
				return
			}
			if c.Reject || c.Token == nil {
				t.Fatal("unexpected successful primary rejection record")
			}
			if err != nil {
				t.Fatal(err)
			}
			if pending.base != base || c.BaseIdentity == nil || !*c.BaseIdentity || !reflect.DeepEqual(base, before) || !c.BaseUnchanged {
				t.Fatal("collector must retain original base")
			}
			expectedToken, expectedRemaining := *c.Token, c.Remaining
			expectedByteCursor := len(c.Source) - len(c.Remaining)
			expectedCursorUTF16 := c.CursorUTF16
			if pending.prime.Kind != "mo" || len(pending.prime.Children) != 1 || pending.prime.Children[0].Kind != "text" {
				t.Fatal("unexpected prime token structure")
			}
			if pending.prime.Children[0].Text != expectedToken || p.source[p.pos:] != expectedRemaining {
				t.Fatalf("token=%q tail=%q want=%q/%q", pending.prime.Children[0].Text, p.source[p.pos:], expectedToken, expectedRemaining)
			}
			if !utf8.ValidString(p.source[:p.pos]) || !utf8.ValidString(p.source[p.pos:]) {
				t.Fatal("collector split a UTF-8 code point")
			}
			if p.pos != expectedByteCursor || expectedByteCursor != len(c.Source)-len(expectedRemaining) {
				t.Fatal("incorrect byte cursor")
			}
			units := 0
			for _, r := range c.Source[:p.pos] {
				units++
				if r > 0xffff {
					units++
				}
			}
			if units != expectedCursorUTF16 {
				t.Fatal("source UTF-16 and byte cursor disagree")
			}
			// Match the inherited D060 private-method contract: the Go-only token
			// factory marker exists during parsing and is removed by Compile.
			props := ownMap(pending.prime)
			if props[vectorFactoryToken] != true {
				t.Fatal("missing parser token provenance")
			}
			delete(props, vectorFactoryToken)
			if !reflect.DeepEqual(explicitMap(pending.prime), c.Attributes) || !reflect.DeepEqual(props, c.Properties) {
				t.Fatal("prime token attributes/properties differ")
			}
		})
	}
}
