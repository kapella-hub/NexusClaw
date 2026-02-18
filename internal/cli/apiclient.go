package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type apiClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func newAPIClient() *apiClient {
	return &apiClient{
		baseURL: viper.GetString("server.url"),
		token:   viper.GetString("auth.token"),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *apiClient) do(method, path string, body any) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshalling request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response: %w", err)
	}

	return data, resp.StatusCode, nil
}

func (c *apiClient) get(path string) ([]byte, int, error) {
	return c.do("GET", path, nil)
}

func (c *apiClient) post(path string, body any) ([]byte, int, error) {
	return c.do("POST", path, body)
}

func (c *apiClient) put(path string, body any) ([]byte, int, error) {
	return c.do("PUT", path, body)
}

func (c *apiClient) delete(path string) ([]byte, int, error) {
	return c.do("DELETE", path, nil)
}

// checkError inspects the HTTP status code and prints an error message from the
// response body if the request was not successful. Returns true if an error
// was present.
func checkError(data []byte, status int) bool {
	if status >= 200 && status < 300 {
		return false
	}
	var errResp struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(data, &errResp) == nil && errResp.Error != "" {
		fmt.Printf("Error: %s\n", errResp.Error)
	} else {
		fmt.Printf("Error: unexpected status %d\n", status)
	}
	return true
}
