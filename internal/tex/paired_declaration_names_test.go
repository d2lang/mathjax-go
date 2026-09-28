// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestPairedDeclarationNamesPrimaryMethod(t *testing.T) {
	data, err := os.ReadFile("testdata/paired_declaration_names_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Methods          map[string]string
		Cases            []struct {
			Raw, Method   string
			ArgumentsRead []string
			Added         []struct{ Key string }
			Error         *struct{ ID, Message string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 149 ||
		!strings.Contains(fixture.Methods["DeclarePairedDelimiter"], "GetCsNameArgument") ||
		!strings.Contains(fixture.Methods["trimSpaces"], ".trim()") {
		t.Fatal("unbound original paired-name observations")
	}
	seen := make(map[string]bool)
	for i, c := range fixture.Cases {
		if c.Method != "paired" || seen[c.Raw] {
			t.Fatal("invalid or duplicate original observation", c)
		}
		seen[c.Raw] = true
		t.Run(fmt.Sprintf("case-%03d", i), func(t *testing.T) {
			got, err := validatedCSNameArgument(c.Raw, "DeclarePairedDelimiter")
			if c.Error != nil {
				gotError, ok := err.(*Error)
				if !ok || gotError.ID != c.Error.ID || gotError.Message != c.Error.Message || got != "" || len(c.Added) != 0 || len(c.ArgumentsRead) != 1 {
					t.Fatalf("original rejection differs: got %q, %v; want %+v", got, err, c.Error)
				}
			} else if err != nil || len(c.Added) != 1 || got != c.Added[0].Key || len(c.ArgumentsRead) != 3 {
				t.Fatalf("original name differs: got %q, %v; want %+v", got, err, c.Added)
			}
		})
	}
}
