package scraper

import (
	"context"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
)

func TestScrapePackage_Cobra(t *testing.T) {
	config := DefaultConfig()
	config.Debug = true
	config.Timeout = 60 * time.Second // Increase timeout for testing
	s, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create scraper: %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pkg, err := s.ScrapePackage(ctx, "github.com/spf13/cobra")
	if err != nil {
		t.Fatalf("Failed to scrape cobra package: %v", err)
	}

	// Validate meaningful content
	if pkg.Name == "" {
		t.Error("Package name should not be empty")
	}
	if pkg.Description == "" {
		t.Error("Package description should not be empty")
	}
	if len(pkg.Functions) == 0 {
		t.Error("Should have at least one function")
	}
	if len(pkg.Types) == 0 {
		t.Error("Should have at least one type")
	}
	t.Logf("Successfully scraped cobra package with %d functions and %d types", len(pkg.Functions), len(pkg.Types))
}

func TestScrapePackage_GoQuery(t *testing.T) {
	config := DefaultConfig()
	config.Debug = true
	config.Timeout = 60 * time.Second
	s, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create scraper: %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pkg, err := s.ScrapePackage(ctx, "github.com/PuerkitoBio/goquery")
	if err != nil {
		t.Fatalf("Failed to scrape goquery package: %v", err)
	}

	// Validate meaningful content
	if pkg.Name == "" {
		t.Error("Package name should not be empty")
	}
	if pkg.Description == "" {
		t.Error("Package description should not be empty")
	}
	if len(pkg.Functions) == 0 {
		t.Error("Should have at least one function")
	}
	if len(pkg.Types) == 0 {
		t.Error("Should have at least one type")
	}
	t.Logf("Successfully scraped goquery package with %d functions and %d types", len(pkg.Functions), len(pkg.Types))
}

func TestScrapeCollyDocs(t *testing.T) {
	// For colly.org, create a custom scraper since it's not pkg.go.dev
	config := DefaultConfig()
	config.Debug = true
	config.Timeout = 60 * time.Second
	s, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create scraper: %v", err)
	}
	defer s.Close()

	_, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Basic scraping for colly docs - fetch and check for content
	url := "https://go-colly.org/docs"
	c := s.collector.Clone()
	c.AllowedDomains = append(c.AllowedDomains, "go-colly.org")

	var content string
	c.OnHTML("body", func(e *colly.HTMLElement) {
		content = e.Text // Extract text content
	})

	if err := c.Visit(url); err != nil {
		t.Fatalf("Failed to scrape colly docs: %v", err)
	}
	c.Wait()

	if content == "" {
		t.Error("Colly docs should contain meaningful content")
	}
	t.Log("Successfully scraped colly documentation with content length:", len(content))
}

func TestScrapePackages_Multiple(t *testing.T) {
	config := DefaultConfig()
	config.Debug = true
	config.Timeout = 60 * time.Second
	s, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create scraper: %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	importPaths := []string{
		"github.com/spf13/cobra",
		"github.com/PuerkitoBio/goquery",
	}
	pkgs, err := s.ScrapePackages(ctx, importPaths)
	if err != nil {
		t.Fatalf("Failed to scrape multiple packages: %v", err)
	}

	if len(pkgs) != len(importPaths) {
		t.Errorf("Expected %d packages, got %d", len(importPaths), len(pkgs))
	}

	for _, pkg := range pkgs {
		if pkg.Name == "" {
			t.Error("Package name should not be empty")
		}
		if len(pkg.Functions)+len(pkg.Types) == 0 {
			t.Error("Package should have functions or types")
		}
	}
	t.Logf("Successfully scraped %d packages", len(pkgs))
}
