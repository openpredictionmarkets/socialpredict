package mcpserver

import "context"

type AccessLevel string

const (
	AccessPublicRead  AccessLevel = "public_read"
	AccessUserRead    AccessLevel = "user_read"
	AccessUserWrite   AccessLevel = "user_write"
	AccessServiceRead AccessLevel = "service_read"
)

type Principal struct {
	Kind     string `json:"kind"`
	Username string `json:"username,omitempty"`
	Service  string `json:"service,omitempty"`
}

type PrincipalResolver interface {
	Resolve(ctx context.Context, required AccessLevel) (Principal, error)
}

type AnonymousResolver struct{}

func (AnonymousResolver) Resolve(_ context.Context, required AccessLevel) (Principal, error) {
	if required != AccessPublicRead {
		return Principal{}, &ToolError{Code: "unauthorized", Message: "authentication required"}
	}
	return Principal{Kind: "anonymous"}, nil
}
