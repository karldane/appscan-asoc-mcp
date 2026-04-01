package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    string
	keyID      string
	keySecret  string
	httpClient *http.Client
}

func New(baseURL, keyID, keySecret string, timeoutSeconds int) *Client {
	return &Client{
		baseURL:   baseURL,
		keyID:     keyID,
		keySecret: keySecret,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

func (c *Client) AuthHeader() string {
	return fmt.Sprintf("%s:%s", c.keyID, c.keySecret)
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) Do(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Api-Key", c.AuthHeader())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Debug logging
	fmt.Printf("[DEBUG] %s %s\n", method, c.baseURL+path)
	fmt.Printf("[DEBUG] X-Api-Key: %s\n", c.AuthHeader())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	// Debug logging
	fmt.Printf("[DEBUG] Response Status: %d\n", resp.StatusCode)
	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("[DEBUG] Response Body: %s\n", string(respBody))
		// Reset body for later use
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
	}

	return resp, nil
}

func (c *Client) Get(path string) (*http.Response, error) {
	return c.Do("GET", path, nil)
}

func (c *Client) Post(path string, body interface{}) (*http.Response, error) {
	return c.Do("POST", path, body)
}

func (c *Client) Put(path string, body interface{}) (*http.Response, error) {
	return c.Do("PUT", path, body)
}

func (c *Client) Delete(path string) (*http.Response, error) {
	return c.Do("DELETE", path, nil)
}

func (c *Client) DoCustom(method, path string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Api-Key", c.AuthHeader())
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.httpClient.Do(req)
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ParseError(resp *http.Response) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read error body: %w", err)
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
		return fmt.Errorf("%s: %s", errResp.Code, errResp.Message)
	}

	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
}

func IsNotFound(resp *http.Response) bool {
	return resp.StatusCode == 404
}

func IsUnauthorized(resp *http.Response) bool {
	return resp.StatusCode == 401
}

func IsForbidden(resp *http.Response) bool {
	return resp.StatusCode == 403
}

func BuildQuery(path string, params map[string]string) string {
	if len(params) == 0 {
		return path
	}
	q := url.Values{}
	for k, v := range params {
		q.Add(k, v)
	}
	return path + "?" + q.Encode()
}

func DecodeJSON(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ParseError(resp)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	return json.Unmarshal(body, target)
}
