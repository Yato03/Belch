// Package detector finds and replaces § fuzz markers in raw HTTP request strings.
// Burp Suite Intruder uses § (U+00A7) as the delimiter around injection points.
package detector

// FuzzPoint represents a single § ... § injection point within a raw request string.
type FuzzPoint struct {
	// Index is the zero-based sequential position of this point (left to right).
	Index int
	// Start is the byte offset of the opening § in the raw string.
	Start int
	// End is the byte offset one past the closing §.
	// raw[Start:End] == "§<default>§"
	End int
	// Default is the original value between the § delimiters.
	Default string
}

// Detect scans raw for all § ... § pairs and returns their FuzzPoints in order.
// Returns an empty slice (never nil) when no markers are found.
func Detect(raw string) []FuzzPoint {
	points := []FuzzPoint{}
	openPos := -1
	pointIdx := 0

	for i := 0; i < len(raw)-1; i++ {
		if raw[i] == 0xc2 && raw[i+1] == 0xa7 {
			if openPos == -1 {
				openPos = i
			} else {
				points = append(points, FuzzPoint{
					Index:   pointIdx,
					Start:   openPos,
					End:     i + 2,
					Default: raw[openPos+2 : i],
				})
				pointIdx++
				openPos = -1
			}
			i++ // skip second byte of §
		}
	}
	return points
}

// InjectPayload returns a copy of raw with each fuzz marker replaced by the
// corresponding entry in payloads (keyed by FuzzPoint.Index).
// Points whose index is absent from payloads keep their Default value.
func InjectPayload(raw string, points []FuzzPoint, payloads map[int]string) string {
	result := raw
	for i := len(points) - 1; i >= 0; i-- {
		p := points[i]
		payload, ok := payloads[p.Index]
		if !ok {
			payload = p.Default
		}
		result = result[:p.Start] + payload + result[p.End:]
	}
	return result
}
