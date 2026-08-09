// Copyright (c) 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type mjOperatorManifestEntry struct {
	Operator      string `json:"operator"`
	Name          string `json:"name"`
	SourceLine    int    `json:"source_line"`
	Reference     string `json:"reference"`
	PayloadSHA256 string `json:"payload_sha256"`
}

type mjOperatorManifestRange struct {
	SourceLine    int    `json:"source_line"`
	First         int    `json:"first"`
	Last          int    `json:"last"`
	TexClass      int    `json:"tex_class"`
	Kind          string `json:"kind"`
	Variant       string `json:"variant"`
	HasVariant    bool   `json:"has_variant"`
	PayloadSHA256 string `json:"payload_sha256"`
}

type mjOperatorManifestSpacing struct {
	TexClass      int    `json:"tex_class"`
	SourceLine    int    `json:"source_line"`
	Lspace        int    `json:"lspace"`
	Rspace        int    `json:"rspace"`
	PayloadSHA256 string `json:"payload_sha256"`
}

type mjOperatorManifestMutation struct {
	Form                  string `json:"form"`
	Operator              string `json:"operator"`
	SourceLine            int    `json:"source_line"`
	Reference             string `json:"reference"`
	PayloadSHA256         string `json:"payload_sha256"`
	HadPrevious           bool   `json:"had_previous"`
	PreviousSourceLine    int    `json:"previous_source_line"`
	PreviousReference     string `json:"previous_reference"`
	PreviousPayloadSHA256 string `json:"previous_payload_sha256"`
}

type mjOperatorManifest struct {
	Schema   int `json:"schema"`
	Upstream struct {
		Project      string `json:"project"`
		Version      string `json:"version"`
		GitCommit    string `json:"git_commit"`
		Repository   string `json:"repository"`
		Source       string `json:"source"`
		SourceSHA256 string `json:"source_sha256"`
	} `json:"upstream"`
	CanonicalPayload string                      `json:"canonical_payload"`
	MOCount          int                         `json:"mo_count"`
	RangeCount       int                         `json:"range_count"`
	SpacingCount     int                         `json:"spacing_count"`
	PrefixCount      int                         `json:"prefix_count"`
	PostfixCount     int                         `json:"postfix_count"`
	InfixCount       int                         `json:"infix_count"`
	OperatorCount    int                         `json:"operator_count"`
	MutationCount    int                         `json:"mutation_count"`
	MO               []mjOperatorManifestEntry   `json:"mo"`
	Ranges           []mjOperatorManifestRange   `json:"ranges"`
	Spacing          []mjOperatorManifestSpacing `json:"spacing"`
	Forms            struct {
		Prefix  []mjOperatorManifestEntry `json:"prefix"`
		Postfix []mjOperatorManifestEntry `json:"postfix"`
		Infix   []mjOperatorManifestEntry `json:"infix"`
	} `json:"forms"`
	Mutations []mjOperatorManifestMutation `json:"mutations"`
}

func TestMJOperatorDictionaryPinnedManifest(t *testing.T) {
	data, err := os.ReadFile("testdata/operator_dictionary_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest mjOperatorManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse operator manifest: %v", err)
	}
	if manifest.Schema != 1 || manifest.Upstream.Project != "MathJax-src" ||
		manifest.Upstream.Version != "3.2.2" ||
		manifest.Upstream.GitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" ||
		manifest.Upstream.Repository != "https://github.com/mathjax/MathJax-src" ||
		manifest.Upstream.Source != "ts/core/MmlTree/OperatorDictionary.ts" ||
		manifest.Upstream.SourceSHA256 != "c771c8fff8dacbb37ca2d546d7f1ea5cbe520504d2838e2b18ad2adbc2ca874f" {
		t.Fatalf("unexpected pinned operator source: schema=%d upstream=%+v", manifest.Schema, manifest.Upstream)
	}
	if manifest.CanonicalPayload == "" {
		t.Fatal("manifest canonical_payload description is empty")
	}
	if manifest.MOCount != 28 || manifest.RangeCount != 50 || manifest.SpacingCount != 7 ||
		manifest.PrefixCount != 97 || manifest.PostfixCount != 126 || manifest.InfixCount != 928 ||
		manifest.OperatorCount != 1151 || manifest.MutationCount != 3 {
		t.Fatalf("unexpected manifest counts: MO=%d ranges=%d spacing=%d prefix=%d postfix=%d infix=%d operators=%d mutations=%d",
			manifest.MOCount, manifest.RangeCount, manifest.SpacingCount,
			manifest.PrefixCount, manifest.PostfixCount, manifest.InfixCount,
			manifest.OperatorCount, manifest.MutationCount)
	}

	if len(manifest.MO) != len(mjOperatorDefinitions) {
		t.Fatalf("manifest MO definitions = %d, Go = %d", len(manifest.MO), len(mjOperatorDefinitions))
	}
	for i, want := range manifest.MO {
		got := mjOperatorDefinitions[i]
		if got.Name != want.Name || got.SourceLine != want.SourceLine || got.Reference != want.Reference {
			t.Errorf("MO[%d] metadata = %#v, manifest = %#v", i, got, want)
		}
		if hash := mjOperatorPayloadHash(t, got.Definition); hash != want.PayloadSHA256 {
			t.Errorf("MO %q payload hash = %s, manifest = %s", got.Name, hash, want.PayloadSHA256)
		}
		validateMJOperatorDefinition(t, "MO."+got.Name, got.Definition)
	}

	if len(manifest.Ranges) != len(mjOperatorRanges) {
		t.Fatalf("manifest ranges = %d, Go = %d", len(manifest.Ranges), len(mjOperatorRanges))
	}
	for i, want := range manifest.Ranges {
		got := mjOperatorRanges[i]
		if got.SourceLine != want.SourceLine || got.First != want.First || got.Last != want.Last ||
			got.TexClass != want.TexClass || got.Kind != want.Kind ||
			got.Variant != want.Variant || got.HasVariant != want.HasVariant {
			t.Errorf("range[%d] = %#v, manifest = %#v", i, got, want)
		}
		if hash := mjOperatorRangePayloadHash(t, got); hash != want.PayloadSHA256 {
			t.Errorf("range[%d] payload hash = %s, manifest = %s", i, hash, want.PayloadSHA256)
		}
		if got.First > got.Last {
			t.Errorf("range[%d] has descending bounds %#x..%#x", i, got.First, got.Last)
		}
		if i > 0 && mjOperatorRanges[i-1].Last >= got.First {
			t.Errorf("range[%d] overlaps or is out of order", i)
		}
	}

	if len(manifest.Spacing) != len(mjOperatorMMLSpacing) {
		t.Fatalf("manifest spacing = %d, Go = %d", len(manifest.Spacing), len(mjOperatorMMLSpacing))
	}
	for i, want := range manifest.Spacing {
		got := mjOperatorMMLSpacing[i]
		if got.TexClass != i || got.TexClass != want.TexClass || got.SourceLine != want.SourceLine ||
			got.Lspace != want.Lspace || got.Rspace != want.Rspace {
			t.Errorf("spacing[%d] = %#v, manifest = %#v", i, got, want)
		}
		if hash := mjOperatorPayloadHash(t, []any{got.Lspace, got.Rspace}); hash != want.PayloadSHA256 {
			t.Errorf("spacing[%d] payload hash = %s, manifest = %s", i, hash, want.PayloadSHA256)
		}
	}

	compareMJOperatorForm(t, "prefix", mjOperatorDictionaryPrefix, manifest.Forms.Prefix)
	compareMJOperatorForm(t, "postfix", mjOperatorDictionaryPostfix, manifest.Forms.Postfix)
	compareMJOperatorForm(t, "infix", mjOperatorDictionaryInfix, manifest.Forms.Infix)

	if len(manifest.Mutations) != len(mjOperatorDictionaryMutations) {
		t.Fatalf("manifest mutations = %d, Go = %d", len(manifest.Mutations), len(mjOperatorDictionaryMutations))
	}
	for i, want := range manifest.Mutations {
		got := mjOperatorDictionaryMutations[i]
		if got.Form != want.Form || got.Operator != want.Operator || got.SourceLine != want.SourceLine ||
			got.Reference != want.Reference || got.HadPrevious != want.HadPrevious ||
			got.PreviousSourceLine != want.PreviousSourceLine || got.PreviousReference != want.PreviousReference {
			t.Errorf("mutation[%d] = %#v, manifest = %#v", i, got, want)
		}
		if hash := mjOperatorPayloadHash(t, got.Definition); hash != want.PayloadSHA256 {
			t.Errorf("mutation[%d] payload hash = %s, manifest = %s", i, hash, want.PayloadSHA256)
		}
		if got.HadPrevious {
			if hash := mjOperatorPayloadHash(t, got.PreviousDefinition); hash != want.PreviousPayloadSHA256 {
				t.Errorf("mutation[%d] previous payload hash = %s, manifest = %s", i, hash, want.PreviousPayloadSHA256)
			}
		}
	}
}

func TestMJOperatorDictionaryRepresentativeValues(t *testing.T) {
	op := findMJOperatorDefinition(t, "OP")
	wantOP := mjOperatorDefinition{
		Lspace: 1, Rspace: 2, TexClass: 1,
		Properties: mjOperatorProperties{
			{Name: "largeop", Value: true},
			{Name: "movablelimits", Value: true},
			{Name: "symmetric", Value: true},
		},
	}
	if !reflect.DeepEqual(op, wantOP) {
		t.Errorf("MO.OP = %#v, want %#v", op, wantOP)
	}
	sum := findMJOperatorEntry(t, mjOperatorDictionaryPrefix, "∑")
	if sum.Reference != "MO.OP" || !reflect.DeepEqual(sum.Definition, wantOP) {
		t.Errorf("prefix summation = %#v", sum)
	}
	factorial := findMJOperatorEntry(t, mjOperatorDictionaryPostfix, "!")
	if factorial.Definition.Lspace != 1 || factorial.Definition.Rspace != 0 || factorial.Definition.TexClass != 5 {
		t.Errorf("postfix factorial = %#v", factorial)
	}
	widehat := findMJOperatorEntry(t, mjOperatorDictionaryInfix, "^")
	if widehat.SourceLine != 1333 || widehat.Reference != "MO.WIDEREL" ||
		widehat.Definition.TexClass != 3 || len(widehat.Definition.Properties) != 2 {
		t.Errorf("final infix circumflex override = %#v", widehat)
	}
	if got := mjOperatorRanges[len(mjOperatorRanges)-1]; got.Variant != "normnal" || !got.HasVariant {
		t.Errorf("last range must preserve upstream variant spelling exactly: %#v", got)
	}
	wantMutations := []struct {
		operator    string
		hadPrevious bool
		line        int
	}{{"^", true, 1333}, {"_", true, 1334}, {"⫝̸", false, 1339}}
	for i, want := range wantMutations {
		got := mjOperatorDictionaryMutations[i]
		if got.Operator != want.operator || got.HadPrevious != want.hadPrevious || got.SourceLine != want.line {
			t.Errorf("mutation[%d] = %#v, want %+v", i, got, want)
		}
	}
}

func compareMJOperatorForm(t *testing.T, form string, got []mjOperatorDictionaryEntry, want []mjOperatorManifestEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s operators = %d, manifest = %d", form, len(got), len(want))
	}
	names := make(map[string]bool, len(got))
	for i, wantEntry := range want {
		gotEntry := got[i]
		if names[gotEntry.Operator] {
			t.Errorf("%s has duplicate operator %q", form, gotEntry.Operator)
		}
		names[gotEntry.Operator] = true
		if gotEntry.Operator != wantEntry.Operator || gotEntry.SourceLine != wantEntry.SourceLine ||
			gotEntry.Reference != wantEntry.Reference {
			t.Errorf("%s[%d] metadata = %#v, manifest = %#v", form, i, gotEntry, wantEntry)
		}
		if hash := mjOperatorPayloadHash(t, gotEntry.Definition); hash != wantEntry.PayloadSHA256 {
			t.Errorf("%s %q payload hash = %s, manifest = %s", form, gotEntry.Operator, hash, wantEntry.PayloadSHA256)
		}
		validateMJOperatorDefinition(t, form+"."+gotEntry.Operator, gotEntry.Definition)
	}
}

func validateMJOperatorDefinition(t *testing.T, label string, definition mjOperatorDefinition) {
	t.Helper()
	names := make(map[string]bool, len(definition.Properties))
	for _, property := range definition.Properties {
		if names[property.Name] {
			t.Errorf("%s has duplicate property %q", label, property.Name)
		}
		names[property.Name] = true
		switch property.Value.(type) {
		case bool, string:
		default:
			t.Errorf("%s property %q has unsupported type %T", label, property.Name, property.Value)
		}
	}
}

func findMJOperatorDefinition(t *testing.T, name string) mjOperatorDefinition {
	t.Helper()
	for _, definition := range mjOperatorDefinitions {
		if definition.Name == name {
			return definition.Definition
		}
	}
	t.Fatalf("operator definition %q not found", name)
	return mjOperatorDefinition{}
}

func findMJOperatorEntry(t *testing.T, entries []mjOperatorDictionaryEntry, operator string) mjOperatorDictionaryEntry {
	t.Helper()
	for _, entry := range entries {
		if entry.Operator == operator {
			return entry
		}
	}
	t.Fatalf("operator %q not found", operator)
	return mjOperatorDictionaryEntry{}
}

type mjOperatorCanonicalObject struct {
	Properties [][2]any `json:"$object"`
}

func mjOperatorRangePayloadHash(t *testing.T, value mjOperatorRange) string {
	t.Helper()
	payload := []any{value.First, value.Last, value.TexClass, value.Kind}
	if value.HasVariant {
		payload = append(payload, value.Variant)
	}
	return mjOperatorPayloadHash(t, payload)
}

func mjOperatorPayloadHash(t *testing.T, value any) string {
	t.Helper()
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(mjCanonicalOperatorValue(t, value)); err != nil {
		t.Fatalf("encode canonical operator value: %v", err)
	}
	encoded := bytes.TrimSuffix(buffer.Bytes(), []byte{'\n'})
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func mjCanonicalOperatorValue(t *testing.T, value any) any {
	t.Helper()
	switch value := value.(type) {
	case nil, string, bool, int:
		return value
	case mjOperatorDefinition:
		return []any{
			value.Lspace,
			value.Rspace,
			value.TexClass,
			mjCanonicalOperatorValue(t, value.Properties),
		}
	case mjOperatorProperties:
		if value == nil {
			return nil
		}
		result := mjOperatorCanonicalObject{Properties: make([][2]any, len(value))}
		for i, property := range value {
			result.Properties[i] = [2]any{property.Name, mjCanonicalOperatorValue(t, property.Value)}
		}
		return result
	case []any:
		result := make([]any, len(value))
		for i, item := range value {
			result[i] = mjCanonicalOperatorValue(t, item)
		}
		return result
	default:
		t.Fatalf("unsupported canonical operator value %T", value)
		return nil
	}
}
