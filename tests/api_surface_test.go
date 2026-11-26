//go:build all

package tests

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestAPISurface pins the production API surface: the route set extracted from
// the committed, generated Swagger document must be bidirectionally equal
// (exact) to the frozen expected list — a route lost anywhere fails, and so
// does a route added anywhere. It reads docs/swagger.json only and performs no
// application request: the annotations follow the handlers, so the test
// survives every code move without editing the expected list.
func TestAPISurface(t *testing.T) {
	raw, err := os.ReadFile("../docs/swagger.json")
	if err != nil {
		t.Fatalf("read swagger: %v", err)
	}
	var doc struct {
		BasePath string                         `json:"basePath"`
		Paths    map[string]map[string]struct{} `json:"paths"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse swagger: %v", err)
	}
	if doc.BasePath != "/v1" {
		t.Fatalf("unexpected basePath: %q", doc.BasePath)
	}

	// Swag paths may carry arbitrary parameter names; the frozen list speaks
	// in normalized placeholders ({id}).
	paramRe := regexp.MustCompile(`\{[^{}]*\}`)
	got := map[string]struct{}{}
	for path, ops := range doc.Paths {
		full := doc.BasePath + paramRe.ReplaceAllString(path, "{id}")
		for _, method := range []string{"get", "post", "put", "delete"} {
			if _, ok := ops[method]; ok {
				got[strings.ToUpper(method)+" "+full] = struct{}{}
			}
		}
	}

	want := map[string]struct{}{
		"POST /v1/user/login":                {},
		"POST /v1/user/register":             {},
		"GET /v1/user/logout":                {},
		"POST /v1/login":                     {},
		"POST /v1/token/refresh":             {},
		"POST /v1/article":                   {},
		"GET /v1/articles":                   {},
		"GET /v1/article/{id}":               {},
		"PUT /v1/article/{id}":               {},
		"DELETE /v1/article/{id}":            {},
		"POST /v1/admin/login":               {},
		"GET /v1/admin/logout":               {},
		"GET /v1/admin/tenants":              {},
		"POST /v1/admin/tenants/{id}/status": {},
	}

	missing := []string{}
	for route := range want {
		if _, ok := got[route]; !ok {
			missing = append(missing, route)
		}
	}
	unexpected := []string{}
	for route := range got {
		if _, ok := want[route]; !ok {
			unexpected = append(unexpected, route)
		}
	}
	if len(missing) > 0 || len(unexpected) > 0 {
		sort.Strings(missing)
		sort.Strings(unexpected)
		t.Fatalf("API surface drifted:\n  lost routes: %v\n  unexpected routes: %v", missing, unexpected)
	}
}
