package webhook

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestWebhookResource_Metadata(t *testing.T) {
	r := NewWebhookResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "sanity",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "sanity_webhook"
	if resp.TypeName != expected {
		t.Fatalf("expected resource type %q, got %q", expected, resp.TypeName)
	}
}

func TestWebhookResource_Schema(t *testing.T) {
	r := NewWebhookResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics error: %v", resp.Diagnostics.Errors())
	}

	requiredAttrs := []string{"id", "project_id", "name", "url", "dataset", "api_version", "http_method", "secret", "filter", "projection", "headers", "include_drafts", "include_all_versions", "is_disabled_by_user", "created_at"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist in schema", attr)
		}
	}

	if !resp.Schema.Attributes["name"].IsRequired() {
		t.Error("expected name attribute to be required")
	}

	if !resp.Schema.Attributes["api_version"].IsRequired() {
		t.Error("expected api_version attribute to be required")
	}

	if !resp.Schema.Attributes["secret"].IsSensitive() {
		t.Error("expected secret attribute to be marked as sensitive")
	}
}
