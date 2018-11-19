package shared

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	mapset "github.com/deckarep/golang-set"
)

// TestRunFilter represents the ways TestRun entities can be filtered in
// the webapp and api.
type TestRunFilter struct {
	SHA      string       `json:"sha,omitempty"`
	Labels   mapset.Set   `json:"labels,omitempty"`
	Aligned  *bool        `json:"aligned,omitempty"`
	From     *time.Time   `json:"from,omitempty"`
	To       *time.Time   `json:"to,omitempty"`
	MaxCount *int         `json:"maxcount,omitempty"`
	Offset   *int         `json:"offset,omitempty"` // Used for paginating with MaxCount.
	Products ProductSpecs `json:"products,omitempty"`
}

// IsDefaultQuery returns whether the params are just an empty query (or,
// the equivalent defaults of an empty query).
func (filter TestRunFilter) IsDefaultQuery() bool {
	return IsLatest(filter.SHA) &&
		(filter.Labels == nil || filter.Labels.Cardinality() < 1) &&
		(filter.Aligned == nil) &&
		(filter.From == nil) &&
		(filter.MaxCount == nil || *filter.MaxCount == 1) &&
		(len(filter.Products) < 1)
}

// OrDefault returns the current filter, or, if it is a default query, returns
// the query used by default in wpt.fyi.
func (filter TestRunFilter) OrDefault() TestRunFilter {
	return filter.OrAlignedStableRuns()
}

// OrAlignedStableRuns returns the current filter, or, if it is a default query, returns
// a query for stable runs, with an aligned SHA.
func (filter TestRunFilter) OrAlignedStableRuns() TestRunFilter {
	if !filter.IsDefaultQuery() {
		return filter
	}
	aligned := true
	filter.Aligned = &aligned
	filter.Labels = mapset.NewSetWith(StableLabel)
	return filter
}

// OrExperimentalRuns returns the current filter, or, if it is a default query, returns
// a query for the latest experimental runs.
func (filter TestRunFilter) OrExperimentalRuns() TestRunFilter {
	if !filter.IsDefaultQuery() {
		return filter
	}
	filter.Labels = mapset.NewSetWith(ExperimentalLabel)
	return filter
}

// OrAlignedExperimentalRunsExceptEdge returns the current filter, or, if it is a default
// query, returns a query for the latest experimental runs.
func (filter TestRunFilter) OrAlignedExperimentalRunsExceptEdge() TestRunFilter {
	if !filter.IsDefaultQuery() {
		return filter
	}
	aligned := true
	filter.Aligned = &aligned
	filter.Products = GetDefaultProducts()
	for i := range filter.Products {
		if filter.Products[i].BrowserName != "edge" {
			filter.Products[i].Labels = mapset.NewSetWith("experimental")
		}
	}
	return filter
}

// MasterOnly returns the current filter, ensuring it has with the master-only
// restriction (a label of "master").
func (filter TestRunFilter) MasterOnly() TestRunFilter {
	if filter.Labels == nil {
		filter.Labels = mapset.NewSet()
	}
	filter.Labels.Add(MasterLabel)
	return filter
}

// IsDefaultProducts returns whether the params products are empty, or the
// equivalent of the default product set.
func (filter TestRunFilter) IsDefaultProducts() bool {
	if len(filter.Products) == 0 {
		return true
	}
	def := GetDefaultProducts()
	if len(filter.Products) != len(def) {
		return false
	}
	for i := range def {
		if def[i] != filter.Products[i] {
			return false
		}
	}
	return true
}

// GetProductsOrDefault parses the 'products' (and legacy 'browsers') params, returning
// the ordered list of products to include, or a default list.
func (filter TestRunFilter) GetProductsOrDefault() (products ProductSpecs) {
	return filter.Products.OrDefault()
}

// ToQuery converts the filter set to a url.Values (set of query params).
func (filter TestRunFilter) ToQuery() (q url.Values) {
	u := url.URL{}
	q = u.Query()
	if !IsLatest(filter.SHA) {
		q.Set("sha", filter.SHA)
	}
	if filter.Labels != nil && filter.Labels.Cardinality() > 0 {
		for label := range filter.Labels.Iter() {
			q.Add("label", label.(string))
		}
	}
	if len(filter.Products) > 0 {
		for _, p := range filter.Products {
			q.Add("product", p.String())
		}
	}
	if filter.Aligned != nil {
		q.Set("aligned", strconv.FormatBool(*filter.Aligned))
	}
	if filter.MaxCount != nil {
		q.Set("max-count", fmt.Sprintf("%v", *filter.MaxCount))
	}
	if filter.From != nil {
		q.Set("from", filter.From.Format(time.RFC3339))
	}
	if filter.To != nil {
		q.Set("to", filter.From.Format(time.RFC3339))
	}
	return q
}

// NextPage returns a filter for the next page of results that
// would match the current filter, based on the given results that were
// loaded.
func (filter TestRunFilter) NextPage(loadedRuns TestRunsByProduct) *TestRunFilter {
	// TODO: Handle to/from ranges.
	if filter.MaxCount != nil {
		// We only have another page if N results were returned for a max of N.
		anyMaxedOut := false
		for _, v := range loadedRuns {
			if len(v) >= *filter.MaxCount {
				anyMaxedOut = true
			}
		}
		if anyMaxedOut {
			offset := *filter.MaxCount
			if filter.Offset != nil {
				offset += *filter.Offset
			}
			filter.Offset = &offset
			return &filter
		}
	}
	return nil
}

// Token returns a base64 encoded copy of the filter.
func (filter TestRunFilter) Token() (string, error) {
	bytes, err := json.Marshal(filter)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
