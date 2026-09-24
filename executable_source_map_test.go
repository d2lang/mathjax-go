// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

// Public TeX errors are rendered as merror SVG. These references check the
// complete public output and message; error IDs are retained observation metadata.
func TestExecutableSourceMapPublicReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/executable_source_map_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG  string
			Display, Public bool
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 132 {
		t.Fatal("unbound executable source-map references")
	}
	count := 0
	for _, c := range fixture.Cases {
		if !c.Public {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Fatalf("complete primary SVG differs\ngot: %s\nwant: %s", got, c.SVG)
			}
		})
	}
	if count != 128 {
		t.Fatalf("public references %d; want 128", count)
	}
}

// The primary's error attribute contains a literal <. Retain that reference,
// but require XML-safe public output and its exact decoded message, not raw SVG
// equality or a normalized primary string.
func TestEscapedDelimiterLessThanXML(t *testing.T) {
	data, err := os.ReadFile("testdata/escaped_delimiter_xml_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, Qualification, PrimarySVG, PrimaryErrorID string
			ExpectedMessage, ExpectedText                        string
			Display                                              bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 2 {
		t.Fatal("unbound escaped less-than XML references")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.TeX != `\<` || c.PrimaryErrorID != "UndefinedControlSequence" ||
				c.ExpectedMessage != "Undefined control sequence "+c.TeX || c.ExpectedText != c.ExpectedMessage ||
				c.Qualification != "XML-safe Go error; not a whole-primary-SVG reference" ||
				!strings.Contains(c.PrimarySVG, `data-mjx-error="`+c.ExpectedMessage+`"`) {
				t.Fatal("unbound primary error qualification")
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			decoder := xml.NewDecoder(strings.NewReader(got))
			depth, roots, errors, texts := 0, 0, 0, 0
			errorDepth, textDepth := 0, 0
			var message string
			var text strings.Builder
			for {
				token, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("public SVG is not valid XML: %v", err)
				}
				switch token := token.(type) {
				case xml.StartElement:
					depth++
					if depth == 1 {
						roots++
						if token.Name.Local != "svg" || token.Name.Space != "http://www.w3.org/2000/svg" {
							t.Fatalf("unexpected XML root: %v", token.Name)
						}
					}
					var kind, errorMessage string
					for _, attr := range token.Attr {
						if attr.Name.Local == "data-mml-node" {
							kind = attr.Value
						}
						if attr.Name.Local == "data-mjx-error" {
							errorMessage = attr.Value
						}
					}
					if token.Name.Local == "g" && kind == "merror" {
						errors++
						errorDepth, message = depth, errorMessage
					}
					if errorDepth > 0 && token.Name.Local == "text" {
						texts++
						textDepth = depth
					}
				case xml.CharData:
					if depth == 0 && strings.TrimSpace(string(token)) != "" {
						t.Fatal("unexpected text outside SVG root")
					}
					if textDepth > 0 {
						text.Write([]byte(token))
					}
				case xml.EndElement:
					if depth == textDepth {
						textDepth = 0
					}
					if depth == errorDepth {
						errorDepth = 0
					}
					depth--
				}
			}
			if roots != 1 || depth != 0 {
				t.Fatalf("XML roots/depth = %d/%d; want 1/0", roots, depth)
			}
			if errors != 1 {
				t.Fatalf("XML merror count = %d; want 1", errors)
			}
			if message != c.ExpectedMessage || texts != 1 || text.String() != c.ExpectedText {
				t.Fatalf("decoded error = %q, text count = %d, text = %q; want %q", message, texts, text.String(), c.ExpectedMessage)
			}
		})
	}
}
