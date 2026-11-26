package client

import "time"

// Project represents a Sanity project.
type Project struct {
	ID                  string                 `json:"id"`
	DisplayName         string                 `json:"displayName"`
	StudioHost          *string                `json:"studioHost"`
	OrganizationID      string                 `json:"organizationId"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	ActivityFeedEnabled bool                   `json:"activityFeedEnabled"`
	CreatedAt           time.Time              `json:"createdAt"`
	UpdatedAt           time.Time              `json:"updatedAt"`
}

// CreateProjectRequest is the payload for creating a project.
type CreateProjectRequest struct {
	DisplayName    string `json:"displayName"`
	OrganizationID string `json:"organizationId,omitempty"`
}

// UpdateProjectRequest is the payload for updating a project.
type UpdateProjectRequest struct {
	DisplayName *string                `json:"displayName,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Dataset represents a Sanity dataset.
type Dataset struct {
	Name           string    `json:"name"`
	ACLMode        string    `json:"aclMode"` // "public" or "private"
	CreatedAt      time.Time `json:"createdAt"`
	CreatedBy      string    `json:"createdByUserId,omitempty"`
	DatasetProfile string    `json:"datasetProfile,omitempty"`
}

// CreateDatasetRequest is the payload for creating a dataset via PUT.
type CreateDatasetRequest struct {
	ACLMode string `json:"aclMode"` // "public" or "private"
}

// CorsOrigin represents a CORS origin configuration.
type CorsOrigin struct {
	ID               int64     `json:"id"`
	ProjectID        string    `json:"projectId"`
	Origin           string    `json:"origin"`
	AllowCredentials bool      `json:"allowCredentials"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// CreateCorsOriginRequest is the payload for creating a CORS origin.
type CreateCorsOriginRequest struct {
	Origin           string `json:"origin"`
	AllowCredentials bool   `json:"allowCredentials"`
}

// Webhook represents a Sanity document webhook.
type Webhook struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Type               string            `json:"type"` // typically "document"
	APIVersion         string            `json:"apiVersion"`
	Dataset            string            `json:"dataset"`
	URL                string            `json:"url"`
	HTTPMethod         string            `json:"httpMethod"`
	Secret             string            `json:"secret,omitempty"`
	Filter             *string           `json:"filter,omitempty"`
	Projection         *string           `json:"projection,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	IncludeDrafts      bool              `json:"includeDrafts"`
	IncludeAllVersions bool              `json:"includeAllVersions"`
	IsDisabledByUser   bool              `json:"isDisabledByUser"`
	CreatedAt          time.Time         `json:"createdAt"`
}

// CreateWebhookRequest is the payload for creating a document webhook.
type CreateWebhookRequest struct {
	Type               string            `json:"type"`
	Name               string            `json:"name"`
	URL                string            `json:"url"`
	Dataset            string            `json:"dataset"`
	HTTPMethod         string            `json:"httpMethod"`
	APIVersion         string            `json:"apiVersion"`
	Secret             string            `json:"secret,omitempty"`
	Filter             *string           `json:"filter,omitempty"`
	Projection         *string           `json:"projection,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	IncludeDrafts      bool              `json:"includeDrafts"`
	IncludeAllVersions bool              `json:"includeAllVersions"`
	IsDisabledByUser   bool              `json:"isDisabledByUser"`
}

// UpdateWebhookRequest is the payload for updating a webhook via PATCH.
type UpdateWebhookRequest struct {
	Name               *string           `json:"name,omitempty"`
	URL                *string           `json:"url,omitempty"`
	Dataset            *string           `json:"dataset,omitempty"`
	HTTPMethod         *string           `json:"httpMethod,omitempty"`
	APIVersion         *string           `json:"apiVersion,omitempty"`
	Secret             *string           `json:"secret,omitempty"`
	Filter             *string           `json:"filter,omitempty"`
	Projection         *string           `json:"projection,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	IncludeDrafts      *bool             `json:"includeDrafts,omitempty"`
	IncludeAllVersions *bool             `json:"includeAllVersions,omitempty"`
	IsDisabledByUser   *bool             `json:"isDisabledByUser,omitempty"`
}

// ProjectInvitation represents a project member invitation.
type ProjectInvitation struct {
	ID         string        `json:"id"`
	Email      string        `json:"email"`
	ProjectID  string        `json:"projectId"`
	Roles      []ProjectRole `json:"roles,omitempty"`
	IsAccepted bool          `json:"isAccepted"`
	IsRevoked  bool          `json:"isRevoked"`
	CreatedAt  time.Time     `json:"createdAt"`
}

// ProjectRole represents a role attached to a member or invitation.
type ProjectRole struct {
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`
}

// CreateInvitationRequest is the payload for inviting a member.
type CreateInvitationRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// CreateInvitationResponse holds the returned invitation ID.
type CreateInvitationResponse struct {
	InvitationID string `json:"invitationId"`
}
