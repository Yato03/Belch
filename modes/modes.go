// Package modes implements Burp Suite Intruder-style attack modes.
// Each function receives the raw request template (with § markers) and returns
// the set of fully-assembled request strings to send.
package modes

import (
	"belch/detector"
	"belch/wordlist"
)

// Result holds one generated request for a single fuzzing iteration.
type Result struct {
	// Payloads maps FuzzPoint.Index → the payload used for this iteration.
	Payloads map[int]string
	// FuzzPointIndex is the active injection point for Sniper mode.
	// Set to -1 for modes where all points are active simultaneously.
	FuzzPointIndex int
	// Request is the complete raw HTTP request string with payloads injected.
	Request string
}

// Sniper iterates over each fuzz point independently.
// For N points and M words it produces N×M results:
// first all M payloads at point 0 (other points keep their default),
// then all M payloads at point 1, and so on.
func Sniper(raw string, points []detector.FuzzPoint, wl *wordlist.Wordlist) []Result {
	results := make([]Result, 0, len(points)*wl.Count())
	for i, _ := range points {
		for _, word := range wl.Words {
			payloads := map[int]string{i: word}
			results = append(results, Result{
				FuzzPointIndex: i,
				Payloads:       payloads,
				Request:        detector.InjectPayload(raw, points, payloads),
			})
		}
	}
	return results
}

// BatteringRam uses the same payload for every fuzz point simultaneously.
// For M words it always produces exactly M results regardless of point count.
func BatteringRam(raw string, points []detector.FuzzPoint, wl *wordlist.Wordlist) []Result {
	results := make([]Result, 0, wl.Count())
	for _, word := range wl.Words {
		payloads := make(map[int]string, len(points))
		for i := range points {
			payloads[i] = word
		}
		results = append(results, Result{
			FuzzPointIndex: -1,
			Payloads:       payloads,
			Request:        detector.InjectPayload(raw, points, payloads),
		})
	}
	return results
}

// Pitchfork pairs fuzz points with their own dedicated wordlist, iterating in lock-step.
// wordlists[i] feeds points[i]. Iteration stops when the shortest list is exhausted.
func Pitchfork(raw string, points []detector.FuzzPoint, wordlists []*wordlist.Wordlist) []Result {
	if len(wordlists) != len(points) {
		return nil
	}
	minLen := wordlists[0].Count()
	for _, wl := range wordlists[1:] {
		if wl.Count() < minLen {
			minLen = wl.Count()
		}
	}
	results := make([]Result, 0, minLen)
	for i := 0; i < minLen; i++ {
		payloads := make(map[int]string, len(points))
		for j := range points {
			payloads[j] = wordlists[j].Words[i]
		}
		results = append(results, Result{
			FuzzPointIndex: -1,
			Payloads:       payloads,
			Request:        detector.InjectPayload(raw, points, payloads),
		})
	}
	return results
}
