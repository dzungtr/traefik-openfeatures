package traefikopenfeatures_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	traefikopenfeatures "github.com/dzungtr/traefik-openfeatures"
)

func TestOpenFeatures(t *testing.T) {
	cfg := traefikopenfeatures.CreateConfig()
	cfg.Provider = "noop"
	cfg.ContextHeaderKeys = []string{"organization", "user"}
	cfg.UserHeader = "user"
	cfg.Service = "my-service"
	cfg.Flags = map[string]string{
		"api_v2":  "bool",
		"version": "string",
		"index":   "int",
		"meta":    "object",
	}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	feature, _ := traefikopenfeatures.New(context.TODO(), next, cfg, "traefik-openfeatures")

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	feature.ServeHTTP(recorder, req)

	assertHeader(t, req, "openfeature_api_v2", "false")
	assertHeader(t, req, "openfeature_version", "")
	assertHeader(t, req, "openfeature_meta", "{}")
	assertHeader(t, req, "openfeature_index", "0")
}

func assertHeader(t *testing.T, req *http.Request, key, expected string) {
	t.Helper()

	if req.Header.Get(key) != expected {
		t.Errorf("invalid header value %s: %s", key, req.Header.Get(key))
	} else {
		t.Logf("feature %s with correct value: %s", key, req.Header.Get(key))
	}
}
