// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package font

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"testing"
)

// effectiveLookupDigest hashes what Lookup, LookupMetrics, LookupPath and
// RemapCodepoint return for every variant, an unknown variant, and every code
// point of planes 0 to 2 together with some values outside them. Only the
// exported API is used, so the digest does not depend on how the tables,
// variant rules and alphanumeric remapping are stored.
func effectiveLookupDigest() string {
	hash := sha256.New()
	codepoints := make([]rune, 0, 0x30000+8)
	for codepoint := rune(0); codepoint < 0x30000; codepoint++ {
		codepoints = append(codepoints, codepoint)
	}
	codepoints = append(codepoints, -1, math.MinInt32, 0x30000, 0xE0001, 0x10FFFF, 0x110000, math.MaxInt32)
	variantNumber := make(map[Variant]uint64)
	for i, variant := range Variants() {
		variantNumber[variant] = uint64(i) + 1
	}
	for _, variant := range append(Variants(), "not-a-variant") {
		writeString(hash, string(variant))
		for _, codepoint := range codepoints {
			glyph, ok := Lookup(variant, codepoint)
			metrics, metricsOK := LookupMetrics(variant, codepoint)
			path, content, pathOK := LookupPath(variant, codepoint)
			remapped := RemapCodepoint(variant, codepoint)
			if metricsOK != ok || pathOK != ok || metrics != glyph.Metrics || path != glyph.Path || content != glyph.Content {
				panic("Lookup, LookupMetrics and LookupPath disagree")
			}
			if !ok && remapped == codepoint {
				if glyph != (Glyph{}) {
					panic("a miss returned a glyph")
				}
				continue
			}
			writeUint64(hash, uint64(uint32(codepoint)))
			writeBool(hash, ok)
			writeUint64(hash, uint64(uint32(glyph.Codepoint)))
			writeUint64(hash, variantNumber[glyph.Variant])
			if glyph.Variant != "" && variantNumber[glyph.Variant] == 0 {
				panic("resolved to an unknown variant")
			}
			for _, value := range [...]float64{metrics.Height, metrics.Depth, metrics.Width, metrics.ItalicCorrection, metrics.Skew} {
				writeUint64(hash, math.Float64bits(value))
			}
			writeBool(hash, metrics.HasItalicCorrection)
			writeBool(hash, metrics.HasSkew)
			writeString(hash, path)
			writeString(hash, content)
			writeUint64(hash, uint64(uint32(remapped)))
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// The digest was taken from the implementation that kept each variant in a
// map[rune]glyphData and the variant rules and alphanumeric bases in maps
// keyed by variant name.
func TestEffectiveLookupDigest(t *testing.T) {
	const expected = "e092a9365e593d46fc208948e48f0f1f6ba4976a264ddda303bbfd61ccf41a50"
	if got := effectiveLookupDigest(); got != expected {
		t.Fatalf("effective lookup digest = %s, want %s", got, expected)
	}
}
