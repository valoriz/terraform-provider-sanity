package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultBaseURL    = "https://api.sanity.io"
	DefaultAPIVersion = "v2024-01-01"
)

// Client handles communication with the Sanity.io Management API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiVersion string
	token      string
}

// Option allows configuring optional Client settings.
type Option func(*Client)

// WithBaseURL overrides the default base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithAPIVersion overrides the default API version.
func WithAPIVersion(version string) Option {
	return func(c *Client) {
		c.apiVersion = version
	}
}

// WithHTTPClient overrides the default http.Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient initializes a new Sanity API client with the given authentication token.
func NewClient(token string, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("sanity API token cannot be empty")
	}

	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    DefaultBaseURL,
		apiVersion: DefaultAPIVersion,
		token:      token,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// doRequest performs an HTTP request and unmarshals the response JSON into v.
func (c *Client) doRequest(ctx context.Context, method, urlPath string, body interface{}, v interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlPath, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "terraform-provider-sanity")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr APIError
		if err := json.Unmarshal(respBytes, &apiErr); err == nil && apiErr.Message != "" {
			apiErr.StatusCode = resp.StatusCode
			return &apiErr
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			ErrorType:  http.StatusText(resp.StatusCode),
			Message:    string(respBytes),
		}
	}

	if v != nil && len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, v); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Project Methods
// ---------------------------------------------------------------------------

// CreateProject creates a new Sanity project.
func (c *Client) CreateProject(ctx context.Context, req CreateProjectRequest) (*Project, error) {
	url := fmt.Sprintf("%s/%s/projects", c.baseURL, c.apiVersion)
	var project Project
	if err := c.doRequest(ctx, http.MethodPost, url, req, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// GetProject retrieves a project by ID.
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	url := fmt.Sprintf("%s/%s/projects/%s", c.baseURL, c.apiVersion, projectID)
	var project Project
	if err := c.doRequest(ctx, http.MethodGet, url, nil, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// UpdateProject updates an existing project.
func (c *Client) UpdateProject(ctx context.Context, projectID string, req UpdateProjectRequest) (*Project, error) {
	url := fmt.Sprintf("%s/%s/projects/%s", c.baseURL, c.apiVersion, projectID)
	var project Project
	if err := c.doRequest(ctx, http.MethodPatch, url, req, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// DeleteProject deletes a project.
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	url := fmt.Sprintf("%s/%s/projects/%s", c.baseURL, c.apiVersion, projectID)
	return c.doRequest(ctx, http.MethodDelete, url, nil, nil)
}

// ---------------------------------------------------------------------------
// Dataset Methods
// ---------------------------------------------------------------------------

// CreateDataset creates a dataset using PUT /projects/{projectId}/datasets/{name}.
func (c *Client) CreateDataset(ctx context.Context, projectID, name string, req CreateDatasetRequest) (*Dataset, error) {
	url := fmt.Sprintf("%s/%s/projects/%s/datasets/%s", c.baseURL, c.apiVersion, projectID, name)
	var dataset Dataset
	if err := c.doRequest(ctx, http.MethodPut, url, req, &dataset); err != nil {
		return nil, err
	}
	if dataset.Name == "" {
		dataset.Name = name
	}
	if dataset.ACLMode == "" {
		dataset.ACLMode = req.ACLMode
	}
	return &dataset, nil
}

// GetDataset fetches a single dataset by inspecting the dataset list.
func (c *Client) GetDataset(ctx context.Context, projectID, name string) (*Dataset, error) {
	datasets, err := c.ListDatasets(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, d := range datasets {
		if d.Name == name {
			return &d, nil
		}
	}
	return nil, &APIError{
		StatusCode: http.StatusNotFound,
		ErrorType:  "Not Found",
		Message:    fmt.Sprintf("Dataset %q not found in project %q", name, projectID),
	}
}

// ListDatasets returns all datasets belonging to a project.
func (c *Client) ListDatasets(ctx context.Context, projectID string) ([]Dataset, error) {
	url := fmt.Sprintf("%s/%s/projects/%s/datasets", c.baseURL, c.apiVersion, projectID)
	var datasets []Dataset
	if err := c.doRequest(ctx, http.MethodGet, url, nil, &datasets); err != nil {
		return nil, err
	}
	return datasets, nil
}

// DeleteDataset deletes a dataset.
func (c *Client) DeleteDataset(ctx context.Context, projectID, name string) error {
	url := fmt.Sprintf("%s/%s/projects/%s/datasets/%s", c.baseURL, c.apiVersion, projectID, name)
	return c.doRequest(ctx, http.MethodDelete, url, nil, nil)
}

// ---------------------------------------------------------------------------
// CORS Origin Methods
// ---------------------------------------------------------------------------

// CreateCorsOrigin creates a new CORS origin for a project.
func (c *Client) CreateCorsOrigin(ctx context.Context, projectID string, req CreateCorsOriginRequest) (*CorsOrigin, error) {
	url := fmt.Sprintf("%s/%s/projects/%s/cors", c.baseURL, c.apiVersion, projectID)
	var origin CorsOrigin
	if err := c.doRequest(ctx, http.MethodPost, url, req, &origin); err != nil {
		return nil, err
	}
	return &origin, nil
}

// GetCorsOrigin retrieves a CORS origin by its integer ID.
func (c *Client) GetCorsOrigin(ctx context.Context, projectID string, corsID int64) (*CorsOrigin, error) {
	origins, err := c.ListCorsOrigins(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, o := range origins {
		if o.ID == corsID {
			return &o, nil
		}
	}
	return nil, &APIError{
		StatusCode: http.StatusNotFound,
		ErrorType:  "Not Found",
		Message:    fmt.Sprintf("CORS origin %d not found in project %q", corsID, projectID),
	}
}

// ListCorsOrigins returns all CORS origins for a project.
func (c *Client) ListCorsOrigins(ctx context.Context, projectID string) ([]CorsOrigin, error) {
	url := fmt.Sprintf("%s/%s/projects/%s/cors", c.baseURL, c.apiVersion, projectID)
	var origins []CorsOrigin
	if err := c.doRequest(ctx, http.MethodGet, url, nil, &origins); err != nil {
		return nil, err
	}
	return origins, nil
}

// DeleteCorsOrigin deletes a CORS origin by its ID.
func (c *Client) DeleteCorsOrigin(ctx context.Context, projectID string, corsID int64) error {
	url := fmt.Sprintf("%s/%s/projects/%s/cors/%d", c.baseURL, c.apiVersion, projectID, corsID)
	return c.doRequest(ctx, http.MethodDelete, url, nil, nil)
}

// ---------------------------------------------------------------------------
// Webhook Methods
// ---------------------------------------------------------------------------

// webhookBaseURL returns the project-scoped webhook host: https://{projectId}.api.sanity.io
func (c *Client) webhookBaseURL(projectID string) string {
	// If custom baseURL was passed (e.g. in tests with httptest), preserve it
	if strings.Contains(c.baseURL, "127.0.0.1") || strings.Contains(c.baseURL, "localhost") {
		return c.baseURL
	}
	return fmt.Sprintf("https://%s.api.sanity.io", projectID)
}

// CreateWebhook registers a new webhook.
func (c *Client) CreateWebhook(ctx context.Context, projectID string, req CreateWebhookRequest) (*Webhook, error) {
	url := fmt.Sprintf("%s/%s/hooks/projects/%s", c.webhookBaseURL(projectID), c.apiVersion, projectID)
	var webhook Webhook
	if err := c.doRequest(ctx, http.MethodPost, url, req, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// GetWebhook retrieves a webhook by its string ID.
func (c *Client) GetWebhook(ctx context.Context, projectID, webhookID string) (*Webhook, error) {
	url := fmt.Sprintf("%s/%s/hooks/projects/%s/%s", c.webhookBaseURL(projectID), c.apiVersion, projectID, webhookID)
	var webhook Webhook
	if err := c.doRequest(ctx, http.MethodGet, url, nil, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// UpdateWebhook partially updates an existing webhook via PATCH.
func (c *Client) UpdateWebhook(ctx context.Context, projectID, webhookID string, req UpdateWebhookRequest) (*Webhook, error) {
	url := fmt.Sprintf("%s/%s/hooks/projects/%s/%s", c.webhookBaseURL(projectID), c.apiVersion, projectID, webhookID)
	var webhook Webhook
	if err := c.doRequest(ctx, http.MethodPatch, url, req, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// DeleteWebhook deletes a webhook by ID.
func (c *Client) DeleteWebhook(ctx context.Context, projectID, webhookID string) error {
	url := fmt.Sprintf("%s/%s/hooks/projects/%s/%s", c.webhookBaseURL(projectID), c.apiVersion, projectID, webhookID)
	return c.doRequest(ctx, http.MethodDelete, url, nil, nil)
}

// ---------------------------------------------------------------------------
// Project Member Invitation Methods
// ---------------------------------------------------------------------------

// CreateInvitation sends an invitation to add a member to a project.
func (c *Client) CreateInvitation(ctx context.Context, projectID string, req CreateInvitationRequest) (*ProjectInvitation, error) {
	url := fmt.Sprintf("%s/%s/invitations/project/%s", c.baseURL, c.apiVersion, projectID)
	var resp CreateInvitationResponse
	if err := c.doRequest(ctx, http.MethodPost, url, req, &resp); err != nil {
		return nil, err
	}

	// Fetch full invitation details
	return c.GetInvitation(ctx, projectID, resp.InvitationID)
}

// GetInvitation fetches a specific project invitation by its ID.
func (c *Client) GetInvitation(ctx context.Context, projectID, invitationID string) (*ProjectInvitation, error) {
	invitations, err := c.ListInvitations(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, inv := range invitations {
		if inv.ID == invitationID {
			return &inv, nil
		}
	}
	return nil, &APIError{
		StatusCode: http.StatusNotFound,
		ErrorType:  "Not Found",
		Message:    fmt.Sprintf("Invitation %q not found in project %q", invitationID, projectID),
	}
}

// ListInvitations lists all pending invitations for a project.
func (c *Client) ListInvitations(ctx context.Context, projectID string) ([]ProjectInvitation, error) {
	url := fmt.Sprintf("%s/%s/invitations/project/%s", c.baseURL, c.apiVersion, projectID)
	var invitations []ProjectInvitation
	if err := c.doRequest(ctx, http.MethodGet, url, nil, &invitations); err != nil {
		return nil, err
	}
	return invitations, nil
}

// DeleteInvitation revokes a project member invitation.
func (c *Client) DeleteInvitation(ctx context.Context, projectID, invitationID string) error {
	url := fmt.Sprintf("%s/%s/invitations/project/%s/%s", c.baseURL, c.apiVersion, projectID, invitationID)
	return c.doRequest(ctx, http.MethodDelete, url, nil, nil)
}
