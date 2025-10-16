package seed

import _ "embed"

// Embed the homepage markdown content at compile time
// This ensures the content is available regardless of filesystem layout
//
//go:embed home.md
var defaultHomeMD []byte
