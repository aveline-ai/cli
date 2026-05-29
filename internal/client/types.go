package client

// UserRef is the embedded representation of a user inside an Item.
type UserRef struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
}

// User is the full /api/me user record.
type User struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
}

// Workspace is the JSON shape of a workspace.
type Workspace struct {
	ID         string  `json:"id"`
	Slug       string  `json:"slug"`
	Name       string  `json:"name"`
	InsertedAt string  `json:"inserted_at"`
	UpdatedAt  string  `json:"updated_at"`
	DeletedAt  *string `json:"deleted_at"`
}

// Item is the JSON shape of an item.
type Item struct {
	ID         string   `json:"id"`
	Slug       string   `json:"slug"`
	Title      string   `json:"title"`
	Body       string   `json:"body"`
	Summary    *string  `json:"summary"`
	Tags       []string `json:"tags"`
	Pinned     bool     `json:"pinned"`
	Owner      UserRef  `json:"owner"`
	CreatedBy  UserRef  `json:"created_by"`
	CreatedVia string   `json:"created_via"`
	InsertedAt string   `json:"inserted_at"`
	UpdatedAt  string   `json:"updated_at"`
	DeletedAt  *string  `json:"deleted_at"`
}

// View is the JSON shape of a saved view.
type View struct {
	ID          string   `json:"id"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	TagFilter   []string `json:"tag_filter"`
	Description *string  `json:"description"`
	InsertedAt  string   `json:"inserted_at"`
	UpdatedAt   string   `json:"updated_at"`
	DeletedAt   *string  `json:"deleted_at"`
}

// MeResponse is the body of GET /api/me.
type MeResponse struct {
	User       User        `json:"user"`
	Workspaces []Workspace `json:"workspaces"`
}

// ViewWithItems is the body of GET /api/workspaces/:slug/views/:view_slug.
type ViewWithItems struct {
	View  View   `json:"view"`
	Items []Item `json:"items"`
}
