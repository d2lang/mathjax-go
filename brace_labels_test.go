package mathjax

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/d2lang/mathjax-go/internal/oracle"
)

type braceLabelCase struct {
	Name      string
	TeX       string
	Display   bool
	MError    bool
	Error     string
	Width     int
	Height    int
	SVGSHA256 string
}

func braceLabelCases(t *testing.T) []braceLabelCase {
	t.Helper()
	data, err := os.ReadFile("testdata/brace-labels.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct{ Cases []braceLabelCase }
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest.Cases
}

// These hashes and dimensions come from D2's pinned MathJax 3.2.2 component,
// independently of the Go compiler and typesetter. Check the entire SVG so
// misplaced labels, missing brace parts and incorrect spacing cannot pass.
func TestBraceLabelRendering(t *testing.T) {
	for _, test := range braceLabelCases(t) {
		t.Run(fmt.Sprintf("%s/display=%t", test.Name, test.Display), func(t *testing.T) {
			options := DefaultOptions()
			options.Display = test.Display
			got, err := RenderWithOptions(test.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); hash != test.SVGSHA256 {
				t.Fatalf("SVG hash = %s, want pinned MathJax %s", hash, test.SVGSHA256)
			}
			if strings.Contains(got, `data-mml-node="merror"`) != test.MError {
				t.Fatalf("unexpected error rendering: %s", got)
			}
			if test.MError && !strings.Contains(got, `data-mjx-error="`+test.Error+`"`) {
				t.Fatalf("missing expected TeX error %q", test.Error)
			}
			if test.Display {
				width, height, err := Measure(test.TeX)
				if err != nil || width != test.Width || height != test.Height {
					t.Fatalf("Measure = %d x %d, %v; want %d x %d", width, height, err, test.Width, test.Height)
				}
			}
		})
	}
}

func TestBraceLabelFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to check the frozen brace-label references")
	}
	if err != nil {
		t.Fatal(err)
	}
	fixtures := braceLabelCases(t)
	cases := make([]oracle.Case, len(fixtures))
	for i, fixture := range fixtures {
		options := DefaultOptions()
		options.Display = fixture.Display
		cases[i] = oracle.Case{TeX: fixture.TeX, Options: options.pipelineOptions()}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	results, err := runner.RenderBatch(ctx, cases)
	if err != nil {
		t.Fatal(err)
	}
	for i, result := range results {
		fixture := fixtures[i]
		if result.Error != "" {
			t.Fatalf("%s: oracle error: %s", fixture.Name, result.Error)
		}
		if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(result.SVG))); hash != fixture.SVGSHA256 {
			t.Errorf("%s/display=%t: frozen oracle hash = %s, want %s", fixture.Name, fixture.Display, hash, fixture.SVGSHA256)
		}
	}
}
