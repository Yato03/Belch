// Package wordlist loads payload lists from text or CSV files.
package wordlist

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Wordlist holds the payloads loaded from a file.
type Wordlist struct {
	Words []string
	Path  string
}

func Load(path string) (*Wordlist, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", path, err)
	}
	defer f.Close()

	wl := &Wordlist{Path: path}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			wl.Words = append(wl.Words, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("load %s: %w", path, err)
	}
	return wl, nil
}

func (w *Wordlist) Count() int {
	return len(w.Words)
}

func FromSlice(words []string) *Wordlist {
	return &Wordlist{Words: words}
}
