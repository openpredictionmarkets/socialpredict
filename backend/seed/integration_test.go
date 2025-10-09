package seed

import (
	"os"
	"strings"
	"testing"

	"socialpredict/handlers/cms/homepage"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestSeedHomepage_Integration_WithActualFile(t *testing.T) {
	// Create fake database for testing
	db := modelstesting.NewFakeDB(t)

	// Migrate the model
	err := db.AutoMigrate(&models.HomepageContent{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Get current working directory and construct path to actual home.md
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// The test is run from backend/seed, so repoRoot should be parent of parent
	repoRoot := cwd + "/../.."

	// Run the seeder with the actual home.md file
	err = SeedHomepage(db, repoRoot)
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

	// Verify HTML was rendered and contains rich content
	if content.HTML == "" {
		t.Fatal("HTML field should not be empty after seeding")
	}

	// Verify the markdown content contains the expected rich content
	if !strings.Contains(content.Markdown, "Enhanced Homepage Content") {
		t.Errorf("Expected markdown to contain title, got: %s", content.Markdown)
	}
	if !strings.Contains(content.Markdown, "BrierFoxForecast (BFF)") {
		t.Errorf("Expected markdown to contain BFF title, got: %s", content.Markdown)
	}
	if !strings.Contains(content.Markdown, "class=") {
		t.Errorf("Expected markdown to contain CSS classes, got: %s", content.Markdown)
	}

	// Verify HTML contains rendered content with preserved classes
	if !strings.Contains(content.HTML, "Enhanced Homepage Content") {
		t.Errorf("Expected HTML to contain title, got: %s", content.HTML)
	}
	if !strings.Contains(content.HTML, "BrierFoxForecast (BFF)") {
		t.Errorf("Expected HTML to contain BFF title, got: %s", content.HTML)
	}
	if !strings.Contains(content.HTML, `class=`) {
		t.Errorf("Expected HTML to contain CSS classes, got: %s", content.HTML)
	}
	if !strings.Contains(content.HTML, `/HomePageLogo.png`) {
		t.Errorf("Expected HTML to contain logo reference, got: %s", content.HTML)
	}

	// Test the service can retrieve the content properly
	repo := homepage.NewGormRepository(db)
	renderer := homepage.NewDefaultRenderer()
	svc := homepage.NewService(repo, renderer)

	retrievedContent, err := svc.GetHome()
	if err != nil {
		t.Fatalf("Failed to get home content via service: %v", err)
	}

	if retrievedContent.HTML != content.HTML {
		t.Errorf("Service returned different HTML than expected")
	}

	t.Logf("Successfully tested integration with rich content (HTML length: %d chars)", len(content.HTML))
}
