package common

// Route defines the route structure
type Route struct {
	ID         string            `json:"id"`
	URI        string            `json:"uri"`
	Predicates []Predicate       `json:"predicates"`
	Filters    []Filter          `json:"filters"`
	Order      int               `json:"order"`
	Metadata   map[string]string `json:"metadata"`
}

// Predicate defines the predicate structure
type Predicate struct {
	Name string      `json:"name"`
	Args interface{} `json:"args"`
}

// Filter defines the filter structure
type Filter struct {
	Name string      `json:"name"`
	Args interface{} `json:"args"`
}
