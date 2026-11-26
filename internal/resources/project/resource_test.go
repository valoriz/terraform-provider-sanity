package project

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProjectResource_Metadata(t *testing.T) {
	r := NewProjectResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "sanity",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "sanity_project"
	if resp.TypeName != expected {
		t.Fatalf("expected resource type %q, got %q", expected, resp.TypeName)
	}
}

func TestProjectResource_Schema(t *testing.T) {
	r := NewProjectResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics error: %v", resp.Diagnostics.Errors())
	}

	requiredAttrs := []string{"id", "project_name", "organization_id", "studio_host", "created_at", "updated_at"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist in schema", attr)
		}
	}

	if !resp.Schema.Attributes["project_name"].IsRequired() {
		t.Error("expected project_name attribute to be required")
	}

	if !resp.Schema.Attributes["id"].IsComputed() {
		t.Error("expected id attribute to be computed")
	}
}
