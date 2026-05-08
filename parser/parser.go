package parser

import (
	"fmt"
	"os"
	"strings"
)

// Request represents a parsed HTTP request.
type Request struct {
	Method  string
	Path    string
	Proto   string
	Host    string
	Headers map[string]string
	Body    string
}

func ParseRequestFromFile(filename string) (*Request, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseRequestFromString(string(data))
}

func ParseRequestFromString(raw string) (*Request, error) {
	var headerSection, body string

	if idx := strings.Index(raw, "\r\n\r\n"); idx != -1 {
		headerSection = raw[:idx]
		body = raw[idx+4:]
	} else if idx := strings.Index(raw, "\n\n"); idx != -1 {
		headerSection = raw[:idx]
		body = raw[idx+2:]
	} else {
		headerSection = raw
	}

	body = strings.TrimRight(body, "\r\n")
	headerSection = strings.ReplaceAll(headerSection, "\r\n", "\n")
	lines := strings.Split(headerSection, "\n")

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty request")
	}

	parts := strings.Fields(lines[0])
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: %q", lines[0])
	}

	req := &Request{
		Method:  parts[0],
		Path:    parts[1],
		Proto:   parts[2],
		Headers: make(map[string]string),
		Body:    body,
	}

	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		idx := strings.Index(line, ": ")
		if idx == -1 {
			continue
		}
		key := line[:idx]
		val := line[idx+2:]
		if key == "Host" {
			req.Host = val
		} else {
			req.Headers[key] = val
		}
	}

	return req, nil
}

func (r *Request) String() string {
	var sb strings.Builder
	sb.WriteString(r.Method + " " + r.Path + " " + r.Proto + "\r\n")
	sb.WriteString("Host: " + r.Host + "\r\n")
	for k, v := range r.Headers {
		sb.WriteString(k + ": " + v + "\r\n")
	}
	sb.WriteString("\r\n")
	sb.WriteString(r.Body)
	return sb.String()
}
