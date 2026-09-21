package svg

import (
	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/mml"
	"testing"
)

func TestExplicitStretchCharacterBypassesAccentRemap(t *testing.T) {
	for _, c := range []struct {
		name     string
		stretch  bool
		alias    rune
		hasAlias bool
		size     int
		sizeSet  bool
		chars    []rune
		want     string
	}{
		{"ordinary accent", false, 0, false, 0, false, nil, "⃗"},
		{"original stretch character only", true, 0, false, 0, false, nil, "⃗"},
		{"explicit alias", true, '→', true, 0, false, nil, "→"},
		{"zero alias is absent", true, 0, true, 0, false, nil, "⃗"},
		{"selected size character", true, 0, false, 1, true, []rune{0, '→'}, "→"},
		{"unselected size character", true, 0, false, 1, false, []rune{0, '→'}, "⃗"},
		{"multipart is not a selected character", true, 0, false, -1, true, []rune{'→'}, "⃗"},
		{"zero selected character", true, 0, false, 0, true, []rune{0}, "⃗"},
	} {
		t.Run(c.name, func(t *testing.T) {
			n := mml.NewNode("mo", nil, nil)
			n.SetProperty("mathaccent", true)
			// A real implicit accent stack is distinct from the mathaccent
			// positioning marker. The pinned registered-MML corpus covers this
			// state and the marker-only negative separately.
			n.Attributes.SetInherited("accent", true)
			stackNode := mml.NewNode("mover", nil, nil)
			stack := &wrapper{node: stackNode}
			parent := &wrapper{node: n, hasStretch: c.stretch, stretch: font.Delimiter{HasAlias: c.hasAlias, Alias: c.alias, SizeChars: c.chars}, size: c.size, sizeSet: c.sizeSet, stretchGlyph: '→', parent: stack}
			stack.children = []*wrapper{{node: mml.NewNode("mi", nil, nil)}, parent}
			if got := remapAccentText(parent, "→"); got != c.want {
				t.Errorf("remap %q, want %q", got, c.want)
			}
		})
	}
}
