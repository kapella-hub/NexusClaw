package sentryapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client is a public SDK client for the NexusClaw Sentry API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Sentry API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// doRequest performs an authenticated HTTP request and unmarshals the response.
func (c *Client) doRequest(ctx context.Context, method, path string, body, result any) error {
	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("sentryapi: marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("sentryapi: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sentryapi: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		if errBody.Error != "" {
			return fmt.Errorf("sentryapi: %s %s: %d %s", method, path, resp.StatusCode, errBody.Error)
		}
		return fmt.Errorf("sentryapi: %s %s: %d", method, path, resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("sentryapi: decode response: %w", err)
		}
	}
	return nil
}

// ListAudit returns all audit log entries.
func (c *Client) ListAudit(ctx context.Context) ([]AuditEntry, error) {
	var entries []AuditEntry
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/sentry/audit", nil, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// ListRules returns all sentry rules.
func (c *Client) ListRules(ctx context.Context) ([]Rule, error) {
	var rules []Rule
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/sentry/rules", nil, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

// CreateRule creates a new sentry rule.
func (c *Client) CreateRule(ctx context.Context, rule *Rule) (*Rule, error) {
	var created Rule
	if err := c.doRequest(ctx, http.MethodPost, "/api/v1/sentry/rules", rule, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateRule updates an existing sentry rule by ID.
func (c *Client) UpdateRule(ctx context.Context, id string, rule *Rule) (*Rule, error) {
	var updated Rule
	if err := c.doRequest(ctx, http.MethodPut, "/api/v1/sentry/rules/"+id, rule, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

// DeleteRule deletes a sentry rule by ID.
func (c *Client) DeleteRule(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, "/api/v1/sentry/rules/"+id, nil, nil)
}

// GetBudget returns the current user's budget cap.
func (c *Client) GetBudget(ctx context.Context) (*BudgetCap, error) {
	var budget BudgetCap
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/sentry/budget", nil, &budget); err != nil {
		return nil, err
	}
	return &budget, nil
}

// UpdateBudget updates the current user's budget cap.
func (c *Client) UpdateBudget(ctx context.Context, budget *BudgetCap) (*BudgetCap, error) {
	var updated BudgetCap
	if err := c.doRequest(ctx, http.MethodPut, "/api/v1/sentry/budget", budget, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}
