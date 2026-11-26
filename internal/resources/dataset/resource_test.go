package dataset

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestDatasetResource_Metadata(t *testing.T) {
	r := NewDatasetResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "sanity",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "sanity_dataset"
	if resp.TypeName != expected {
		t.Fatalf("expected resource type %q, got %q", expected, resp.TypeName)
	}
}

func TestDatasetResource_Schema(t *testing.T) {
	r := NewDatasetResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics error: %v", resp.Diagnostics.Errors())
	}

	requiredAttrs := []string{"id", "project_id", "name", "acl_mode", "created_at"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist in schema", attr)
		}
	}

	if !resp.Schema.Attributes["name"].IsRequired() {
		t.Error("expected name attribute to be required")
	}

	if !resp.Schema.Attributes["project_id"].IsRequired() {
		t.Error("expected project_id attribute to be required")
	}
}
