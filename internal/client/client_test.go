package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient_Validation(t *testing.T) {
	_, err := NewClient("")
	if err == nil {
		t.Fatal("expected error when initializing client with empty token, got nil")
	}

	c, err := NewClient("test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.token != "test-token" {
		t.Fatalf("expected token 'test-token', got %q", c.token)
	}
}

func TestClient_ProjectCRUD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2024-01-01/projects":
			var req CreateProjectRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(Project{
				ID:             "proj-123",
				DisplayName:    req.DisplayName,
				OrganizationID: req.OrganizationID,
			})

		case r.Method == http.MethodGet && r.URL.Path == "/v2024-01-01/projects/proj-123":
			_ = json.NewEncoder(w).Encode(Project{
				ID:          "proj-123",
				DisplayName: "Test Project",
			})

		case r.Method == http.MethodDelete && r.URL.Path == "/v2024-01-01/projects/proj-123":
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c, err := NewClient("test-token", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// 1. Create
	proj, err := c.CreateProject(ctx, CreateProjectRequest{
		DisplayName:    "Test Project",
		OrganizationID: "org-1",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if proj.ID != "proj-123" || proj.DisplayName != "Test Project" {
		t.Fatalf("unexpected project: %+v", proj)
	}

	// 2. Read
	gotProj, err := c.GetProject(ctx, "proj-123")
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if gotProj.DisplayName != "Test Project" {
		t.Fatalf("unexpected project display name: %s", gotProj.DisplayName)
	}

	// 3. Delete
	if err := c.DeleteProject(ctx, "proj-123"); err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
}

func TestClient_CorsCRUD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2024-01-01/projects/p1/cors":
			var req CreateCorsOriginRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(CorsOrigin{
				ID:               3294326,
				ProjectID:        "p1",
				Origin:           req.Origin,
				AllowCredentials: req.AllowCredentials,
			})

		case r.Method == http.MethodGet && r.URL.Path == "/v2024-01-01/projects/p1/cors":
			_ = json.NewEncoder(w).Encode([]CorsOrigin{
				{
					ID:               3294326,
					ProjectID:        "p1",
					Origin:           "http://localhost:3000",
					AllowCredentials: true,
				},
			})

		case r.Method == http.MethodDelete && r.URL.Path == "/v2024-01-01/projects/p1/cors/3294326":
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c, _ := NewClient("test-token", WithBaseURL(server.URL))
	ctx := context.Background()

	// Create
	created, err := c.CreateCorsOrigin(ctx, "p1", CreateCorsOriginRequest{
		Origin:           "http://localhost:3000",
		AllowCredentials: true,
	})
	if err != nil {
		t.Fatalf("CreateCorsOrigin failed: %v", err)
	}
	if created.ID != 3294326 {
		t.Fatalf("expected ID 3294326, got %d", created.ID)
	}

	// Get
	got, err := c.GetCorsOrigin(ctx, "p1", 3294326)
	if err != nil {
		t.Fatalf("GetCorsOrigin failed: %v", err)
	}
	if got.Origin != "http://localhost:3000" {
		t.Fatalf("expected origin 'http://localhost:3000', got %s", got.Origin)
	}

	// Delete
	if err := c.DeleteCorsOrigin(ctx, "p1", 3294326); err != nil {
		t.Fatalf("DeleteCorsOrigin failed: %v", err)
	}
}

func TestClient_WebhookCRUD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2024-01-01/hooks/projects/p1":
			var req CreateWebhookRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(Webhook{
				ID:         "hook-123",
				Name:       req.Name,
				URL:        req.URL,
				Dataset:    req.Dataset,
				APIVersion: req.APIVersion,
				Secret:     req.Secret,
			})

		case r.Method == http.MethodPatch && r.URL.Path == "/v2024-01-01/hooks/projects/p1/hook-123":
			var req UpdateWebhookRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			_ = json.NewEncoder(w).Encode(Webhook{
				ID:     "hook-123",
				Name:   *req.Name,
				URL:    *req.URL,
				Secret: "super-secret",
			})

		case r.Method == http.MethodDelete && r.URL.Path == "/v2024-01-01/hooks/projects/p1/hook-123":
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c, _ := NewClient("test-token", WithBaseURL(server.URL))
	ctx := context.Background()

	// 1. Create
	hook, err := c.CreateWebhook(ctx, "p1", CreateWebhookRequest{
		Name:       "test-hook",
		URL:        "https://example.com/hook",
		Dataset:    "production",
		APIVersion: "v2024-01-01",
		Secret:     "super-secret",
	})
	if err != nil {
		t.Fatalf("CreateWebhook failed: %v", err)
	}
	if hook.ID != "hook-123" || hook.Secret != "super-secret" {
		t.Fatalf("unexpected webhook response: %+v", hook)
	}

	// 2. Patch Update
	newName := "updated-hook"
	newURL := "https://example.com/new-hook"
	patched, err := c.UpdateWebhook(ctx, "p1", "hook-123", UpdateWebhookRequest{
		Name: &newName,
		URL:  &newURL,
	})
	if err != nil {
		t.Fatalf("UpdateWebhook failed: %v", err)
	}
	if patched.Name != "updated-hook" {
		t.Fatalf("expected updated name, got %s", patched.Name)
	}

	// 3. Delete
	if err := c.DeleteWebhook(ctx, "p1", "hook-123"); err != nil {
		t.Fatalf("DeleteWebhook failed: %v", err)
	}
}

func TestClient_IsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(APIError{
			StatusCode: 404,
			ErrorType:  "Not Found",
			Message:    "Resource not found",
		})
	}))
	defer server.Close()

	c, _ := NewClient("test-token", WithBaseURL(server.URL))
	_, err := c.GetProject(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !IsNotFound(err) {
		t.Fatalf("expected IsNotFound(err) to be true for 404, got false (err: %v)", err)
	}
}

func TestClient_InvitationCRUD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2024-01-01/invitations/project/p1":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(CreateInvitationResponse{
				InvitationID: "inv-123",
			})

		case r.Method == http.MethodGet && r.URL.Path == "/v2024-01-01/invitations/project/p1":
			_ = json.NewEncoder(w).Encode([]ProjectInvitation{
				{
					ID:        "inv-123",
					Email:     "dev@example.com",
					ProjectID: "p1",
					Roles:     []ProjectRole{{Name: "editor"}},
				},
			})

		case r.Method == http.MethodDelete && r.URL.Path == "/v2024-01-01/invitations/project/p1/inv-123":
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c, _ := NewClient("test-token", WithBaseURL(server.URL))
	ctx := context.Background()

	// 1. Create
	inv, err := c.CreateInvitation(ctx, "p1", CreateInvitationRequest{
		Email: "dev@example.com",
		Role:  "editor",
	})
	if err != nil {
		t.Fatalf("CreateInvitation failed: %v", err)
	}
	if inv.ID != "inv-123" || inv.Email != "dev@example.com" {
		t.Fatalf("unexpected invitation response: %+v", inv)
	}

	// 2. Read
	got, err := c.GetInvitation(ctx, "p1", "inv-123")
	if err != nil {
		t.Fatalf("GetInvitation failed: %v", err)
	}
	if got.Email != "dev@example.com" {
		t.Fatalf("unexpected invitation email: %s", got.Email)
	}

	// 3. Delete
	if err := c.DeleteInvitation(ctx, "p1", "inv-123"); err != nil {
		t.Fatalf("DeleteInvitation failed: %v", err)
	}
}
