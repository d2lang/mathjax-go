// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

type d2RichManifest struct {
	SchemaVersion int `json:"schema_version"`
	Oracle        struct {
		Project                string `json:"project"`
		Version                string `json:"version"`
		MathJaxGitCommit       string `json:"mathjax_git_commit"`
		D2BundleGitCommit      string `json:"d2_bundle_git_commit"`
		FormulaSourceGitCommit string `json:"formula_source_git_commit"`
		Runtime                string `json:"runtime"`
		State                  string `json:"state"`
		Display                bool   `json:"display"`
		PolyfillsSHA256        string `json:"polyfills_sha256"`
		MathJaxBundleSHA256    string `json:"mathjax_bundle_sha256"`
		SetupSHA256            string `json:"setup_sha256"`
		Provenance             string `json:"provenance"`
	} `json:"oracle"`
	CaseCount int            `json:"case_count"`
	Registry  d2RichRegistry `json:"registry"`
	Cases     []d2RichCase   `json:"cases"`
}

type d2RichRegistry struct {
	Global      []d2RichPair       `json:"global"`
	Definitions []d2RichDefinition `json:"definitions"`
}

type d2RichDefinition struct {
	Kind     string       `json:"kind"`
	Defaults []d2RichPair `json:"defaults"`
	Flags    d2RichFlags  `json:"flags"`
}

type d2RichCase struct {
	Name string      `json:"name"`
	TeX  string      `json:"tex"`
	Tree *d2RichNode `json:"tree"`
}

type d2RichNode struct {
	Kind       string           `json:"kind"`
	Attributes d2RichAttributes `json:"attributes"`
	Properties []d2RichPair     `json:"properties"`
	Flags      d2RichFlags      `json:"flags"`
	TeXClass   any              `json:"texClass"`
	PrevClass  any              `json:"prevClass"`
	PrevLevel  any              `json:"prevLevel"`
	Children   []*d2RichNode    `json:"children"`
	Text       *string          `json:"text,omitempty"`
	XML        *string          `json:"xml,omitempty"`
}

type d2RichAttributes struct {
	Explicit  []d2RichPair `json:"explicit"`
	Inherited []d2RichPair `json:"inherited"`
	Defaults  []d2RichPair `json:"defaults"`
	Global    []d2RichPair `json:"global"`
}

type d2RichPair struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type d2RichFlags struct {
	Token              bool `json:"token"`
	Embellished        bool `json:"embellished"`
	Spacelike          bool `json:"spacelike"`
	LinebreakContainer bool `json:"linebreakContainer"`
	HasNewline         bool `json:"hasNewline"`
	Arity              any  `json:"arity"`
	Inferred           bool `json:"inferred"`
	NotParent          bool `json:"notParent"`
	CoreIndex          int  `json:"coreIndex"`
}

var d2RichRegistryKinds = []string{
	"math", "mi", "mn", "mo", "mtext", "mspace", "ms", "mrow", "inferredMrow",
	"mfrac", "msqrt", "mroot", "mstyle", "merror", "mpadded", "mphantom", "mfenced",
	"menclose", "maction", "msub", "msup", "msubsup", "munder", "mover", "munderover",
	"mmultiscripts", "mprescripts", "none", "mtable", "mlabeledtr", "mtr", "mtd",
	"maligngroup", "malignmark", "mglyph", "semantics", "annotation", "annotation-xml",
	"TeXAtom", "MathChoice", "text", "XML",
}

var d2RichCaseNames = []string{
	"basic", "fraction", "multiline", "scripts", "delimiters", "ams_matrix", "ams_align",
	"mathtools", "amscd", "braket", "cancel", "cases", "color", "gensymb", "mhchem",
	"physics", "invalid_merror", "d2_huge", "d2_emc2", "d2_gibberish_sum",
	"d2_linear_program", "d2_equation_split", "d2_limit", "d2_quadratic",
	"d2_physics_plugin", "d2_displaylines", "d2_add",
}

func TestD2RichMathJax322Oracle(t *testing.T) {
	data, err := os.ReadFile("testdata/d2_rich_mml_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest d2RichManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	validateD2RichManifest(t, &manifest)

	registryMismatches := 0
	caseMismatches := 0
	first := ""
	firstCase := ""
	record := func(scope string, want, got any) {
		if reflect.DeepEqual(want, got) {
			return
		}
		if strings.HasPrefix(scope, "registry") {
			registryMismatches++
		} else {
			caseMismatches++
		}
		path, wantValue, gotValue := d2RichFirstDifference(reflect.ValueOf(want), reflect.ValueOf(got), scope)
		detail := fmt.Sprintf("%s: got %s, want %s", path, d2RichShortValue(gotValue), d2RichShortValue(wantValue))
		if first == "" {
			first = detail
		}
		if !strings.HasPrefix(scope, "registry") && firstCase == "" {
			firstCase = detail
		}
	}

	actualRegistry, err := d2RichGoRegistry()
	if err != nil {
		t.Fatal(err)
	}
	record("registry.global", manifest.Registry.Global, actualRegistry.Global)
	for i := range manifest.Registry.Definitions {
		record("registry.definitions["+manifest.Registry.Definitions[i].Kind+"]",
			manifest.Registry.Definitions[i], actualRegistry.Definitions[i])
	}

	for _, test := range manifest.Cases {
		root, err := NewCompiler().Compile(test.TeX, true)
		if err != nil {
			caseMismatches++
			if first == "" {
				first = fmt.Sprintf("cases[%s]: Compile returned %v", test.Name, err)
			}
			if firstCase == "" {
				firstCase = fmt.Sprintf("cases[%s]: Compile returned %v", test.Name, err)
			}
			continue
		}
		record("cases["+test.Name+"]", test.Tree, d2RichGoNode(root))
	}

	if registryMismatches+caseMismatches != 0 {
		t.Fatalf("%d rich MML oracle unit mismatches (%d registry, %d cases); first: %s; first case: %s",
			registryMismatches+caseMismatches, registryMismatches, caseMismatches, first, firstCase)
	}
}

func validateD2RichManifest(t *testing.T, manifest *d2RichManifest) {
	t.Helper()
	if manifest.SchemaVersion != 1 {
		t.Fatalf("oracle schema version = %d, want 1", manifest.SchemaVersion)
	}
	if manifest.Oracle.Project != "MathJax" || manifest.Oracle.Version != "3.2.2" ||
		manifest.Oracle.MathJaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
		t.Fatalf("unexpected MathJax provenance: %#v", manifest.Oracle)
	}
	if manifest.Oracle.D2BundleGitCommit != "541941e895fcdc69ad3109770296dcac670c5be7" ||
		manifest.Oracle.FormulaSourceGitCommit != "5666d9337c77a4803b9cd60cb1c5a24439d0f949" {
		t.Fatalf("unexpected D2 provenance: %#v", manifest.Oracle)
	}
	if manifest.Oracle.Runtime != "Goja" || manifest.Oracle.State != "MathJax._.core.MathItem.STATE.COMPILED" || !manifest.Oracle.Display {
		t.Fatalf("unexpected extraction settings: %#v", manifest.Oracle)
	}
	if manifest.Oracle.PolyfillsSHA256 != "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01" ||
		manifest.Oracle.MathJaxBundleSHA256 != "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869" ||
		manifest.Oracle.SetupSHA256 != "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881" {
		t.Fatal("frozen D2 asset checksums changed")
	}
	if manifest.CaseCount != len(d2RichCaseNames) || len(manifest.Cases) != len(d2RichCaseNames) {
		t.Fatalf("oracle has case_count=%d and %d cases, want %d", manifest.CaseCount, len(manifest.Cases), len(d2RichCaseNames))
	}
	for i, name := range d2RichCaseNames {
		if manifest.Cases[i].Name != name || manifest.Cases[i].TeX == "" || manifest.Cases[i].Tree == nil {
			t.Fatalf("oracle case %d = %q, want complete %q", i, manifest.Cases[i].Name, name)
		}
	}
	if len(manifest.Registry.Definitions) != len(d2RichRegistryKinds) {
		t.Fatalf("oracle has %d registry definitions, want %d", len(manifest.Registry.Definitions), len(d2RichRegistryKinds))
	}
	for i, kind := range d2RichRegistryKinds {
		if manifest.Registry.Definitions[i].Kind != kind {
			t.Fatalf("registry definition %d = %q, want %q", i, manifest.Registry.Definitions[i].Kind, kind)
		}
	}
}

func d2RichGoRegistry() (d2RichRegistry, error) {
	registry := d2RichRegistry{
		Global:      d2RichPairs(texMMLFactory.Globals().Keys(), texMMLFactory.Globals().Get),
		Definitions: make([]d2RichDefinition, 0, len(d2RichRegistryKinds)),
	}
	for _, kind := range d2RichRegistryKinds {
		definition, ok := texMMLFactory.Definition(kind)
		if !ok {
			return d2RichRegistry{}, fmt.Errorf("missing Go MML registry definition %q", kind)
		}
		sample := texMMLFactory.Create(kind)
		switch {
		case definition.Flags.Arity == -1:
			sample.SetChildren([]*mml.Node{texMMLFactory.Create("inferredMrow")})
		case definition.Flags.Arity > 0 && definition.Flags.Arity != mmlUnboundedArity:
			children := make([]*mml.Node, definition.Flags.Arity)
			for i := range children {
				children[i] = texMMLFactory.Create("mi")
			}
			sample.SetChildren(children)
		}
		refreshDynamicFlags(sample)
		registry.Definitions = append(registry.Definitions, d2RichDefinition{
			Kind:     kind,
			Defaults: d2RichPairs(definition.Defaults.Keys(), definition.Defaults.Get),
			Flags:    d2RichGoFlags(sample.Flags),
		})
	}
	return registry, nil
}

func d2RichGoNode(node *mml.Node) *d2RichNode {
	if node == nil {
		return nil
	}
	kind := node.Kind
	if kind == "mrow" && node.Flags.Inferred {
		kind = "inferredMrow"
	}
	var texClass any
	if node.TeXClass != mml.TeXClassNone || node.Kind == "text" || node.Kind == "mspace" {
		texClass = d2RichGoValue(node.TeXClass)
	} else if explicit, ok := node.Property("texClass"); ok {
		texClass = d2RichGoValue(explicit)
	}
	var prevClass, prevLevel any
	if node.Kind == "text" {
		prevClass = d2RichGoValue(node.PrevClass)
		prevLevel = d2RichGoValue(node.PrevLevel)
	}
	result := &d2RichNode{
		Kind: kind,
		Attributes: d2RichAttributes{
			Explicit:  d2RichPairs(node.Attributes.ExplicitNames(), node.Attributes.Explicit().Get),
			Inherited: d2RichPairs(node.Attributes.InheritedNames(), node.Attributes.Inherited().Get),
			Defaults:  d2RichPairs(node.Attributes.DefaultNames(), node.Attributes.Defaults().Get),
			Global:    d2RichPairs(node.Attributes.GlobalNames(), node.Attributes.Globals().Get),
		},
		Properties: d2RichPairs(node.Properties.Keys(), node.Properties.Get),
		Flags:      d2RichGoFlags(node.Flags),
		// Go uses TeXClassNone as the storage sentinel for MathJax's null
		// _texClass.  A texClass property distinguishes an explicitly supplied
		// NONE operator (notably U+2061) from that unset sentinel.  MML text
		// leaves inherit AbstractMmlEmptyNode's fixed NONE/0 previous state;
		// other nodes remain null until the output-jax TeX-spacing pass.
		TeXClass:  texClass,
		PrevClass: prevClass,
		PrevLevel: prevLevel,
		Children:  make([]*d2RichNode, len(node.Children)),
	}
	if node.Kind == "text" {
		text := node.Text
		result.Text = &text
	}
	for i, child := range node.Children {
		result.Children[i] = d2RichGoNode(child)
	}
	return result
}

func d2RichGoFlags(flags mml.Flags) d2RichFlags {
	return d2RichFlags{
		Token:              flags.Token,
		Embellished:        flags.Embellished,
		Spacelike:          flags.Spacelike,
		LinebreakContainer: flags.LinebreakContainer,
		HasNewline:         flags.HasNewline,
		Arity:              d2RichGoArity(flags.Arity),
		Inferred:           flags.Inferred,
		NotParent:          flags.NotParent,
		CoreIndex:          flags.CoreIndex,
	}
}

func d2RichGoArity(arity int) any {
	if arity == mmlUnboundedArity {
		return map[string]any{"$": "infinity"}
	}
	return float64(arity)
}

func d2RichPairs(names []string, get func(string) (mml.Property, bool)) []d2RichPair {
	pairs := make([]d2RichPair, 0, len(names))
	for _, name := range names {
		value, _ := get(name)
		pairs = append(pairs, d2RichPair{Name: name, Value: d2RichGoValue(value)})
	}
	return pairs
}

func d2RichGoValue(value any) any {
	if value == nil {
		return nil
	}
	if mml.IsInherit(value) {
		return map[string]any{"$": "inherit"}
	}
	switch value := value.(type) {
	case mjSourceUndefined:
		return map[string]any{"$": "undefined"}
	case mjSourceInfinity:
		if value == mjPositiveInfinity {
			return map[string]any{"$": "infinity"}
		}
		return map[string]any{"$": "-infinity"}
	case mml.TeXClass:
		return float64(value)
	case string:
		return value
	case bool:
		return value
	case int:
		return float64(value)
	case int8:
		return float64(value)
	case int16:
		return float64(value)
	case int32:
		return float64(value)
	case int64:
		return float64(value)
	case uint:
		return float64(value)
	case uint8:
		return float64(value)
	case uint16:
		return float64(value)
	case uint32:
		return float64(value)
	case uint64:
		return float64(value)
	case float32:
		return d2RichFloat(float64(value))
	case float64:
		return d2RichFloat(value)
	}
	return d2RichReflectValue(reflect.ValueOf(value))
}

func d2RichFloat(value float64) any {
	switch {
	case math.IsInf(value, 1):
		return map[string]any{"$": "infinity"}
	case math.IsInf(value, -1):
		return map[string]any{"$": "-infinity"}
	case math.IsNaN(value):
		return map[string]any{"$": "nan"}
	default:
		return value
	}
}

func d2RichReflectValue(value reflect.Value) any {
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil
	}
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(value.Uint())
	case reflect.Float32, reflect.Float64:
		return d2RichFloat(value.Float())
	case reflect.Array, reflect.Slice:
		values := make([]any, value.Len())
		for i := range values {
			values[i] = d2RichGoValue(value.Index(i).Interface())
		}
		return map[string]any{"$": "array", "values": values}
	case reflect.Map:
		keys := make([]string, 0, value.Len())
		for _, key := range value.MapKeys() {
			keys = append(keys, fmt.Sprint(key.Interface()))
		}
		sort.Strings(keys)
		pairs := make([]any, 0, len(keys))
		for _, key := range keys {
			mapValue := value.MapIndex(reflect.ValueOf(key).Convert(value.Type().Key()))
			pairs = append(pairs, map[string]any{"name": key, "value": d2RichGoValue(mapValue.Interface())})
		}
		return map[string]any{"$": "object", "pairs": pairs}
	}
	return map[string]any{"$": "go:" + value.Type().String(), "value": fmt.Sprint(value.Interface())}
}

func d2RichFirstDifference(want, got reflect.Value, path string) (string, any, any) {
	for want.IsValid() && (want.Kind() == reflect.Interface || want.Kind() == reflect.Pointer) {
		if want.IsNil() {
			break
		}
		want = want.Elem()
	}
	for got.IsValid() && (got.Kind() == reflect.Interface || got.Kind() == reflect.Pointer) {
		if got.IsNil() {
			break
		}
		got = got.Elem()
	}
	if !want.IsValid() || !got.IsValid() || want.Type() != got.Type() {
		return path, d2RichInterface(want), d2RichInterface(got)
	}
	if reflect.DeepEqual(d2RichInterface(want), d2RichInterface(got)) {
		return path, d2RichInterface(want), d2RichInterface(got)
	}
	switch want.Kind() {
	case reflect.Struct:
		for i := 0; i < want.NumField(); i++ {
			if !reflect.DeepEqual(want.Field(i).Interface(), got.Field(i).Interface()) {
				return d2RichFirstDifference(want.Field(i), got.Field(i), path+"."+want.Type().Field(i).Name)
			}
		}
	case reflect.Slice, reflect.Array:
		if want.Len() != got.Len() {
			return path + ".length", want.Len(), got.Len()
		}
		for i := 0; i < want.Len(); i++ {
			if !reflect.DeepEqual(want.Index(i).Interface(), got.Index(i).Interface()) {
				return d2RichFirstDifference(want.Index(i), got.Index(i), fmt.Sprintf("%s[%d]", path, i))
			}
		}
	case reflect.Map:
		keys := make([]string, 0, want.Len()+got.Len())
		seen := make(map[string]bool)
		for _, source := range []reflect.Value{want, got} {
			for _, key := range source.MapKeys() {
				name := fmt.Sprint(key.Interface())
				if !seen[name] {
					seen[name] = true
					keys = append(keys, name)
				}
			}
		}
		sort.Strings(keys)
		for _, name := range keys {
			key := reflect.ValueOf(name).Convert(want.Type().Key())
			wantValue, gotValue := want.MapIndex(key), got.MapIndex(key)
			if !reflect.DeepEqual(d2RichInterface(wantValue), d2RichInterface(gotValue)) {
				return d2RichFirstDifference(wantValue, gotValue, path+"["+name+"]")
			}
		}
	}
	return path, d2RichInterface(want), d2RichInterface(got)
}

func d2RichInterface(value reflect.Value) any {
	if !value.IsValid() {
		return nil
	}
	return value.Interface()
}

func d2RichShortValue(value any) string {
	text := fmt.Sprintf("%#v", value)
	if len(text) > 240 {
		return text[:237] + "..."
	}
	return text
}
