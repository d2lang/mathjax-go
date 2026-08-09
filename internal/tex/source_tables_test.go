// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"testing"
)

type mjSourceManifest struct {
	Schema   int `json:"schema"`
	Upstream struct {
		Project    string `json:"project"`
		Version    string `json:"version"`
		GitCommit  string `json:"git_commit"`
		Repository string `json:"repository"`
	} `json:"upstream"`
	CanonicalPayload string `json:"canonical_payload"`
	MapCount         int    `json:"map_count"`
	EntryCount       int    `json:"entry_count"`
	DynamicMapCount  int    `json:"dynamic_map_count"`
	PackageCount     int    `json:"package_count"`
	D2PackageClosure struct {
		D2GitCommit           string   `json:"d2_git_commit"`
		SelectionSource       string   `json:"selection_source"`
		SelectionSourceSHA256 string   `json:"selection_source_sha256"`
		DependencySource      string   `json:"dependency_source"`
		DependencySourceHash  string   `json:"dependency_source_sha256"`
		SelectedPackageOrder  []string `json:"selected_package_order"`
		DependencyOnly        []string `json:"dependency_only_packages"`
	} `json:"d2_package_closure"`
	Maps []struct {
		Source        string `json:"source"`
		Line          int    `json:"line"`
		Name          string `json:"name"`
		Kind          string `json:"kind"`
		Parser        string `json:"parser"`
		Methods       string `json:"methods"`
		Pattern       string `json:"pattern"`
		Flags         string `json:"flags"`
		Dynamic       bool   `json:"dynamic"`
		Priority      int    `json:"priority"`
		PayloadSHA256 string `json:"payload_sha256"`
		Entries       []struct {
			Name          string `json:"name"`
			PayloadSHA256 string `json:"payload_sha256"`
		} `json:"entries"`
	} `json:"maps"`
	Packages []struct {
		Source               string                  `json:"source"`
		Line                 int                     `json:"line"`
		Name                 string                  `json:"name"`
		Handlers             []mjSourceHandler       `json:"handlers"`
		Fallbacks            []mjSourceReference     `json:"fallbacks"`
		Items                []mjSourceReference     `json:"items"`
		Tags                 []mjSourceReference     `json:"tags"`
		OptionsPayloadSHA256 string                  `json:"options_payload_sha256"`
		DynamicMapNames      []string                `json:"dynamic_map_names"`
		DynamicHandlers      []mjSourceHandler       `json:"dynamic_handlers"`
		DynamicPriority      int                     `json:"dynamic_priority"`
		Hooks                []mjSourceHook          `json:"hooks"`
		Postprocessors       []mjSourcePostprocessor `json:"postprocessors"`
	} `json:"packages"`
}

func TestMJSourcePinnedManifest(t *testing.T) {
	data, err := os.ReadFile("testdata/source_tables_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest mjSourceManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse source manifest: %v", err)
	}
	if manifest.Schema != 2 || manifest.Upstream.Project != "MathJax-src" ||
		manifest.Upstream.Version != "3.2.2" ||
		manifest.Upstream.GitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" ||
		manifest.Upstream.Repository != "https://github.com/mathjax/MathJax-src" {
		t.Fatalf("unexpected pinned upstream: schema=%d upstream=%+v", manifest.Schema, manifest.Upstream)
	}
	if manifest.CanonicalPayload == "" {
		t.Fatal("manifest canonical_payload description is empty")
	}
	if manifest.MapCount != 56 || manifest.EntryCount != 1273 ||
		manifest.DynamicMapCount != 5 || manifest.PackageCount != 14 {
		t.Fatalf("manifest counts = %d maps/%d entries/%d dynamic/%d packages, want 56/1273/5/14",
			manifest.MapCount, manifest.EntryCount, manifest.DynamicMapCount, manifest.PackageCount)
	}
	if len(manifest.Maps) != len(mjSourceMaps) {
		t.Fatalf("manifest maps = %d, Go maps = %d", len(manifest.Maps), len(mjSourceMaps))
	}

	entryCount := 0
	for i, wantMap := range manifest.Maps {
		gotMap := mjSourceMaps[i]
		if gotMap.Name != wantMap.Name || string(gotMap.Kind) != wantMap.Kind ||
			gotMap.Source != wantMap.Source || gotMap.Line != wantMap.Line ||
			gotMap.Parser != wantMap.Parser || gotMap.Methods != wantMap.Methods ||
			gotMap.Pattern != wantMap.Pattern || gotMap.Flags != wantMap.Flags ||
			gotMap.Dynamic != wantMap.Dynamic || gotMap.Priority != wantMap.Priority {
			t.Errorf("map[%d] metadata = %+v, manifest = %+v", i, gotMap, wantMap)
		}
		if gotHash := mjSourceMapPayloadHash(t, gotMap.Entries); gotHash != wantMap.PayloadSHA256 {
			t.Errorf("map %q ordered payload hash = %s, manifest = %s", gotMap.Name, gotHash, wantMap.PayloadSHA256)
		}
		if len(gotMap.Entries) != len(wantMap.Entries) {
			t.Errorf("map %q entries = %d, manifest = %d", gotMap.Name, len(gotMap.Entries), len(wantMap.Entries))
			continue
		}
		for j, wantEntry := range wantMap.Entries {
			gotEntry := gotMap.Entries[j]
			if gotEntry.Name != wantEntry.Name {
				t.Errorf("map %q entry[%d] = %q, manifest = %q", gotMap.Name, j, gotEntry.Name, wantEntry.Name)
			}
			if gotHash := mjSourcePayloadHash(t, gotEntry.Value); gotHash != wantEntry.PayloadSHA256 {
				t.Errorf("map %q entry %q payload hash = %s, manifest = %s", gotMap.Name, gotEntry.Name, gotHash, wantEntry.PayloadSHA256)
			}
			entryCount++
		}
	}
	if entryCount != manifest.EntryCount {
		t.Errorf("compared %d entries, manifest declares %d", entryCount, manifest.EntryCount)
	}

	if len(manifest.Packages) != len(mjSourcePackages) {
		t.Fatalf("manifest packages = %d, Go packages = %d", len(manifest.Packages), len(mjSourcePackages))
	}
	for i, wantPackage := range manifest.Packages {
		gotPackage := mjSourcePackages[i]
		if gotPackage.Name != wantPackage.Name || gotPackage.Source != wantPackage.Source || gotPackage.Line != wantPackage.Line {
			t.Errorf("package[%d] metadata = %q/%q/%d, manifest = %q/%q/%d", i,
				gotPackage.Name, gotPackage.Source, gotPackage.Line,
				wantPackage.Name, wantPackage.Source, wantPackage.Line)
		}
		compareMJSourceHandlers(t, gotPackage.Name+" handlers", gotPackage.Handlers, wantPackage.Handlers)
		compareMJSourceReferences(t, gotPackage.Name+" fallbacks", gotPackage.Fallbacks, wantPackage.Fallbacks)
		compareMJSourceReferences(t, gotPackage.Name+" items", gotPackage.Items, wantPackage.Items)
		compareMJSourceReferences(t, gotPackage.Name+" tags", gotPackage.Tags, wantPackage.Tags)
		if gotHash := mjSourcePayloadHash(t, gotPackage.Options); gotHash != wantPackage.OptionsPayloadSHA256 {
			t.Errorf("package %q options payload hash = %s, manifest = %s", gotPackage.Name, gotHash, wantPackage.OptionsPayloadSHA256)
		}
		if !slices.Equal(gotPackage.DynamicMapNames, wantPackage.DynamicMapNames) {
			t.Errorf("package %q dynamic maps = %v, manifest = %v", gotPackage.Name, gotPackage.DynamicMapNames, wantPackage.DynamicMapNames)
		}
		compareMJSourceHandlers(t, gotPackage.Name+" dynamic handlers", gotPackage.DynamicHandlers, wantPackage.DynamicHandlers)
		if gotPackage.DynamicPriority != wantPackage.DynamicPriority {
			t.Errorf("package %q dynamic priority = %d, manifest = %d", gotPackage.Name, gotPackage.DynamicPriority, wantPackage.DynamicPriority)
		}
		if !slices.Equal(gotPackage.Hooks, wantPackage.Hooks) {
			t.Errorf("package %q hooks = %#v, manifest = %#v", gotPackage.Name, gotPackage.Hooks, wantPackage.Hooks)
		}
		if !slices.Equal(gotPackage.Postprocessors, wantPackage.Postprocessors) {
			t.Errorf("package %q postprocessors = %#v, manifest = %#v", gotPackage.Name, gotPackage.Postprocessors, wantPackage.Postprocessors)
		}
	}

	closure := manifest.D2PackageClosure
	if closure.D2GitCommit != "541941e895fcdc69ad3109770296dcac670c5be7" ||
		closure.SelectionSource != "d2renderers/d2latex/setup.js" ||
		closure.SelectionSourceSHA256 != "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881" ||
		closure.DependencySource != "components/src/dependencies.js" ||
		closure.DependencySourceHash != "afdd598f4dcd58c68583a3bd33a9ec2fe2dfacf5c01f76202e3d87d07c1d22b8" {
		t.Errorf("unexpected D2 closure provenance: %+v", closure)
	}
	if !slices.Equal(mjD2SelectedPackageOrder, closure.SelectedPackageOrder) {
		t.Errorf("D2 selected package order = %v, manifest = %v", mjD2SelectedPackageOrder, closure.SelectedPackageOrder)
	}
	if !slices.Equal(mjD2DependencyPackages, closure.DependencyOnly) {
		t.Errorf("D2 dependency-only packages = %v, manifest = %v", mjD2DependencyPackages, closure.DependencyOnly)
	}
}

func TestMJSourceMapCoverage(t *testing.T) {
	if len(mjSourceMaps) != 56 {
		t.Fatalf("len(mjSourceMaps) = %d, want 56", len(mjSourceMaps))
	}
	total, dynamic := 0, 0
	for _, sourceMap := range mjSourceMaps {
		total += len(sourceMap.Entries)
		if sourceMap.Dynamic {
			dynamic++
		}
	}
	if total != 1273 || dynamic != 5 {
		t.Errorf("source coverage = %d entries/%d dynamic maps, want 1273/5", total, dynamic)
	}
}

func TestMJSourceMapsHaveUniqueOrderedEntries(t *testing.T) {
	mapNames := make(map[string]bool)
	for _, sourceMap := range mjSourceMaps {
		if mapNames[sourceMap.Name] {
			t.Errorf("duplicate map %q", sourceMap.Name)
		}
		mapNames[sourceMap.Name] = true
		entries := make(map[string]bool)
		for _, entry := range sourceMap.Entries {
			if entries[entry.Name] {
				t.Errorf("map %q has duplicate entry %q", sourceMap.Name, entry.Name)
			}
			entries[entry.Name] = true
			validateMJSourceValue(t, sourceMap.Name+"."+entry.Name, entry.Value)
		}
	}
}

func TestMJSourceRepresentativePayloads(t *testing.T) {
	tests := []struct {
		mapName string
		entry   string
		want    any
	}{
		{"special", "{", "Open"},
		{"mathchar0mi", "alpha", "α"},
		{"not_remap", "∃", "∄"},
		{"remap", "-", "−"},
		{"AMSmath-mathchar0mo", "iiiint", mjSourceList{"⨌", mjSourceObject{{Name: "texClass", Value: 1}}}},
		{"AMSmath-macros", "negmedspace", mjSourceList{"Spacer", -4.0 / 18}},
		{"AMSmath-environment", "smallmatrix", mjSourceList{"Array", nil, nil, nil, "c", "0.333em", ".2em", "S", 1}},
		{"AMSsymbols-mathchar0mi", "Bbbk", mjSourceList{"k", mjSourceObject{{Name: "mathvariant", Value: "double-struck"}}}},
		{"AMSsymbols-mathchar0mo", "rightleftharpoons", mjSourceList{"⇌", mjSourceObject{{Name: "variantForm", Value: true}}}},
		{"mathtools-macros", "adjustlimits", mjSourceList{
			"MacroWithTemplate",
			"\\mathop{{#1}\\vphantom{{#3}}}_{{#2}\\vphantom{{#4}}}\\mathop{{#3}\\vphantom{{#1}}}_{{#4}\\vphantom{{#2}}}",
			4, mjUndefined, "_", mjUndefined, "_",
		}},
		{"Braket-macros", "Braket", mjSourceList{"Braket", "⟨", "⟩", true, mjPositiveInfinity}},
		{"cases-env", "subnumcases", mjSourceList{"NumCases", "cases"}},
		{"Physics-vector-mo", "gradientnabla", mjSourceList{"∇", mjSourceObject{{Name: "mathvariant", Value: "bold"}}}},
		{"Physics-expressions-macros", "Trace", mjSourceList{"Expression", false, "Tr"}},
		{"Physics-characters", "|", mjSourceList{"AutoClose", 0}},
		{"empheq-macros", "empheqbiglVert", mjSourceList{"EmpheqMO", "‖"}},
		{"Newcommand-macros", "let", "Let"},
		{"color", "fcolorbox", "FColorBox"},
		{"cancel", "xcancel", mjSourceList{"Cancel", "updiagonalstrike downdiagonalstrike"}},
		{"gensymb-symbols", "degree", "°"},
		{"mhchem", "ce", mjSourceList{"Machine", "ce"}},
	}
	for _, test := range tests {
		got := findMJSourceEntry(t, test.mapName, test.entry)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s.%s = %#v, want %#v", test.mapName, test.entry, got, test.want)
		}
	}
	operatorLetters := findMJSourceMap(t, "AMSmath-operatorLetter")
	if operatorLetters.Pattern != "[-*]" || operatorLetters.Flags != "i" || operatorLetters.Parser != "AmsMethods.operatorLetter" {
		t.Errorf("operator-letter regexp = %#v", operatorLetters)
	}
	letter := findMJSourceMap(t, "letter")
	if letter.Pattern != "[a-z]" || letter.Flags != "i" || letter.Parser != "ParseMethods.variable" {
		t.Errorf("base letter regexp = %#v", letter)
	}
}

func TestMJSourcePackageCoverageAndReferences(t *testing.T) {
	wantPackages := []string{
		"base", "ams", "mathtools", "amscd", "braket", "cases", "physics",
		"empheq", "newcommand", "color", "enclose", "cancel", "gensymb", "mhchem",
	}
	if len(mjSourcePackages) != len(wantPackages) {
		t.Fatalf("len(mjSourcePackages) = %d, want %d", len(mjSourcePackages), len(wantPackages))
	}
	knownMaps := make(map[string]bool)
	for _, sourceMap := range mjSourceMaps {
		knownMaps[sourceMap.Name] = true
	}
	dynamicNames := make(map[string]bool)
	for i, pkg := range mjSourcePackages {
		if pkg.Name != wantPackages[i] {
			t.Errorf("package[%d] = %q, want %q", i, pkg.Name, wantPackages[i])
		}
		validateMJSourceValue(t, pkg.Name+".options", pkg.Options)
		for _, mapName := range pkg.DynamicMapNames {
			if dynamicNames[mapName] {
				t.Errorf("dynamic map %q is owned by more than one package", mapName)
			}
			dynamicNames[mapName] = true
			sourceMap := findMJSourceMap(t, mapName)
			if !sourceMap.Dynamic || sourceMap.Priority != pkg.DynamicPriority || len(sourceMap.Entries) != 0 {
				t.Errorf("package %q dynamic map %q = dynamic %v/priority %d/%d entries", pkg.Name,
					mapName, sourceMap.Dynamic, sourceMap.Priority, len(sourceMap.Entries))
			}
		}
		for _, handler := range append(append([]mjSourceHandler{}, pkg.Handlers...), pkg.DynamicHandlers...) {
			for _, mapName := range handler.Maps {
				if !knownMaps[mapName] {
					t.Errorf("package %q %s handler references unknown map %q", pkg.Name, handler.Kind, mapName)
				}
			}
		}
	}
	if len(dynamicNames) != 5 {
		t.Errorf("dynamic map count = %d, want 5", len(dynamicNames))
	}
	for _, sourceMap := range mjSourceMaps {
		if sourceMap.Dynamic && !dynamicNames[sourceMap.Name] {
			t.Errorf("dynamic map %q is not owned by a package", sourceMap.Name)
		}
	}

	closure := append(append([]string{}, mjD2SelectedPackageOrder...), mjD2DependencyPackages...)
	if len(closure) != 14 {
		t.Errorf("D2 package closure has %d packages, want 14", len(closure))
	}
	closureNames := make(map[string]bool)
	for _, name := range closure {
		if closureNames[name] {
			t.Errorf("D2 package closure repeats %q", name)
		}
		closureNames[name] = true
		findMJSourcePackage(t, name)
	}
	for _, pkg := range mjSourcePackages {
		if !closureNames[pkg.Name] {
			t.Errorf("source package %q is outside D2's closure", pkg.Name)
		}
	}

	mathtools := findMJSourcePackage(t, "mathtools")
	if mathtools.DynamicPriority != -5 || !reflect.DeepEqual(mathtools.Postprocessors, []mjSourcePostprocessor{{Name: "fixPrescripts", Priority: -6}}) {
		t.Errorf("mathtools dynamic/postprocessor config = %#v", mathtools)
	}
	newcommand := findMJSourcePackage(t, "newcommand")
	if !reflect.DeepEqual(newcommand.Options, mjSourceObject{{Name: "maxMacros", Value: 1000}}) || newcommand.DynamicPriority != -1 {
		t.Errorf("newcommand options/init = %#v", newcommand)
	}
	base := findMJSourcePackage(t, "base")
	if !reflect.DeepEqual(base.Postprocessors, []mjSourcePostprocessor{{Name: "filterNonscript", Priority: -4}}) ||
		len(base.Fallbacks) != 3 {
		t.Errorf("base fallback/postprocessor config = %#v", base)
	}
	color := findMJSourcePackage(t, "color")
	if !reflect.DeepEqual(color.Hooks, []mjSourceHook{{Kind: "config", Implementation: "config"}}) {
		t.Errorf("color config hook = %#v", color.Hooks)
	}
}

func validateMJSourceValue(t *testing.T, path string, value any) {
	t.Helper()
	switch value := value.(type) {
	case nil, string, bool, int, float64, mjSourceInfinity, mjSourceUndefined, mjSourceExpression:
		return
	case mjSourceExpandable:
		validateMJSourceValue(t, path+".$expandable", value.Value)
	case mjSourceList:
		for i, child := range value {
			validateMJSourceValue(t, path+"[]", child)
			_ = i
		}
	case mjSourceObject:
		names := make(map[string]bool)
		for _, property := range value {
			if names[property.Name] {
				t.Errorf("%s has duplicate object property %q", path, property.Name)
			}
			names[property.Name] = true
			validateMJSourceValue(t, path+"."+property.Name, property.Value)
		}
	default:
		t.Errorf("%s has unsupported payload type %T", path, value)
	}
}

func findMJSourceMap(t *testing.T, name string) mjSourceMap {
	t.Helper()
	for _, sourceMap := range mjSourceMaps {
		if sourceMap.Name == name {
			return sourceMap
		}
	}
	t.Fatalf("source map %q not found", name)
	return mjSourceMap{}
}

func findMJSourceEntry(t *testing.T, mapName, entryName string) any {
	t.Helper()
	sourceMap := findMJSourceMap(t, mapName)
	for _, entry := range sourceMap.Entries {
		if entry.Name == entryName {
			return entry.Value
		}
	}
	t.Fatalf("source entry %q.%q not found", mapName, entryName)
	return nil
}

func findMJSourcePackage(t *testing.T, name string) mjSourcePackage {
	t.Helper()
	for _, pkg := range mjSourcePackages {
		if pkg.Name == name {
			return pkg
		}
	}
	t.Fatalf("source package %q not found", name)
	return mjSourcePackage{}
}

func compareMJSourceHandlers(t *testing.T, label string, got, want []mjSourceHandler) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s count = %d, manifest = %d", label, len(got), len(want))
		return
	}
	for i := range want {
		if got[i].Kind != want[i].Kind || !slices.Equal(got[i].Maps, want[i].Maps) {
			t.Errorf("%s[%d] = %#v, manifest = %#v", label, i, got[i], want[i])
		}
	}
}

func compareMJSourceReferences(t *testing.T, label string, got, want []mjSourceReference) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s = %#v, manifest = %#v", label, got, want)
	}
}

type mjCanonicalObject struct {
	Properties [][2]any `json:"$object"`
}

type mjCanonicalNumber struct {
	Number string `json:"$number"`
}

type mjCanonicalUndefined struct {
	Undefined bool `json:"$undefined"`
}

type mjCanonicalExpandable struct {
	Value any `json:"$expandable"`
}

type mjCanonicalExpression struct {
	Source string `json:"$expression"`
}

func mjSourceMapPayloadHash(t *testing.T, entries []mjSourceEntry) string {
	t.Helper()
	payload := make(mjSourceList, len(entries))
	for i, entry := range entries {
		payload[i] = mjSourceList{entry.Name, entry.Value}
	}
	return mjSourcePayloadHash(t, payload)
}

func mjSourcePayloadHash(t *testing.T, value any) string {
	t.Helper()
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(mjCanonicalSourceValue(t, value)); err != nil {
		t.Fatalf("encode canonical source value: %v", err)
	}
	encoded := bytes.TrimSuffix(buffer.Bytes(), []byte{'\n'})
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func mjCanonicalSourceValue(t *testing.T, value any) any {
	t.Helper()
	switch value := value.(type) {
	case nil, string, bool, int, float64:
		return value
	case mjSourceInfinity:
		return mjCanonicalNumber{Number: "Infinity"}
	case mjSourceUndefined:
		return mjCanonicalUndefined{Undefined: true}
	case mjSourceExpandable:
		return mjCanonicalExpandable{Value: mjCanonicalSourceValue(t, value.Value)}
	case mjSourceExpression:
		return mjCanonicalExpression{Source: string(value)}
	case mjSourceList:
		result := make([]any, len(value))
		for i, child := range value {
			result[i] = mjCanonicalSourceValue(t, child)
		}
		return result
	case mjSourceObject:
		result := mjCanonicalObject{Properties: make([][2]any, len(value))}
		for i, property := range value {
			result.Properties[i] = [2]any{property.Name, mjCanonicalSourceValue(t, property.Value)}
		}
		return result
	default:
		t.Fatalf("unsupported canonical source value %T", value)
		return nil
	}
}
