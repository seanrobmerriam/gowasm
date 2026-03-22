package router

// RouteContext carries the matched route's data.
type RouteContext struct {
	// Path is the current matched path, e.g. "/users/42"
	Path string

	// Params contains named URL parameters extracted from the pattern.
	// For pattern "/users/:id" matched against "/users/42": {"id": "42"}
	Params map[string]string

	// Query contains parsed query string parameters.
	// These come from the fragment query: #/search?q=foo → {"q": "foo"}
	Query map[string]string

	// Pattern is the registered pattern that matched, e.g. "/users/:id"
	Pattern string
}

// Get returns a named param or query value, checking params first.
// Returns "" if not found.
func (rc RouteContext) Get(key string) string {
	if v, ok := rc.Params[key]; ok {
		return v
	}
	if v, ok := rc.Query[key]; ok {
		return v
	}
	return ""
}

// parseQuery parses a query string into a map.
// e.g. "q=foo&page=2" → {"q": "foo", "page": "2"}
func parseQuery(qs string) map[string]string {
	result := make(map[string]string)
	if len(qs) == 0 {
		return result
	}
	for _, pair := range splitString(qs, "&") {
		kv := splitString(pair, "=")
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

// splitString splits a string by delimiter without using strings package.
func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		found := true
		for j := 0; j < len(sep); j++ {
			if s[i+j] != sep[j] {
				found = false
				break
			}
		}
		if found {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}
