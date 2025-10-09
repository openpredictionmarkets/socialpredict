package seed

import (
	"os"
	"strings"
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestSeedHomepage_RendersHTML(t *testing.T) {
	// Create fake database for testing
	db := modelstesting.NewFakeDB(t)

	// Migrate the model
	err := db.AutoMigrate(&models.HomepageContent{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	seedDir := tempDir + "/backend/seed"
	err = os.MkdirAll(seedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test markdown file with rich content (similar to actual home.md)
	testMarkdown := `# Enhanced Homepage Content

<div class="flex flex-col sm:flex-row items-center mb-8">
  <img src="/HomePageLogo.png" alt="BrierFoxForecast Logo" class="w-24 h-24 sm:w-32 sm:h-32 mb-4 sm:mb-0 sm:mr-6" />
  <div class="flex flex-col justify-center h-full text-center sm:text-left">
    <h1 class="text-3xl sm:text-4xl font-bold text-custom-gray-light mb-2">BrierFoxForecast (BFF)</h1>
    <p class="text-lg text-custom-gray-light">An alpha project powered by SocialPredict's open-source prediction market platform.</p>
  </div>
</div>

<div class="space-y-8">
  <section class="bg-gray-800 rounded-lg p-6 shadow-lg">
    <h2 class="text-xl font-semibold mb-3 text-custom-gray-light">About BFF</h2>
    <p class="text-base mb-4">BFF is a platform for predictions on politics, finance, business, world news, and more.</p>
  </section>
</div>`

	err = os.WriteFile(seedDir+"/home.md", []byte(testMarkdown), 0644)
	if err != nil {
		t.Fatalf("Failed to create test markdown file: %v", err)
	}

	// Run the seeder
	err = SeedHomepage(db, tempDir)
	if err != nil {
		t.Fatalf("SeedHomepage failed: %v", err)
	}

	// Verify the content was seeded correctly
	var content models.HomepageContent
	err = db.Where("slug = ?", "home").First(&content).Error
	if err != nil {
		t.Fatalf("Failed to find seeded content: %v", err)
	}

	// Verify basic fields
	if content.Title != "Home" {
		t.Errorf("Expected title 'Home', got '%s'", content.Title)
	}
	if content.Format != "markdown" {
		t.Errorf("Expected format 'markdown', got '%s'", content.Format)
	}
	if content.Markdown != testMarkdown {
		t.Errorf("Expected markdown content to match, got: %s", content.Markdown)
	}

	// Most importantly, verify HTML was rendered
	if content.HTML == "" {
		t.Fatal("HTML field should not be empty after seeding")
	}

	// Verify HTML contains expected rendered elements from rich content
	if !strings.Contains(content.HTML, "<h1>Enhanced Homepage Content</h1>") {
		t.Errorf("Expected HTML to contain main title, got: %s", content.HTML)
	}
	if !strings.Contains(content.HTML, "BrierFoxForecast (BFF)") {
		t.Errorf("Expected HTML to contain BFF title, got: %s", content.HTML)
	}
	if !strings.Contains(content.HTML, "About BFF") {
		t.Errorf("Expected HTML to contain About section, got: %s", content.HTML)
	}
	if !strings.Contains(content.HTML, `class="bg-gray-800`) {
		t.Errorf("Expected HTML to contain Tailwind classes, got: %s", content.HTML)
	}
	if !strings.Contains(content.HTML, `/HomePageLogo.png`) {
		t.Errorf("Expected HTML to contain logo reference, got: %s", content.HTML)
	}

	t.Logf("Successfully seeded rich content with HTML length: %d chars", len(content.HTML))
}

func TestSeedHomepage_FallbackContent(t *testing.T) {
	// Create fake database for testing
	db := modelstesting.NewFakeDB(t)

	// Migrate the model
	err := db.AutoMigrate(&models.HomepageContent{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Use non-existent directory so it falls back to default content
	tempDir := "/nonexistent"

	// Run the seeder
	err = SeedHomepage(db, tempDir)
	if err != nil {
		t.Fatalf("SeedHomepage failed: %v", err)
	}

	// Verify the fallback content was seeded
	var content models.HomepageContent
	err = db.Where("slug = ?", "home").First(&content).Error
	if err != nil {
		t.Fatalf("Failed to find seeded content: %v", err)
	}

	expectedMarkdown := "# Welcome to BrierFoxForecast\n\nThis is the seeded home page."
	if content.Markdown != expectedMarkdown {
		t.Errorf("Expected fallback markdown, got: %s", content.Markdown)
	}

	// Verify HTML was still rendered for fallback content
	if content.HTML == "" {
		t.Fatal("HTML field should not be empty even for fallback content")
	}
	if !strings.Contains(content.HTML, "<h1>Welcome to BrierFoxForecast</h1>") {
		t.Errorf("Expected rendered fallback HTML, got: %s", content.HTML)
	}
}

func TestSeedHomepage_DoesNotDuplicateExisting(t *testing.T) {
	// Create fake database for testing
	db := modelstesting.NewFakeDB(t)

	// Migrate the model
	err := db.AutoMigrate(&models.HomepageContent{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Pre-seed some content
	existing := models.HomepageContent{
		Slug:     "home",
		Title:    "Existing",
		Format:   "markdown",
		Markdown: "existing content",
		HTML:     "<p>existing content</p>",
		Version:  1,
	}
	db.Create(&existing)

	// Run the seeder - should not create duplicate
	err = SeedHomepage(db, "/tmp")
	if err != nil {
		t.Fatalf("SeedHomepage failed: %v", err)
	}

	// Verify only one record exists and it wasn't changed
	var count int64
	db.Model(&models.HomepageContent{}).Where("slug = ?", "home").Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}

	var content models.HomepageContent
	db.Where("slug = ?", "home").First(&content)
	if content.Title != "Existing" {
		t.Errorf("Existing content was modified: %s", content.Title)
	}
}
