package parser_test

import (
	"strings"
	"testing"

	"belch/parser"
)

// ── ParseRequestFromFile ────────────────────────────────────────────────────

func TestParseFile_Method(t *testing.T) {
	req, err := parser.ParseRequestFromFile("../request.req")
	if err != nil {
		t.Fatalf("ParseRequestFromFile returned error: %v", err)
	}
	if req.Method != "POST" {
		t.Errorf("Method: got %q, want %q", req.Method, "POST")
	}
}

func TestParseFile_Path(t *testing.T) {
	req, err := parser.ParseRequestFromFile("../request.req")
	if err != nil {
		t.Fatalf("ParseRequestFromFile returned error: %v", err)
	}
	if req.Path != "/login" {
		t.Errorf("Path: got %q, want %q", req.Path, "/login")
	}
}

func TestParseFile_Proto(t *testing.T) {
	req, err := parser.ParseRequestFromFile("../request.req")
	if err != nil {
		t.Fatalf("ParseRequestFromFile returned error: %v", err)
	}
	if req.Proto != "HTTP/2" {
		t.Errorf("Proto: got %q, want %q", req.Proto, "HTTP/2")
	}
}

func TestParseFile_Host(t *testing.T) {
	req, err := parser.ParseRequestFromFile("../request.req")
	if err != nil {
		t.Fatalf("ParseRequestFromFile returned error: %v", err)
	}
	want := "0ad700a2036d5e5c806ad046007200ea.web-security-academy.net"
	if req.Host != want {
		t.Errorf("Host: got %q, want %q", req.Host, want)
	}
}

func TestParseFile_ContentTypeHeader(t *testing.T) {
	req, err := parser.ParseRequestFromFile("../request.req")
	if err != nil {
		t.Fatalf("ParseRequestFromFile returned error: %v", err)
	}
	got, ok := req.Headers["Content-Type"]
	if !ok {
		t.Fatal("Headers: key \"Content-Type\" is missing")
	}
	want := "application/x-www-form-urlencoded"
	if got != want {
		t.Errorf("Content-Type: got %q, want %q", got, want)
	}
}

func TestParseFile_CookieHeader(t *testing.T) {
	req, err := parser.ParseRequestFromFile("../request.req")
	if err != nil {
		t.Fatalf("ParseRequestFromFile returned error: %v", err)
	}
	cookie, ok := req.Headers["Cookie"]
	if !ok {
		t.Fatal("Headers: key \"Cookie\" is missing")
	}
	if !strings.HasPrefix(cookie, "session=") {
		t.Errorf("Cookie header should start with \"session=\", got %q", cookie)
	}
}

func TestParseFile_Body(t *testing.T) {
	req, err := parser.ParseRequestFromFile("../request.req")
	if err != nil {
		t.Fatalf("ParseRequestFromFile returned error: %v", err)
	}
	want := "csrf=i14OYgo57DheDz28pFsky14ks2YQr8nO&username=prueba&password=prueba"
	if req.Body != want {
		t.Errorf("Body:\n  got  %q\n  want %q", req.Body, want)
	}
}

func TestParseFile_HeaderCount(t *testing.T) {
	req, err := parser.ParseRequestFromFile("../request.req")
	if err != nil {
		t.Fatalf("ParseRequestFromFile returned error: %v", err)
	}
	// Host is stored in Request.Host, not duplicated in Headers.
	const wantCount = 19
	if len(req.Headers) != wantCount {
		t.Errorf("Header count: got %d, want %d", len(req.Headers), wantCount)
	}
}

// ── ParseRequestFromString ──────────────────────────────────────────────────

func TestParseString_GetRequest(t *testing.T) {
	raw := "GET /hello HTTP/1.1\r\nHost: example.com\r\nX-Custom: my-value\r\n\r\n"
	req, err := parser.ParseRequestFromString(raw)
	if err != nil {
		t.Fatalf("ParseRequestFromString error: %v", err)
	}
	if req.Method != "GET" {
		t.Errorf("Method: got %q, want GET", req.Method)
	}
	if req.Path != "/hello" {
		t.Errorf("Path: got %q, want /hello", req.Path)
	}
	if req.Host != "example.com" {
		t.Errorf("Host: got %q, want example.com", req.Host)
	}
	if req.Headers["X-Custom"] != "my-value" {
		t.Errorf("X-Custom: got %q, want my-value", req.Headers["X-Custom"])
	}
}

func TestParseString_WithBody(t *testing.T) {
	raw := "POST /submit HTTP/1.1\r\nHost: example.com\r\nContent-Type: application/x-www-form-urlencoded\r\n\r\nfoo=bar&baz=qux"
	req, err := parser.ParseRequestFromString(raw)
	if err != nil {
		t.Fatalf("ParseRequestFromString error: %v", err)
	}
	if req.Body != "foo=bar&baz=qux" {
		t.Errorf("Body: got %q, want foo=bar&baz=qux", req.Body)
	}
}

func TestParseString_EmptyBody(t *testing.T) {
	raw := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
	req, err := parser.ParseRequestFromString(raw)
	if err != nil {
		t.Fatalf("ParseRequestFromString error: %v", err)
	}
	if req.Body != "" {
		t.Errorf("Body should be empty for bodyless request, got %q", req.Body)
	}
}

// ── String() round-trip ─────────────────────────────────────────────────────

func TestString_RoundTrip_Method(t *testing.T) {
	orig, err := parser.ParseRequestFromFile("../request.req")
	if err != nil {
		t.Fatalf("ParseRequestFromFile error: %v", err)
	}
	raw := orig.String()
	if raw == "" {
		t.Fatal("String() returned empty string")
	}
	reparsed, err := parser.ParseRequestFromString(raw)
	if err != nil {
		t.Fatalf("ParseRequestFromString on round-trip result error: %v", err)
	}
	if orig.Method != reparsed.Method {
		t.Errorf("round-trip Method: %q → %q", orig.Method, reparsed.Method)
	}
}

func TestString_RoundTrip_Body(t *testing.T) {
	orig, err := parser.ParseRequestFromFile("../request.req")
	if err != nil {
		t.Fatalf("ParseRequestFromFile error: %v", err)
	}
	raw := orig.String()
	reparsed, err := parser.ParseRequestFromString(raw)
	if err != nil {
		t.Fatalf("ParseRequestFromString error: %v", err)
	}
	if orig.Body != reparsed.Body {
		t.Errorf("round-trip Body:\n  orig     %q\n  reparsed %q", orig.Body, reparsed.Body)
	}
}

func TestString_RoundTrip_Headers(t *testing.T) {
	orig, err := parser.ParseRequestFromFile("../request.req")
	if err != nil {
		t.Fatalf("ParseRequestFromFile error: %v", err)
	}
	raw := orig.String()
	reparsed, err := parser.ParseRequestFromString(raw)
	if err != nil {
		t.Fatalf("ParseRequestFromString error: %v", err)
	}
	for key, want := range orig.Headers {
		if got := reparsed.Headers[key]; got != want {
			t.Errorf("round-trip header %q: got %q, want %q", key, got, want)
		}
	}
}

// ── Error handling ──────────────────────────────────────────────────────────

func TestParseFile_NonExistent(t *testing.T) {
	_, err := parser.ParseRequestFromFile("nonexistent_file_xyz.req")
	if err == nil {
		t.Error("expected an error for a non-existent file, got nil")
	}
}
