package executor_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fuzzer/executor"
	"fuzzer/parser"
)

func newReq(method, path string, headers map[string]string, body string) *parser.Request {
	if headers == nil {
		headers = make(map[string]string)
	}
	return &parser.Request{
		Method:  method,
		Path:    path,
		Proto:   "HTTP/1.1",
		Headers: headers,
		Body:    body,
	}
}

// ── StatusCode ────────────────────────────────────────────────────────────────

func TestExecute_Status200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	resp, err := executor.Execute(newReq("GET", "/", nil, ""),
		&executor.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode: got %d, want 200", resp.StatusCode)
	}
}

func TestExecute_Status404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	resp, err := executor.Execute(newReq("GET", "/notfound", nil, ""),
		&executor.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("StatusCode: got %d, want 404", resp.StatusCode)
	}
}

func TestExecute_Status302(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "/new")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	resp, err := executor.Execute(newReq("GET", "/old", nil, ""),
		&executor.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if resp.StatusCode != 302 {
		t.Errorf("StatusCode: got %d, want 302 (executor must not follow redirects)",
			resp.StatusCode)
	}
}

// ── Body ──────────────────────────────────────────────────────────────────────

func TestExecute_ResponseBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("hello world")) //nolint:errcheck
	}))
	defer srv.Close()

	resp, err := executor.Execute(newReq("GET", "/", nil, ""),
		&executor.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if resp.Body != "hello world" {
		t.Errorf("Body: got %q, want \"hello world\"", resp.Body)
	}
}

func TestExecute_ResponseLength(t *testing.T) {
	body := "hello world"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(body)) //nolint:errcheck
	}))
	defer srv.Close()

	resp, err := executor.Execute(newReq("GET", "/", nil, ""),
		&executor.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if resp.Length != len(body) {
		t.Errorf("Length: got %d, want %d", resp.Length, len(body))
	}
}

// ── POST with request body ────────────────────────────────────────────────────

func TestExecute_POST_SendsBody(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = string(b)
	}))
	defer srv.Close()

	req := newReq("POST", "/login",
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		"username=admin&password=admin")

	_, err := executor.Execute(req, &executor.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if got != "username=admin&password=admin" {
		t.Errorf("server received body: got %q, want \"username=admin&password=admin\"", got)
	}
}

func TestExecute_POST_ForwardsHeaders(t *testing.T) {
	var ct string
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ct = r.Header.Get("Content-Type")
	}))
	defer srv.Close()

	req := newReq("POST", "/",
		map[string]string{"Content-Type": "application/json"},
		`{"key":"val"}`)

	_, err := executor.Execute(req, &executor.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if ct != "application/json" {
		t.Errorf("server Content-Type: got %q, want \"application/json\"", ct)
	}
}

// ── Duration ──────────────────────────────────────────────────────────────────

func TestExecute_DurationPositive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok")) //nolint:errcheck
	}))
	defer srv.Close()

	resp, err := executor.Execute(newReq("GET", "/", nil, ""),
		&executor.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if resp.Duration <= 0 {
		t.Error("Duration should be > 0")
	}
}

// ── Response headers ──────────────────────────────────────────────────────────

func TestExecute_ResponseHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Test-Header", "fuzz-value")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	resp, err := executor.Execute(newReq("GET", "/", nil, ""),
		&executor.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if resp.Headers == nil {
		t.Fatal("Response.Headers is nil")
	}
	got := resp.Headers["X-Test-Header"]
	if got != "fuzz-value" {
		t.Errorf("X-Test-Header: got %q, want \"fuzz-value\"", got)
	}
}

// ── Method forwarding ─────────────────────────────────────────────────────────

func TestExecute_ForwardsMethod(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
	}))
	defer srv.Close()

	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		_, err := executor.Execute(newReq(method, "/", nil, ""),
			&executor.Options{BaseURL: srv.URL})
		if err != nil {
			t.Fatalf("%s: Execute error: %v", method, err)
		}
		if !strings.EqualFold(gotMethod, method) {
			t.Errorf("method forwarding: sent %s, server got %s", method, gotMethod)
		}
	}
}

// ── nil options ───────────────────────────────────────────────────────────────

func TestExecute_NilOptions_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute panicked with nil opts: %v", r)
		}
	}()
	req := newReq("GET", "/", nil, "")
	req.Host = "127.0.0.1:1"
	executor.Execute(req, nil) //nolint:errcheck
}
