package svg

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
)

type positionRoundingFixture struct {
	Multiscripts []struct {
		Name, Alignment            string
		Width, Scale, Column, Want float64
	}
	Roots []struct {
		Name       string
		Root, Surd layout.BBox
		SurdSize   int
		Height     float64
		Want       [3]float64
	}
}

func loadPositionRoundingFixture(t *testing.T) positionRoundingFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/position_rounding_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture positionRoundingFixture
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Multiscripts) != 5 || len(fixture.Roots) != 1 {
		t.Fatal("rounding fixture counts")
	}
	return fixture
}
func requirePositionRoundingPrimary(t *testing.T, got, want []float64) {
	t.Helper()
	gotBits, wantBits := []string{}, []string{}
	equal := len(got) == len(want)
	for _, v := range got {
		gotBits = append(gotBits, fmt.Sprintf("%016x", math.Float64bits(v)))
	}
	for i, v := range want {
		wantBits = append(wantBits, fmt.Sprintf("%016x", math.Float64bits(v)))
		if i >= len(got) || math.Float64bits(got[i]) != math.Float64bits(v) {
			equal = false
		}
	}
	data, err := json.Marshal(map[string]any{"got": got, "want": want, "gotBits": gotBits, "wantBits": wantBits})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("D116_NUMERIC_OBSERVATION %s", data)
	if !equal {
		t.Fatalf("D116_PRIMARY_NUMERIC_MISMATCH %s", data)
	}
}
func TestMultiscriptAlignmentRoundingPrimary(t *testing.T) {
	fixture := loadPositionRoundingFixture(t)
	for _, c := range fixture.Multiscripts {
		t.Run(c.Name, func(t *testing.T) {
			// Exercise the actual inlinable helper with a dynamic scaled-width argument.
			got := multiscriptAlign(c.Alignment, c.Width*c.Scale, c.Column)
			requirePositionRoundingPrimary(t, []float64{got}, []float64{c.Want})
		})
	}
}
func TestRootDimensionsRoundingPrimary(t *testing.T) {
	fixture := loadPositionRoundingFixture(t)
	for _, c := range fixture.Roots {
		t.Run(c.Name, func(t *testing.T) {
			// Construct the actual wrapper, then supply the primary's measured boxes
			// to isolate rootDimensions from the separate script-scale policy.
			base := wrapperTrancheToken("mi", "x", mml.TeXClassOrd)
			index := wrapperTrancheToken("mn", "3", mml.TeXClassOrd)
			w := wrapperTrancheWrapper(mml.NewNode("mroot", nil, nil, base, index))
			w.children[w.rootIndex()].bbox = c.Root.Clone()
			w.children[w.rootIndex()].bboxComputed = true
			w.children[w.surdIndex()].size = c.SurdSize
			x, h, dx := w.rootDimensions(&c.Surd, c.Height)
			got := [3]float64{x, h, dx}
			requirePositionRoundingPrimary(t, got[:], c.Want[:])
		})
	}
}
