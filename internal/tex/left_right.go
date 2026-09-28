// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on BaseMethods.LeftRight and BaseItems.LeftItem (MathJax 3.2.2).
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// A delimiter closes every intervening row continuation. Keep its kind and
// captured color until the owning LeftItem receives it, including across an
// OverItem denominator or a SetFont continuation.
type rowDelimiterItem struct {
	middle           bool
	delimiter, color string
}

func (p *parser) leftRight(name string) ([]*mml.Node, error) {
	open, err := p.readDelimiter(name, false)
	if err != nil {
		return nil, err
	}
	outerDelimiter := p.rowDelimiter
	defer func() { p.rowDelimiter = outerDelimiter }()
	var children []*mml.Node
	for {
		p.rowDelimiter = nil
		children, _, err = p.parseRowWithPrefix(0, true, children)
		if err != nil {
			return nil, err
		}
		item := p.rowDelimiter
		if !item.middle {
			fenced := p.leftRightFenced(open, row(children, true), item.delimiter, true)
			if item.color != "" {
				fenced.Children[len(fenced.Children)-1].Attributes.Set("mathcolor", item.color)
			}
			return []*mml.Node{fenced}, nil
		}
		// Middle closes styles and restores the environment at Left entry.
		// Its empty CLOSE/OPEN atoms retain the original TeX spacing, while
		// the middle mo has no explicit texClass (the dictionary decides it).
		middle := p.token("mo", item.delimiter)
		middle.Attributes.Set("stretchy", true)
		if item.color != "" {
			middle.Attributes.Set("mathcolor", item.color)
		}
		children = append(children,
			texAtom(forcedRow(nil, true), mml.TeXClassClose),
			middle,
			texAtom(forcedRow(nil, true), mml.TeXClassOpen),
		)
	}
}
