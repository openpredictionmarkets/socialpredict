package homepage

import (
	"bytes"
	"errors"
	"socialpredict/models"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
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
	return &DefaultRenderer{
		md: goldmark.New(),
		pm: bluemonday.UGCPolicy(),
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
