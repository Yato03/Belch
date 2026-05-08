package detector_test

import (
	"os"
	"strings"
	"testing"

	"belch/detector"
)

// ── Detect ──────────────────────────────────────────────────────────────────

func TestDetect_NoMarkers(t *testing.T) {
	points := detector.Detect("username=prueba&password=prueba")
	if len(points) != 0 {
		t.Errorf("Detect: got %d points, want 0", len(points))
	}
}

func TestDetect_SinglePoint(t *testing.T) {
	raw := "username=prueba&password=§prueba§"
	points := detector.Detect(raw)
	if len(points) != 1 {
		t.Fatalf("Detect: got %d points, want 1", len(points))
	}
	p := points[0]
	if p.Index != 0 {
		t.Errorf("FuzzPoint.Index: got %d, want 0", p.Index)
	}
	if p.Default != "prueba" {
		t.Errorf("FuzzPoint.Default: got %q, want \"prueba\"", p.Default)
	}
}

func TestDetect_TwoPoints(t *testing.T) {
	raw := "username=§prueba§&password=§prueba§"
	points := detector.Detect(raw)
	if len(points) != 2 {
		t.Fatalf("Detect: got %d points, want 2", len(points))
	}
	if points[0].Index != 0 {
		t.Errorf("points[0].Index: got %d, want 0", points[0].Index)
	}
	if points[1].Index != 1 {
		t.Errorf("points[1].Index: got %d, want 1", points[1].Index)
	}
	if points[0].Default != "prueba" {
		t.Errorf("points[0].Default: got %q, want \"prueba\"", points[0].Default)
	}
	if points[1].Default != "prueba" {
		t.Errorf("points[1].Default: got %q, want \"prueba\"", points[1].Default)
	}
}

func TestDetect_EmptyDefault(t *testing.T) {
	raw := "field=§§"
	points := detector.Detect(raw)
	if len(points) != 1 {
		t.Fatalf("Detect: got %d points, want 1", len(points))
	}
	if points[0].Default != "" {
		t.Errorf("FuzzPoint.Default for §§: got %q, want empty string", points[0].Default)
	}
}

func TestDetect_BytePositions(t *testing.T) {
	// § is U+00A7, encoded as 2 bytes (0xC2 0xA7) in UTF-8.
	raw := "prefix §hello§ suffix"
	points := detector.Detect(raw)
	if len(points) != 1 {
		t.Fatalf("Detect: got %d points, want 1", len(points))
	}
	p := points[0]
	snippet := raw[p.Start:p.End]
	want := "§hello§"
	if snippet != want {
		t.Errorf("raw[Start:End]: got %q, want %q (Start=%d, End=%d)",
			snippet, want, p.Start, p.End)
	}
	if p.Default != "hello" {
		t.Errorf("Default: got %q, want \"hello\"", p.Default)
	}
}

func TestDetect_FromSniperFile(t *testing.T) {
	content, err := os.ReadFile("testdata/sample_sniper.req")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	points := detector.Detect(string(content))
	if len(points) != 1 {
		t.Errorf("sample_sniper.req: got %d points, want 1", len(points))
	}
}

func TestDetect_FromBatteringRamFile(t *testing.T) {
	content, err := os.ReadFile("testdata/sample_battering_ram.req")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	points := detector.Detect(string(content))
	if len(points) != 2 {
		t.Errorf("sample_battering_ram.req: got %d points, want 2", len(points))
	}
}

// ── InjectPayload ────────────────────────────────────────────────────────────

func TestInject_Single(t *testing.T) {
	raw := "username=§admin§&password=secret"
	points := detector.Detect(raw)
	if len(points) != 1 {
		t.Fatalf("prerequisite Detect: got %d, want 1", len(points))
	}
	result := detector.InjectPayload(raw, points, map[int]string{0: "root"})
	want := "username=root&password=secret"
	if result != want {
		t.Errorf("InjectPayload:\n  got  %q\n  want %q", result, want)
	}
}

func TestInject_Multi(t *testing.T) {
	raw := "user=§a§&pass=§b§"
	points := detector.Detect(raw)
	if len(points) != 2 {
		t.Fatalf("prerequisite Detect: got %d, want 2", len(points))
	}
	result := detector.InjectPayload(raw, points, map[int]string{0: "carlos", 1: "letmein"})
	want := "user=carlos&pass=letmein"
	if result != want {
		t.Errorf("InjectPayload multi:\n  got  %q\n  want %q", result, want)
	}
}

func TestInject_KeepsDefault(t *testing.T) {
	raw := "user=§carlos§&pass=§secret§"
	points := detector.Detect(raw)
	if len(points) != 2 {
		t.Fatalf("prerequisite Detect: got %d, want 2", len(points))
	}
	result := detector.InjectPayload(raw, points, map[int]string{1: "newpass"})
	want := "user=carlos&pass=newpass"
	if result != want {
		t.Errorf("InjectPayload keep-default:\n  got  %q\n  want %q", result, want)
	}
}

func TestInject_EmptyPayloadsMap(t *testing.T) {
	raw := "a=§x§&b=§y§"
	points := detector.Detect(raw)
	if len(points) != 2 {
		t.Fatalf("prerequisite Detect: got %d, want 2", len(points))
	}
	result := detector.InjectPayload(raw, points, map[int]string{})
	want := "a=x&b=y"
	if result != want {
		t.Errorf("InjectPayload empty map:\n  got  %q\n  want %q", result, want)
	}
}

func TestInject_PreservesNonMarkerContent(t *testing.T) {
	raw := "PREFIX §A§ MIDDLE §B§ SUFFIX"
	points := detector.Detect(raw)
	result := detector.InjectPayload(raw, points, map[int]string{0: "X", 1: "Y"})
	if !strings.Contains(result, "PREFIX") {
		t.Error("PREFIX was lost")
	}
	if !strings.Contains(result, "MIDDLE") {
		t.Error("MIDDLE was lost")
	}
	if !strings.Contains(result, "SUFFIX") {
		t.Error("SUFFIX was lost")
	}
	want := "PREFIX X MIDDLE Y SUFFIX"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}
