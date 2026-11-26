package invitation

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestInvitationResource_Metadata(t *testing.T) {
	r := NewInvitationResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "sanity",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "sanity_project_invitation"
	if resp.TypeName != expected {
		t.Fatalf("expected resource type %q, got %q", expected, resp.TypeName)
	}
}

func TestInvitationResource_Schema(t *testing.T) {
	r := NewInvitationResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics error: %v", resp.Diagnostics.Errors())
	}

	requiredAttrs := []string{"id", "project_id", "email", "role", "is_accepted", "is_revoked", "created_at"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist in schema", attr)
		}
	}

	if !resp.Schema.Attributes["email"].IsRequired() {
		t.Error("expected email attribute to be required")
	}

	if !resp.Schema.Attributes["role"].IsRequired() {
		t.Error("expected role attribute to be required")
	}

	if !resp.Schema.Attributes["project_id"].IsRequired() {
		t.Error("expected project_id attribute to be required")
	}
}
