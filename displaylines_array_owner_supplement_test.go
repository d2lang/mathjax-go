// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestDisplaylinesArrayOwnerSupplement(t *testing.T) {
	data, err := os.ReadFile("testdata/displaylines_array_owner_supplement.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "ad47d670d80fbbd7e17fa5fc506f0f292a469e3ba9226f6dbf8dcd4ba143d43b" {
		t.Fatal("changed complete source supplement")
	}
	var supplement struct {
		MathjaxGitCommit, ActualMergedBase string
		Cases                              []struct {
			Name, TeX, Partition, QualifiedChange, Group, PublicationCategory string
			Index                                                             int
			Display                                                           *bool
			Original                                                          map[string]string
			HistoricalSourceMemberships                                       []struct{ CompleteSourceRecord json.RawMessage }
			ActualPublishedReferences                                         []struct {
				Source                      struct{ Path, GitBlob, SHA256 string }
				JSONPointer, Classification string
				Record                      json.RawMessage
			}
		}
	}
	if err := json.Unmarshal(data, &supplement); err != nil {
		t.Fatal(err)
	}
	if supplement.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || supplement.ActualMergedBase != "b745ce2b28400d18cd98fb7d14cd4f4280e936de" || len(supplement.Cases) != 74 {
		t.Fatal("unbound supplement profile")
	}
	baseData, err := os.ReadFile("testdata/displaylines_array_owner_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var base struct {
		Cases []displaylinesArrayOwnerReference
	}
	if err := json.Unmarshal(baseData, &base); err != nil {
		t.Fatal(err)
	}
	inputs := map[string]bool{}
	for _, c := range base.Cases {
		if c.Display == nil {
			t.Fatal("missing original116 display")
		}
		key := fmt.Sprintf("%t:%s", *c.Display, c.TeX)
		if inputs[key] {
			t.Fatal("duplicate original116 input")
		}
		inputs[key] = true
	}
	if len(inputs) != 116 {
		t.Fatal("changed original116 inventory")
	}
	cache := map[string][]json.RawMessage{}
	readActual := func(name, hash string) []json.RawMessage {
		t.Helper()
		if rows, ok := cache[name]; ok {
			return rows
		}
		switch name {
		case "testdata/ordinary_array_owner_residuals.json.gz", "testdata/ordinary_array_owner_repair_observations.json.gz", "testdata/hfill_residuals.json":
		default:
			t.Fatal("unexpected prior source", name)
		}
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatal("prior source bytes changed", name)
		}
		if strings.HasSuffix(name, ".gz") {
			z, err := gzip.NewReader(bytes.NewReader(b))
			if err != nil {
				t.Fatal(err)
			}
			b, err = io.ReadAll(z)
			z.Close()
			if err != nil {
				t.Fatal(err)
			}
		}
		var doc struct{ Cases []json.RawMessage }
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatal(err)
		}
		cache[name] = doc.Cases
		return doc.Cases
	}
	decode := func(raw json.RawMessage) any {
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	groups, publication := map[string]int{}, map[string]int{}
	indices, names := map[int]bool{}, map[string]bool{}
	actualReferences, sourceMemberships := 0, 0
	for _, c := range supplement.Cases {
		if c.Display == nil || c.Name == "" || c.Index < 0 || c.Index >= 74 || indices[c.Index] || names[c.Name] {
			t.Fatal("unbound supplement input")
		}
		indices[c.Index], names[c.Name] = true, true
		key := fmt.Sprintf("%t:%s", *c.Display, c.TeX)
		if inputs[key] {
			t.Fatal("supplement overlaps another source input")
		}
		inputs[key] = true
		if c.Partition != "strict-svg" || c.QualifiedChange != "fix" || len(c.Original) != 1 || c.Original["svg"] == "" || strings.Contains(c.Original["svg"], "data-mjx-error=") {
			t.Fatal("changed complete non-error original")
		}
		displaylinesArrayOwnerSVG(t, c.Original["svg"])
		if len(c.HistoricalSourceMemberships) != 1 {
			t.Fatal("lost historical source membership")
		}
		for _, m := range c.HistoricalSourceMemberships {
			var source struct {
				TeX      string
				Display  *bool
				Original map[string]string
			}
			if err := json.Unmarshal(m.CompleteSourceRecord, &source); err != nil {
				t.Fatal(err)
			}
			if source.Display == nil || source.TeX != c.TeX || *source.Display != *c.Display || !reflect.DeepEqual(source.Original, c.Original) {
				t.Fatal("historical original changed")
			}
			sourceMemberships++
		}
		primaryReferences := 0
		for _, r := range c.ActualPublishedReferences {
			rows := readActual(r.Source.Path, r.Source.SHA256)
			var index int
			if _, err := fmt.Sscanf(r.JSONPointer, "/cases/%d", &index); err != nil || r.JSONPointer != fmt.Sprintf("/cases/%d", index) || index < 0 || index >= len(rows) {
				t.Fatal("unbound actual source pointer")
			}
			if !reflect.DeepEqual(decode(rows[index]), decode(r.Record)) {
				t.Fatal("complete actual source record changed", r.Source.Path, index)
			}
			var source struct {
				TeX, Partition string
				Display        *bool
				Original       map[string]string
			}
			if err := json.Unmarshal(r.Record, &source); err != nil {
				t.Fatal(err)
			}
			if source.Display == nil || source.TeX != c.TeX || *source.Display != *c.Display || !reflect.DeepEqual(source.Original, c.Original) {
				t.Fatal("actual original changed")
			}
			switch r.Classification {
			case "held-full-original-no-render-assertion":
				if r.Source.Path != "testdata/ordinary_array_owner_residuals.json.gz" || source.Partition != "held-svg" || c.PublicationCategory != "priorHeldCompleteOriginal" {
					t.Fatal("changed held promotion")
				}
				primaryReferences++
			case "raw-original-no-active-Go-consumer":
				if r.Source.Path != "testdata/hfill_residuals.json" || c.PublicationCategory != "priorRawCompleteOriginal" {
					t.Fatal("changed raw promotion")
				}
				primaryReferences++
			case "passive-observation-no-active-Go-consumer":
				if r.Source.Path != "testdata/ordinary_array_owner_repair_observations.json.gz" {
					t.Fatal("changed passive reference")
				}
			default:
				t.Fatal("unknown source reference classification")
			}
			actualReferences++
		}
		if primaryReferences != 1 {
			t.Fatal("missing primary prior source")
		}
		groups[c.Group]++
		publication[c.PublicationCategory]++
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = *c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.Original["svg"] {
				t.Fatalf("complete original SVG differs: got %x want %x", sha256.Sum256([]byte(got)), sha256.Sum256([]byte(c.Original["svg"])))
			}
		})
	}
	if !reflect.DeepEqual(groups, map[string]int{"nested-array-center": 64, "cell-split-and-hfill": 10}) || !reflect.DeepEqual(publication, map[string]int{"priorHeldCompleteOriginal": 64, "priorRawCompleteOriginal": 10}) || actualReferences != 82 || sourceMemberships != 74 || len(inputs) != 190 {
		t.Fatal("changed supplement accounting", groups, publication, actualReferences, sourceMemberships, len(inputs))
	}
}
