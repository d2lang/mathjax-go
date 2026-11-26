//go:build ignore

// gen_outline_stream writes data_outlines_gen.br, data_outlines_gen.bin
// compressed with brotli, which outline.go embeds. Run go generate in this
// directory after generate.py rewrites data_outlines_gen.bin;
// TestGlyphOutlineStream fails until the two match again.
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/d2lang/brotli"
)

func main() {
	records, err := os.ReadFile("data_outlines_gen.bin")
	if err != nil {
		log.Fatal(err)
	}
	// The window is the smallest that spans the records, so that a record can
	// refer to any earlier one and the decoder's ring buffer stays as small
	// as that allows.
	lgwin := 10
	for (1<<lgwin)-16 < len(records) {
		lgwin++
	}
	if lgwin > 24 {
		log.Fatalf("%d bytes of outlines do not fit a brotli window", len(records))
	}
	var stream bytes.Buffer
	w := brotli.NewWriterOptions(&stream, brotli.WriterOptions{Quality: 11, LGWin: lgwin})
	if _, err := w.Write(records); err != nil {
		log.Fatal(err)
	}
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("data_outlines_gen.br", stream.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("data_outlines_gen.br: %d bytes of outlines, %d compressed (window 2^%d)\n", len(records), stream.Len(), lgwin)
}
