package modes_test

import (
	"strings"
	"testing"

	"fuzzer/detector"
	"fuzzer/modes"
	"fuzzer/wordlist"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func loadSmall(t *testing.T) *wordlist.Wordlist {
	t.Helper()
	wl, err := wordlist.Load("testdata/small_wordlist.txt")
	if err != nil {
		t.Fatalf("load small_wordlist.txt: %v", err)
	}
	return wl
}

func loadUsers(t *testing.T) *wordlist.Wordlist {
	t.Helper()
	wl, err := wordlist.Load("testdata/users.txt")
	if err != nil {
		t.Fatalf("load users.txt: %v", err)
	}
	return wl
}

func loadPasswords(t *testing.T) *wordlist.Wordlist {
	t.Helper()
	wl, err := wordlist.Load("testdata/passwords.txt")
	if err != nil {
		t.Fatalf("load passwords.txt: %v", err)
	}
	return wl
}

func mustDetect(t *testing.T, raw string, wantCount int) []detector.FuzzPoint {
	t.Helper()
	points := detector.Detect(raw)
	if len(points) != wantCount {
		t.Fatalf("detector.Detect: got %d points, want %d", len(points), wantCount)
	}
	return points
}

// ── Sniper ────────────────────────────────────────────────────────────────────

func TestSniper_SinglePoint_ResultCount(t *testing.T) {
	raw := "password=§secret§"
	points := mustDetect(t, raw, 1)
	wl := loadSmall(t)

	results := modes.Sniper(raw, points, wl)
	want := 1 * wl.Count()
	if len(results) != want {
		t.Errorf("Sniper 1-point result count: got %d, want %d", len(results), want)
	}
}

func TestSniper_TwoPoints_ResultCount(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.Sniper(raw, points, wl)
	want := len(points) * wl.Count()
	if len(results) != want {
		t.Errorf("Sniper 2-point result count: got %d, want %d", len(results), want)
	}
}

func TestSniper_FuzzPointIndex_FirstGroup(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.Sniper(raw, points, wl)
	for i := 0; i < wl.Count(); i++ {
		if results[i].FuzzPointIndex != 0 {
			t.Errorf("results[%d].FuzzPointIndex: got %d, want 0", i, results[i].FuzzPointIndex)
		}
	}
}

func TestSniper_FuzzPointIndex_SecondGroup(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.Sniper(raw, points, wl)
	for i := wl.Count(); i < 2*wl.Count(); i++ {
		if results[i].FuzzPointIndex != 1 {
			t.Errorf("results[%d].FuzzPointIndex: got %d, want 1", i, results[i].FuzzPointIndex)
		}
	}
}

func TestSniper_FirstPayload_InjectedCorrectly(t *testing.T) {
	raw := "password=§secret§"
	points := mustDetect(t, raw, 1)
	wl := loadSmall(t)

	results := modes.Sniper(raw, points, wl)
	if len(results) == 0 {
		t.Fatal("Sniper returned no results")
	}
	if !strings.Contains(results[0].Request, "alpha") {
		t.Errorf("results[0].Request should contain \"alpha\", got: %q", results[0].Request)
	}
}

func TestSniper_ActivePoint_GetsPayload_OtherKeepsDefault(t *testing.T) {
	raw := "user=§carlos§&pass=§secret§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.Sniper(raw, points, wl)
	req := results[0].Request
	if !strings.Contains(req, "alpha") {
		t.Errorf("first result should contain active payload \"alpha\", got: %q", req)
	}
	if !strings.Contains(req, "secret") {
		t.Errorf("first result should keep default \"secret\" for inactive point, got: %q", req)
	}
	if strings.Contains(req, "§") {
		t.Errorf("result still contains § markers: %q", req)
	}
}

func TestSniper_AllPayloadsPresent(t *testing.T) {
	raw := "x=§val§"
	points := mustDetect(t, raw, 1)
	wl := loadSmall(t)

	results := modes.Sniper(raw, points, wl)
	if len(results) != 3 {
		t.Fatalf("want 3 results, got %d", len(results))
	}
	words := []string{"alpha", "beta", "gamma"}
	for i, want := range words {
		if !strings.Contains(results[i].Request, want) {
			t.Errorf("results[%d].Request should contain %q, got: %q", i, want, results[i].Request)
		}
	}
}

func TestSniper_NoMarkersInResult(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.Sniper(raw, points, wl)
	for i, r := range results {
		if strings.Contains(r.Request, "§") {
			t.Errorf("results[%d].Request still contains § markers: %q", i, r.Request)
		}
	}
}

func TestSniper_PayloadsMap_Populated(t *testing.T) {
	raw := "x=§val§"
	points := mustDetect(t, raw, 1)
	wl := loadSmall(t)

	results := modes.Sniper(raw, points, wl)
	if len(results) == 0 {
		t.Fatal("no results")
	}
	if len(results[0].Payloads) == 0 {
		t.Error("Result.Payloads map is empty")
	}
}

// ── BatteringRam ──────────────────────────────────────────────────────────────

func TestBatteringRam_OnePoint_ResultCount(t *testing.T) {
	raw := "password=§secret§"
	points := mustDetect(t, raw, 1)
	wl := loadSmall(t)

	results := modes.BatteringRam(raw, points, wl)
	if len(results) != wl.Count() {
		t.Errorf("BatteringRam 1-point count: got %d, want %d", len(results), wl.Count())
	}
}

func TestBatteringRam_TwoPoints_ResultCount(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.BatteringRam(raw, points, wl)
	if len(results) != wl.Count() {
		t.Errorf("BatteringRam 2-point count: got %d, want %d (M words, not N×M)",
			len(results), wl.Count())
	}
}

func TestBatteringRam_SamePayloadInAllPoints(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.BatteringRam(raw, points, wl)
	if len(results) == 0 {
		t.Fatal("no results")
	}
	req := results[0].Request
	if strings.Count(req, "alpha") != 2 {
		t.Errorf("first result should have \"alpha\" in both positions, got: %q", req)
	}
}

func TestBatteringRam_AllWordsCovered(t *testing.T) {
	raw := "x=§v§"
	points := mustDetect(t, raw, 1)
	wl := loadSmall(t)

	results := modes.BatteringRam(raw, points, wl)
	if len(results) != 3 {
		t.Fatalf("want 3 results, got %d", len(results))
	}
	words := []string{"alpha", "beta", "gamma"}
	for i, want := range words {
		if !strings.Contains(results[i].Request, want) {
			t.Errorf("results[%d].Request should contain %q, got: %q", i, want, results[i].Request)
		}
	}
}

func TestBatteringRam_FuzzPointIndex_IsMinusOne(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.BatteringRam(raw, points, wl)
	for i, r := range results {
		if r.FuzzPointIndex != -1 {
			t.Errorf("results[%d].FuzzPointIndex: got %d, want -1", i, r.FuzzPointIndex)
		}
	}
}

func TestBatteringRam_NoMarkersInResult(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.BatteringRam(raw, points, wl)
	for i, r := range results {
		if strings.Contains(r.Request, "§") {
			t.Errorf("results[%d].Request still contains § markers: %q", i, r.Request)
		}
	}
}

func TestBatteringRam_FewerResultsThanSniper(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	sniperResults := modes.Sniper(raw, points, wl)
	ramResults := modes.BatteringRam(raw, points, wl)
	if len(ramResults) >= len(sniperResults) {
		t.Errorf("BatteringRam (%d) should produce fewer results than Sniper (%d)",
			len(ramResults), len(sniperResults))
	}
}

// ── Pitchfork ─────────────────────────────────────────────────────────────────

func TestPitchfork_EqualLists_ResultCount(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.Pitchfork(raw, points, []*wordlist.Wordlist{wl, wl})
	if len(results) != 3 {
		t.Errorf("Pitchfork equal lists: got %d results, want 3", len(results))
	}
}

func TestPitchfork_StopsAtShortestList(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	users := loadUsers(t)
	passwords := loadPasswords(t)

	results := modes.Pitchfork(raw, points, []*wordlist.Wordlist{users, passwords})
	if len(results) != 6 {
		t.Errorf("Pitchfork stops at shortest: got %d results, want 6", len(results))
	}
}

func TestPitchfork_SinglePoint(t *testing.T) {
	raw := "x=§v§"
	points := mustDetect(t, raw, 1)
	wl := loadSmall(t)

	results := modes.Pitchfork(raw, points, []*wordlist.Wordlist{wl})
	if len(results) != 3 {
		t.Errorf("Pitchfork 1 point: got %d results, want 3", len(results))
	}
}

func TestPitchfork_LockStep_FirstPair(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	users := wordlist.FromSlice([]string{"carlos", "peter", "root"})
	passwords := wordlist.FromSlice([]string{"pass123", "qwerty", "secret"})

	results := modes.Pitchfork(raw, points, []*wordlist.Wordlist{users, passwords})
	req := results[0].Request
	if !strings.Contains(req, "carlos") {
		t.Errorf("results[0]: should contain \"carlos\", got: %q", req)
	}
	if !strings.Contains(req, "pass123") {
		t.Errorf("results[0]: should contain \"pass123\", got: %q", req)
	}
}

func TestPitchfork_LockStep_SecondPair(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	users := wordlist.FromSlice([]string{"carlos", "peter", "root"})
	passwords := wordlist.FromSlice([]string{"pass123", "qwerty", "secret"})

	results := modes.Pitchfork(raw, points, []*wordlist.Wordlist{users, passwords})
	if len(results) < 2 {
		t.Fatalf("need at least 2 results, got %d", len(results))
	}
	req := results[1].Request
	if !strings.Contains(req, "peter") {
		t.Errorf("results[1]: should contain \"peter\", got: %q", req)
	}
	if !strings.Contains(req, "qwerty") {
		t.Errorf("results[1]: should contain \"qwerty\", got: %q", req)
	}
}

func TestPitchfork_LockStep_AllPairs(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	users := wordlist.FromSlice([]string{"carlos", "peter", "root"})
	passwords := wordlist.FromSlice([]string{"pass123", "qwerty", "secret"})

	results := modes.Pitchfork(raw, points, []*wordlist.Wordlist{users, passwords})
	if len(results) != 3 {
		t.Fatalf("want 3 results, got %d", len(results))
	}
	pairs := [][2]string{
		{"carlos", "pass123"},
		{"peter", "qwerty"},
		{"root", "secret"},
	}
	for i, pair := range pairs {
		req := results[i].Request
		if !strings.Contains(req, pair[0]) || !strings.Contains(req, pair[1]) {
			t.Errorf("results[%d]: want user=%q pass=%q, got: %q", i, pair[0], pair[1], req)
		}
	}
}

func TestPitchfork_FuzzPointIndex_IsMinusOne(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.Pitchfork(raw, points, []*wordlist.Wordlist{wl, wl})
	for i, r := range results {
		if r.FuzzPointIndex != -1 {
			t.Errorf("results[%d].FuzzPointIndex: got %d, want -1", i, r.FuzzPointIndex)
		}
	}
}

func TestPitchfork_MismatchedLengths_ReturnsNil(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.Pitchfork(raw, points, []*wordlist.Wordlist{wl})
	if results != nil {
		t.Errorf("mismatched wordlist count should return nil, got %d results", len(results))
	}
}

func TestPitchfork_NoMarkersInResult(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := mustDetect(t, raw, 2)
	wl := loadSmall(t)

	results := modes.Pitchfork(raw, points, []*wordlist.Wordlist{wl, wl})
	for i, r := range results {
		if strings.Contains(r.Request, "§") {
			t.Errorf("results[%d].Request still contains § markers: %q", i, r.Request)
		}
	}
}
