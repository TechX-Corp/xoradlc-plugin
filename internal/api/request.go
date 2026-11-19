package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestOpts configures Request.
type RequestOpts struct {
	Body    any
	Form    map[string]string
	Token   string
	Timeout time.Duration
}

// Request performs an HTTP call and returns (status, json | text | nil). Network error: status 0.
func Request(method, rawURL string, opts RequestOpts) (int, any) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	var body io.Reader
	headers := http.Header{}
	headers.Set("Accept", "application/json")

	if opts.Form != nil {
		values := url.Values{}
		for k, v := range opts.Form {
			values.Set(k, v)
		}
		body = strings.NewReader(values.Encode())
		headers.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if opts.Body != nil {
		b, err := json.Marshal(opts.Body)
		if err != nil {
			return 0, err.Error()
		}
		body = bytes.NewReader(b)
		headers.Set("Content-Type", "application/json")
	}
	if opts.Token != "" {
		headers.Set("Authorization", "Bearer "+opts.Token)
	}

	req, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		return 0, err.Error()
	}
	req.Header = headers

	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}
	if len(raw) == 0 {
		return resp.StatusCode, nil
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return resp.StatusCode, string(raw)
	}
	return resp.StatusCode, decoded
}
