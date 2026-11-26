package cors_origin

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestCorsOriginResource_Metadata(t *testing.T) {
	r := NewCorsOriginResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "sanity",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "sanity_cors_origin"
	if resp.TypeName != expected {
		t.Fatalf("expected resource type %q, got %q", expected, resp.TypeName)
	}
}

func TestCorsOriginResource_Schema(t *testing.T) {
	r := NewCorsOriginResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics error: %v", resp.Diagnostics.Errors())
	}

	requiredAttrs := []string{"id", "project_id", "origin", "allow_credentials", "created_at", "updated_at"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist in schema", attr)
		}
	}

	if !resp.Schema.Attributes["origin"].IsRequired() {
		t.Error("expected origin attribute to be required")
	}

	if !resp.Schema.Attributes["project_id"].IsRequired() {
		t.Error("expected project_id attribute to be required")
	}

	if !resp.Schema.Attributes["id"].IsComputed() {
		t.Error("expected id attribute to be computed")
	}
}
