package font

import (
	"os"
	"testing"
)

// The embedded stream decompresses to exactly the records that generate.py
// wrote.
func TestGlyphOutlineStream(t *testing.T) {
	records, err := os.ReadFile("data_outlines_gen.bin")
	if err != nil {
		t.Fatal(err)
	}
	if glyphOutlineData() != string(records) {
		t.Fatal("data_outlines_gen.br does not decompress to data_outlines_gen.bin; run go generate")
	}
}
