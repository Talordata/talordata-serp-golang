package talordata

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	DefaultBaseURL  = "https://serpapi.talordata.net"
	DefaultPath     = "/serp/v1/request"
	DefaultTimeout  = 30 * time.Second
	defaultUserAgent = "talordata-serp-golang/0.1.0"
)

var ErrAPITokenNotProvided = errors.New("please provide apiToken or set TALORDATA_API_TOKEN")

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http %d", e.StatusCode)
}

type Client struct {
	APIToken   string
	Timeout    time.Duration
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(apiToken string) *Client {
	resolved := resolveToken(apiToken)
	return &Client{
		APIToken: resolved,
		Timeout:  DefaultTimeout,
		BaseURL:  DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

func Search(params map[string]interface{}) (interface{}, error) {
	return NewClient("").Search(params)
}

func SearchJSON(params map[string]interface{}) (interface{}, error) {
	return NewClient("").SearchJSON(params)
}

func SearchHTML(params map[string]interface{}) (string, error) {
	return NewClient("").SearchHTML(params)
}

func RawSearch(params map[string]interface{}) (string, error) {
	return NewClient("").RawSearch(params)
}

func (c *Client) Search(params map[string]interface{}) (interface{}, error) {
	payload := cloneMap(params)
	if _, ok := payload["json"]; !ok {
		payload["json"] = "1"
	}
	return c.execute(payload)
}

func (c *Client) SearchJSON(params map[string]interface{}) (interface{}, error) {
	return c.Search(params)
}

func (c *Client) SearchHTML(params map[string]interface{}) (string, error) {
	payload := cloneMap(params)
	payload["json"] = "3"

	result, raw, err := c.executeWithRaw(payload)
	if err != nil {
		return "", err
	}

	if html, ok := result.(string); ok {
		return html, nil
	}
	return raw, nil
}

func (c *Client) RawSearch(params map[string]interface{}) (string, error) {
	_, raw, err := c.executeWithRaw(params)
	return raw, err
}

func (c *Client) execute(params map[string]interface{}) (interface{}, error) {
	result, _, err := c.executeWithRaw(params)
	return result, err
}

func (c *Client) executeWithRaw(params map[string]interface{}) (interface{}, string, error) {
	raw, err := c.request(http.MethodPost, DefaultPath, params)
	if err != nil {
		return nil, "", err
	}
	result := unwrapPayload(raw)
	return result, raw, nil
}

func (c *Client) request(method, path string, params map[string]interface{}) (string, error) {
	token := resolveToken(c.APIToken)
	if token == "" {
		return "", ErrAPITokenNotProvided
	}

	baseURL := strings.TrimRight(c.BaseURL, "/")
	requestURL := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		requestURL = baseURL + path
	}

	form := url.Values{}
	for key, value := range normalizeParams(params) {
		form.Set(key, value)
	}

	req, err := http.NewRequest(method, requestURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json, text/html;q=0.9, */*;q=0.8")
	req.Header.Set("User-Agent", defaultUserAgent)

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: c.Timeout}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	text := string(body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &HTTPError{StatusCode: resp.StatusCode, Body: text}
	}
	return text, nil
}

func resolveToken(apiToken string) string {
	if apiToken != "" {
		return apiToken
	}
	if value := os.Getenv("TALORDATA_API_TOKEN"); value != "" {
		return value
	}
	return os.Getenv("TALORDATA_SERP_API_TOKEN")
}

func normalizeParams(params map[string]interface{}) map[string]string {
	normalized := map[string]string{}
	for key, value := range params {
		if value == nil {
			continue
		}
		switch typed := value.(type) {
		case bool:
			if typed {
				normalized[key] = "1"
			} else {
				normalized[key] = "0"
			}
		default:
			normalized[key] = fmt.Sprint(value)
		}
	}
	return normalized
}

func cloneMap(params map[string]interface{}) map[string]interface{} {
	if params == nil {
		return map[string]interface{}{}
	}
	cloned := make(map[string]interface{}, len(params))
	for key, value := range params {
		cloned[key] = value
	}
	return cloned
}

func unwrapPayload(raw string) interface{} {
	var payload interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return raw
	}

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return payload
	}

	data, ok := payloadMap["data"]
	if !ok {
		return payloadMap
	}

	if html, ok := data.(string); ok {
		return html
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return payloadMap
	}

	if jsonString, ok := dataMap["json"].(string); ok {
		var nested interface{}
		if err := json.Unmarshal([]byte(jsonString), &nested); err == nil {
			copied := cloneInterfaceMap(dataMap)
			copied["json"] = nested
			return copied
		}
	}

	return dataMap
}

func cloneInterfaceMap(source map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{}, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}
