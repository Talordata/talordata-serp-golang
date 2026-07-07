package talordata

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(baseURL string) *Client {
	client := NewClient("secret-token")
	client.BaseURL = baseURL
	client.HTTPClient = &http.Client{}
	return client
}

func TestSearchDefaultsToJSON1AndUnwrapsData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != DefaultPath {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.Form.Get("json"); got != "1" {
			t.Fatalf("expected json=1, got %q", got)
		}
		if got := r.Form.Get("q"); got != "car" {
			t.Fatalf("expected q=car, got %q", got)
		}
		if got := r.Header.Get("Origin"); got != "sdk_golang" {
			t.Fatalf("expected Origin=sdk_golang, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":0,"task_id":"task-1","data":{"search_metadata":{"status":"Success"}}}`)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.Search(map[string]interface{}{
		"engine": "google",
		"q":      "car",
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	searchMetadata, ok := resultMap["search_metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected search_metadata map, got %T", resultMap["search_metadata"])
	}
	if got := searchMetadata["status"]; got != "Success" {
		t.Fatalf("expected status Success, got %#v", got)
	}
}

func TestSearchParsesJSON2NestedJSONString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.Form.Get("json"); got != "2" {
			t.Fatalf("expected json=2, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":0,"task_id":"task-2","data":{"html":"<html>car</html>","json":"{\"search_metadata\":{\"status\":\"Success\"}}"}}`)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.Search(map[string]interface{}{
		"engine": "google",
		"json":   2,
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if got := resultMap["html"]; got != "<html>car</html>" {
		t.Fatalf("expected html payload, got %#v", got)
	}
	nested, ok := resultMap["json"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected parsed nested json map, got %T", resultMap["json"])
	}
	searchMetadata, ok := nested["search_metadata"].(map[string]interface{})
	if !ok || searchMetadata["status"] != "Success" {
		t.Fatalf("unexpected nested json payload: %#v", nested)
	}
}

func TestSearchHTMLUnwrapsHTMLString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.Form.Get("json"); got != "3" {
			t.Fatalf("expected json=3, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":0,"task_id":"task-3","data":"<html><body>car</body></html>"}`)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	html, err := client.SearchHTML(map[string]interface{}{
		"engine": "google",
		"q":      "car",
	})
	if err != nil {
		t.Fatalf("search html failed: %v", err)
	}
	if html != "<html><body>car</body></html>" {
		t.Fatalf("unexpected html: %q", html)
	}
}

func TestSearchFailsWithoutToken(t *testing.T) {
	client := &Client{
		APIToken:   "",
		BaseURL:    "https://serpapi.talordata.net",
		HTTPClient: &http.Client{},
	}

	_, err := client.Search(map[string]interface{}{"engine": "google"})
	if err == nil {
		t.Fatal("expected missing token error")
	}
	if err != ErrAPITokenNotProvided {
		t.Fatalf("expected ErrAPITokenNotProvided, got %v", err)
	}
}
