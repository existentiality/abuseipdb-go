package abuseipdbgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const apiRoot = "https://api.abuseipdb.com/api/v2"

type Client struct {
	ApiKey string
}

// Create a new AbuseIPDB Client
func New(apiKey string) Client {
	return Client{ApiKey: apiKey}
}

type jsonError struct {
	Errors []struct {
		Detail string `json:"detail"`
		Status int    `json:"status"`
	} `json:"errors"`
}

type requestData struct {
	Headers map[string]string
	Body    *bytes.Buffer
}

func (c Client) sendRequest(method string, path string, query string, extraRequestData *requestData) (*http.Response, error) {
	parsedQuery, err := url.ParseQuery(query)

	if err != nil {
		return nil, fmt.Errorf("malformed query: %s", query)
	}

	// URL Encode the query
	encodedQuery := parsedQuery.Encode()
	requestPath, err := url.JoinPath(apiRoot, path)
	requestURL := fmt.Sprintf("%s?%s", requestPath, encodedQuery)

	// If specified, use the custom request body, otherwise use an empty buffer
	// Most requests do not require the use of a custom body
	var body *bytes.Buffer
	if extraRequestData != nil {
		body = extraRequestData.Body
	} else {
		body = new(bytes.Buffer)
	}

	req, err := http.NewRequest(method, requestURL, body)

	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %s", err)
	}

	// Set proper headers
	req.Header.Set("Key", c.ApiKey)
	req.Header.Set("User-Agent", "abuseipdb-go (https://github.com/existentiality/abuseipdb-go)")

	// Default Content-Type to application/x-www-form-urlencoded, can be overwritten with custom headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Apply custom headers for specific request
	if extraRequestData != nil {
		for name, content := range extraRequestData.Headers {
			req.Header.Set(name, content)
		}
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %s", err)
	}

	// Non-OK status code
	// Read the provided error and relay it to the user
	if res.StatusCode < 200 || res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		jsonBody := *new(jsonError)
		json.Unmarshal(body, &jsonBody)

		for _, v := range jsonBody.Errors {
			return nil, fmt.Errorf("%s [%d]", v.Detail, v.Status)
		}
	}

	return res, nil
}
