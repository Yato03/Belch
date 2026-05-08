package wordlist_test

import (
	"os"
	"path/filepath"
	"testing"

	"belch/wordlist"
)

// ── Load ────────────────────────────────────────────────────────────────────

func TestLoad_Count(t *testing.T) {
	wl, err := wordlist.Load("testdata/wordlist.txt")
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if wl.Count() != 10 {
		t.Errorf("Count: got %d, want 10", wl.Count())
	}
}

func TestLoad_WordsSliceLength(t *testing.T) {
	wl, err := wordlist.Load("testdata/wordlist.txt")
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if len(wl.Words) != 10 {
		t.Errorf("len(Words): got %d, want 10", len(wl.Words))
	}
}

func TestLoad_FirstWord(t *testing.T) {
	wl, err := wordlist.Load("testdata/wordlist.txt")
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if len(wl.Words) == 0 {
		t.Fatal("Words is empty")
	}
	if wl.Words[0] != "admin" {
		t.Errorf("Words[0]: got %q, want \"admin\"", wl.Words[0])
	}
}

func TestLoad_LastWord(t *testing.T) {
	wl, err := wordlist.Load("testdata/wordlist.txt")
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if len(wl.Words) == 0 {
		t.Fatal("Words is empty")
	}
	last := wl.Words[len(wl.Words)-1]
	if last != "password123" {
		t.Errorf("last word: got %q, want \"password123\"", last)
	}
}

func TestLoad_NoEmptyWords(t *testing.T) {
	wl, err := wordlist.Load("testdata/wordlist.txt")
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	for i, w := range wl.Words {
		if w == "" {
			t.Errorf("Words[%d] is empty — blank lines must be excluded", i)
		}
	}
}

func TestLoad_NoWhitespace(t *testing.T) {
	wl, err := wordlist.Load("testdata/wordlist.txt")
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	for i, w := range wl.Words {
		if len(w) > 0 && (w[0] == ' ' || w[len(w)-1] == ' ' || w[len(w)-1] == '\r') {
			t.Errorf("Words[%d] has untrimmed whitespace: %q", i, w)
		}
	}
}

func TestLoad_PathIsStored(t *testing.T) {
	wl, err := wordlist.Load("testdata/wordlist.txt")
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if wl.Path != "testdata/wordlist.txt" {
		t.Errorf("Path: got %q, want \"testdata/wordlist.txt\"", wl.Path)
	}
}

func TestLoad_Users(t *testing.T) {
	wl, err := wordlist.Load("testdata/users.txt")
	if err != nil {
		t.Fatalf("Load users.txt error: %v", err)
	}
	if wl.Count() != 8 {
		t.Errorf("users.txt Count: got %d, want 8", wl.Count())
	}
}

func TestLoad_Passwords(t *testing.T) {
	wl, err := wordlist.Load("testdata/passwords.txt")
	if err != nil {
		t.Fatalf("Load passwords.txt error: %v", err)
	}
	if wl.Count() != 6 {
		t.Errorf("passwords.txt Count: got %d, want 6", wl.Count())
	}
}

func TestLoad_SkipsBlankLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sparse.txt")
	content := "alpha\n\nbeta\n\n\ngamma\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	wl, err := wordlist.Load(path)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if wl.Count() != 3 {
		t.Errorf("Count after blank-line skip: got %d, want 3", wl.Count())
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := wordlist.Load("testdata/does_not_exist.txt")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

// ── FromSlice ────────────────────────────────────────────────────────────────

func TestFromSlice_Count(t *testing.T) {
	wl := wordlist.FromSlice([]string{"a", "b", "c"})
	if wl.Count() != 3 {
		t.Errorf("FromSlice Count: got %d, want 3", wl.Count())
	}
}

func TestFromSlice_Words(t *testing.T) {
	words := []string{"x", "y", "z"}
	wl := wordlist.FromSlice(words)
	for i, want := range words {
		if wl.Words[i] != want {
			t.Errorf("Words[%d]: got %q, want %q", i, wl.Words[i], want)
		}
	}
}
