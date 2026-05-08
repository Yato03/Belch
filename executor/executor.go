// Package executor sends assembled HTTP requests and captures responses.
package executor

import (
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"time"

	"belch/parser"
)

// Response holds the HTTP response captured after executing a request.
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       string
	Length     int
	Duration   time.Duration
}

// Options configures how Execute behaves.
type Options struct {
	BaseURL    string
	SkipVerify bool
	Timeout    time.Duration
}

// Execute sends req using the provided options and returns the response.
// When opts is nil, default options are used (no base URL override, 30s timeout).
func Execute(req *parser.Request, opts *Options) (*Response, error) {
	if opts == nil {
		opts = &Options{}
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: opts.SkipVerify},
	}
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	var url string
	if opts.BaseURL != "" {
		url = opts.BaseURL + req.Path
	} else {
		url = "https://" + req.Host + req.Path
	}

	httpReq, err := http.NewRequest(req.Method, url, strings.NewReader(req.Body))
	if err != nil {
		return nil, err
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	start := time.Now()
	httpResp, err := client.Do(httpReq)
	duration := time.Since(start)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes)

	headers := make(map[string]string, len(httpResp.Header))
	for k, v := range httpResp.Header {
		headers[k] = v[0]
	}

	return &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    headers,
		Body:       body,
		Length:     len(body),
		Duration:   duration,
	}, nil
}
