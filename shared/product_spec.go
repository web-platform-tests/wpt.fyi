// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/json"
	"sort"
	"strings"

	mapset "github.com/deckarep/golang-set"
)

// ProductSpec is a struct representing a parsed product spec string.
type ProductSpec struct {
	ProductAtRevision

	Labels mapset.Set
}

// BrowserMatches returns whether the browser matches the given run.
func (p ProductSpec) BrowserMatches(run TestRun) bool {
	return run.BrowserName == p.BrowserName
}

// Matches returns whether the spec matches the given run.
func (p ProductSpec) Matches(run TestRun) bool {
	if !p.BrowserMatches(run) {
		return false
	}
	if !IsLatest(p.Revision) && p.Revision != run.Revision {
		return false
	}
	if p.Labels != nil && p.Labels.Cardinality() > 0 {
		runLabels := run.LabelsSet()
		if !p.Labels.IsSubset(runLabels) {
			return false
		}
	}
	if p.BrowserVersion != "" {
		// Make "6" not match "60.123" by adding trailing dots to both.
		if !strings.HasPrefix(run.BrowserVersion+".", p.BrowserVersion+".") {
			return false
		}
	}
	return true
}

// IsExperimental returns true if the product spec is restricted to experimental
// runs (i.e. has the label "experimental").
func (p ProductSpec) IsExperimental() bool {
	return p.Labels != nil && p.Labels.Contains(ExperimentalLabel)
}

// DisplayName returns a capitalized version of the product's name.
func (p ProductSpec) DisplayName() string {
	switch p.BrowserName {
	case "chrome":
		return "Chrome"
	case "edge":
		return "Edge"
	case "firefox":
		return "Firefox"
	case "safari":
		return "Safari"
	default:
		return p.BrowserName
	}
}

// ProductSpecs is a helper type for a slice of ProductSpec structs.
type ProductSpecs []ProductSpec

// Products gets the slice of products specified in the ProductSpecs slice.
func (p ProductSpecs) Products() []Product {
	result := make([]Product, len(p))
	for i, spec := range p {
		result[i] = spec.Product
	}
	return result
}

// OrDefault returns the current product specs, or the default if the set is empty.
func (p ProductSpecs) OrDefault() ProductSpecs {
	if len(p) < 1 {
		return GetDefaultProducts()
	}
	return p
}

// Strings returns the array of the ProductSpec items as their string
// representations.
func (p ProductSpecs) Strings() []string {
	result := make([]string, len(p))
	for i, spec := range p {
		result[i] = spec.String()
	}
	return result
}

func (p ProductSpec) String() string {
	s := p.Product.String()
	if p.Labels != nil {
		p.Labels.Remove("") // Remove the empty label, if present.
		if p.Labels.Cardinality() > 0 {
			labels := make([]string, 0, p.Labels.Cardinality())
			for l := range p.Labels.Iter() {
				labels = append(labels, l.(string))
			}
			sort.Strings(labels) // Deterministic String() output.
			s += "[" + strings.Join(labels, ",") + "]"
		}
	}
	if !IsLatest(p.Revision) {
		s += "@" + p.Revision
	}
	return s
}

func (p ProductSpecs) Len() int           { return len(p) }
func (p ProductSpecs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ProductSpecs) Less(i, j int) bool { return p[i].String() < p[j].String() }

// MarshalJSON treats the set as an array so it can be marshalled.
func (p ProductSpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON parses an array so that ProductSpec can be unmarshalled.
func (p *ProductSpec) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p, err = ParseProductSpec(s)
	return err
}

// UnmarshalYAML parses an array so that ProductSpec can be unmarshalled.
func (p *ProductSpec) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	*p, err = ParseProductSpec(s)
	return err
}
