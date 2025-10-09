package homepage

import (
	"bytes"
	"errors"
	"socialpredict/models"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

type Renderer interface {
	MarkdownToHTML(md string) (string, error)
	SanitizeHTML(html string) string
}

type DefaultRenderer struct {
	md goldmark.Markdown
	pm *bluemonday.Policy
}

func NewDefaultRenderer() *DefaultRenderer {
	// Create a permissive policy for rich HTML content with Tailwind classes
	policy := bluemonday.NewPolicy()

	// Allow all common HTML structural elements
	policy.AllowElements(
		"div", "section", "p", "span",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li",
		"a", "img",
		"strong", "em", "b", "i",
		"br", "hr",
	)

	// Allow class attribute on all elements (for Tailwind CSS)
	policy.AllowAttrs("class", "id").Globally()

	// Allow specific attributes for links
	policy.AllowAttrs("href", "target", "rel", "title").OnElements("a")

	// Allow specific attributes for images
	policy.AllowAttrs("src", "alt", "width", "height", "title").OnElements("img")

	// Allow URL schemes for links and images
	policy.AllowURLSchemes("http", "https", "mailto")

	return &DefaultRenderer{
		md: goldmark.New(
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
			),
		),
		pm: policy,
	}
}

func (r *DefaultRenderer) MarkdownToHTML(md string) (string, error) {
	var w = new(bytes.Buffer)
	if err := r.md.Convert([]byte(md), w); err != nil {
		return "", err
	}
	return w.String(), nil
}

func (r *DefaultRenderer) SanitizeHTML(html string) string {
	return r.pm.Sanitize(html)
}

// Use higher-order injection for renderer & repository for testability.
type Service struct {
	repo     Repository
	renderer Renderer
}

func NewService(repo Repository, renderer Renderer) *Service {
	return &Service{repo: repo, renderer: renderer}
}

type UpdateInput struct {
	Title     string
	Format    string // "markdown" or "html"
	Markdown  string
	HTML      string
	UpdatedBy string
	Version   uint
}

func (s *Service) GetHome() (*models.HomepageContent, error) {
	return s.repo.GetBySlug("home")
}

func (s *Service) UpdateHome(in UpdateInput) (*models.HomepageContent, error) {
	item, err := s.repo.GetBySlug("home")
	if err != nil {
		return nil, err
	}
	// optimistic lock check (optional; keep consistent with your style)
	if in.Version != 0 && in.Version != item.Version {
		return nil, errors.New("version mismatch")
	}

	item.Title = in.Title
	item.Format = in.Format
	item.UpdatedBy = in.UpdatedBy
	item.Version = item.Version + 1

	switch in.Format {
	case "markdown":
		item.Markdown = in.Markdown
		rendered, err := s.renderer.MarkdownToHTML(in.Markdown)
		if err != nil {
			return nil, err
		}
		item.HTML = s.renderer.SanitizeHTML(rendered)
	case "html":
		item.HTML = s.renderer.SanitizeHTML(in.HTML)
		item.Markdown = "" // optional: clear or keep last md
	default:
		return nil, errors.New("unsupported format")
	}

	if err := s.repo.Save(item); err != nil {
		return nil, err
	}
	return item, nil
}
