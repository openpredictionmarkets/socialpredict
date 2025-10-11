package homepage

import (
	"errors"
	"strings"
	"testing"

	"socialpredict/models"
)

// Mock repository for testing
type mockRepository struct {
	items   map[string]*models.HomepageContent
	saveErr error
	getErr  error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		items: make(map[string]*models.HomepageContent),
	}
}

func (m *mockRepository) GetBySlug(slug string) (*models.HomepageContent, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	item, exists := m.items[slug]
	if !exists {
		return nil, errors.New("not found")
	}
	return item, nil
}

func (m *mockRepository) Save(item *models.HomepageContent) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.items[item.Slug] = item
	return nil
}

// Mock renderer for testing
type mockRenderer struct {
	markdownToHTMLErr error
	htmlOutput        string
	sanitizeOutput    string
}

func newMockRenderer() *mockRenderer {
	return &mockRenderer{
		htmlOutput:     "<p>test</p>",
		sanitizeOutput: "<p>sanitized</p>",
	}
}

func (m *mockRenderer) MarkdownToHTML(md string) (string, error) {
	if m.markdownToHTMLErr != nil {
		return "", m.markdownToHTMLErr
	}
	if md == "# Test" {
		return "<h1>Test</h1>", nil
	}
	return m.htmlOutput, nil
}

func (m *mockRenderer) SanitizeHTML(html string) string {
	if strings.Contains(html, "<script>") {
		return strings.ReplaceAll(html, "<script>", "")
	}
	// For testing, just return the input HTML if no script tags
	return html
}

func TestService_GetHome(t *testing.T) {
	repo := newMockRepository()
	renderer := newMockRenderer()
	svc := NewService(repo, renderer)

	// Test when item exists
	expected := &models.HomepageContent{
		Slug:     "home",
		Title:    "Test Home",
		Format:   "markdown",
		Markdown: "# Test",
		HTML:     "<h1>Test</h1>",
		Version:  1,
	}
	repo.items["home"] = expected

	result, err := svc.GetHome()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.Title != expected.Title {
		t.Errorf("Expected title %s, got %s", expected.Title, result.Title)
	}

	// Test when item doesn't exist
	delete(repo.items, "home")
	_, err = svc.GetHome()
	if err == nil {
		t.Fatal("Expected error when item doesn't exist")
	}
}

func TestService_UpdateHome_Markdown(t *testing.T) {
	repo := newMockRepository()
	renderer := newMockRenderer()
	svc := NewService(repo, renderer)

	// Seed initial item
	initialItem := &models.HomepageContent{
		Slug:     "home",
		Title:    "Old Title",
		Format:   "markdown",
		Markdown: "# Old Content",
		HTML:     "<h1>Old Content</h1>",
		Version:  1,
	}
	repo.items["home"] = initialItem

	// Update with markdown
	input := UpdateInput{
		Title:     "New Title",
		Format:    "markdown",
		Markdown:  "# Test",
		UpdatedBy: "admin",
		Version:   1,
	}

	result, err := svc.UpdateHome(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Title != "New Title" {
		t.Errorf("Expected title 'New Title', got '%s'", result.Title)
	}
	if result.Format != "markdown" {
		t.Errorf("Expected format 'markdown', got '%s'", result.Format)
	}
	if result.Markdown != "# Test" {
		t.Errorf("Expected markdown '# Test', got '%s'", result.Markdown)
	}
	if result.HTML != "<h1>Test</h1>" {
		t.Errorf("Expected HTML '<h1>Test</h1>', got '%s'", result.HTML)
	}
	if result.Version != 2 {
		t.Errorf("Expected version 2, got %d", result.Version)
	}
	if result.UpdatedBy != "admin" {
		t.Errorf("Expected UpdatedBy 'admin', got '%s'", result.UpdatedBy)
	}
}

func TestService_UpdateHome_HTML(t *testing.T) {
	repo := newMockRepository()
	renderer := newMockRenderer()
	svc := NewService(repo, renderer)

	// Seed initial item
	initialItem := &models.HomepageContent{
		Slug:     "home",
		Title:    "Old Title",
		Format:   "markdown",
		Markdown: "# Old Content",
		HTML:     "<h1>Old Content</h1>",
		Version:  1,
	}
	repo.items["home"] = initialItem

	// Update with HTML
	input := UpdateInput{
		Title:     "New Title",
		Format:    "html",
		HTML:      "<p>Direct HTML</p>",
		UpdatedBy: "admin",
		Version:   1,
	}

	result, err := svc.UpdateHome(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Format != "html" {
		t.Errorf("Expected format 'html', got '%s'", result.Format)
	}
	if result.HTML != "<p>Direct HTML</p>" {
		t.Errorf("Expected sanitized HTML, got '%s'", result.HTML)
	}
	if result.Markdown != "" {
		t.Errorf("Expected markdown to be cleared, got '%s'", result.Markdown)
	}
}

func TestService_UpdateHome_VersionMismatch(t *testing.T) {
	repo := newMockRepository()
	renderer := newMockRenderer()
	svc := NewService(repo, renderer)

	// Seed initial item
	initialItem := &models.HomepageContent{
		Slug:    "home",
		Title:   "Test",
		Version: 5,
	}
	repo.items["home"] = initialItem

	// Try to update with wrong version
	input := UpdateInput{
		Title:   "New Title",
		Format:  "markdown",
		Version: 3, // Wrong version
	}

	_, err := svc.UpdateHome(input)
	if err == nil {
		t.Fatal("Expected version mismatch error")
	}
	if !strings.Contains(err.Error(), "version mismatch") {
		t.Errorf("Expected version mismatch error, got '%s'", err.Error())
	}
}

func TestService_UpdateHome_UnsupportedFormat(t *testing.T) {
	repo := newMockRepository()
	renderer := newMockRenderer()
	svc := NewService(repo, renderer)

	// Seed initial item
	initialItem := &models.HomepageContent{
		Slug:    "home",
		Version: 1,
	}
	repo.items["home"] = initialItem

	// Try to update with unsupported format
	input := UpdateInput{
		Format:  "pdf", // Unsupported
		Version: 1,
	}

	_, err := svc.UpdateHome(input)
	if err == nil {
		t.Fatal("Expected unsupported format error")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected unsupported format error, got '%s'", err.Error())
	}
}

func TestService_UpdateHome_RepositoryError(t *testing.T) {
	repo := newMockRepository()
	renderer := newMockRenderer()
	svc := NewService(repo, renderer)

	// Set up repository to return error on save
	repo.saveErr = errors.New("database error")

	// Seed initial item
	initialItem := &models.HomepageContent{
		Slug:    "home",
		Version: 1,
	}
	repo.items["home"] = initialItem

	input := UpdateInput{
		Format:  "markdown",
		Version: 1,
	}

	_, err := svc.UpdateHome(input)
	if err == nil {
		t.Fatal("Expected repository error")
	}
	if !strings.Contains(err.Error(), "database error") {
		t.Errorf("Expected database error, got '%s'", err.Error())
	}
}

func TestSanitization(t *testing.T) {
	repo := newMockRepository()
	renderer := newMockRenderer()
	svc := NewService(repo, renderer)

	// Seed initial item
	initialItem := &models.HomepageContent{
		Slug:    "home",
		Version: 1,
	}
	repo.items["home"] = initialItem

	// Test that script tags are sanitized
	input := UpdateInput{
		Format:  "html",
		HTML:    "<p>Safe content</p><script>alert('xss')</script>",
		Version: 1,
	}

	result, err := svc.UpdateHome(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if strings.Contains(result.HTML, "<script>") {
		t.Error("Expected script tags to be sanitized")
	}
}
