package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

const (
	ContentTypeJSON = "application/json"
)

type Client struct {
	url        string
	httpClient *http.Client
}

func New(host string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		url:        host,
		httpClient: httpClient,
	}
}

func (c *Client) SendRequest(ctx context.Context, data interface{}, options ...func(*http.Request)) ([]byte, error) {
	var jsonValue []byte

	if data != nil {
		var err error

		jsonValue, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("can't marshal input data: %w", err)
		}
	}

	return c.request(ctx, jsonValue, options...)
}

func (c *Client) request(ctx context.Context, jsonValue []byte, options ...func(*http.Request)) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("can't wrap new request: %w", err)
	}

	request.URL, err = url.Parse(c.url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	for _, option := range options {
		option(request)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("can't do new request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read response body: %w", err)
	}

	return body, nil
}

func MethodOption(method string) func(req *http.Request) {
	return func(req *http.Request) {
		req.Method = method
	}
}

func AuthOption(authHeader string) func(req *http.Request) {
	return func(req *http.Request) {
		req.Header.Set("Authorization", authHeader)
	}
}

func HeaderOption(key, value string) func(req *http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func ContentTypeOption(contentType string) func(req *http.Request) {
	return func(req *http.Request) {
		req.Header.Set("Content-Type", contentType)
	}
}

func UserAgentOption(userAgent string) func(req *http.Request) {
	return func(req *http.Request) {
		req.Header.Set("User-Agent", userAgent)
	}
}

func ParamOption(key, value string) func(req *http.Request) {
	return func(req *http.Request) {
		q := req.URL.Query()
		q.Add(key, value)
		req.URL.RawQuery = q.Encode()
	}
}

func PathOption(subPath string) func(req *http.Request) {
	return func(req *http.Request) {
		req.URL.Path = path.Join(req.URL.Path, subPath)
	}
}
