package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func oauthRequest(method, rawURL string, form map[string]string, token string) (int, any) {
	timeout := 30 * time.Second
	client := &http.Client{Timeout: timeout}
	var body io.Reader
	headers := http.Header{}
	headers.Set("Accept", "application/json")
	if form != nil {
		values := url.Values{}
		for k, v := range form {
			values.Set(k, v)
		}
		body = strings.NewReader(values.Encode())
		headers.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
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

func oauthRequestJSON(method, rawURL string, jsonBody any, token string) (int, any) {
	client := &http.Client{Timeout: 30 * time.Second}
	var body io.Reader
	if jsonBody != nil {
		b, err := json.Marshal(jsonBody)
		if err != nil {
			return 0, err.Error()
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Accept", "application/json")
	if jsonBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
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