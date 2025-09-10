package scraper

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/moseye/docinator/internal/models"
	"github.com/moseye/docinator/pkg/parser"
)

// ScrapingConfig holds configuration for the scraper
type ScrapingConfig struct {
	MaxConcurrency int           // Maximum concurrent requests
	Delay          time.Duration // Delay between requests
	Timeout        time.Duration // Request timeout
	UserAgent      string        // User agent string
	Debug          bool          // Enable debug logging
	TestMode       bool          // Enable test mode for mock data
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *ScrapingConfig {
	return &ScrapingConfig{
		MaxConcurrency: 2,                // Respectful concurrency
		Delay:          2 * time.Second,  // 2 second delay between requests
		Timeout:        30 * time.Second, // 30 second timeout
		UserAgent:      "docinator-scraper/1.0 (+https://github.com/moseye/docinator)",
		Debug:          false,
		TestMode:       false,
	}
}

// Scraper handles web scraping operations using Colly
type Scraper struct {
	config    *ScrapingConfig
	collector *colly.Collector
	parser    *parser.Parser
	mu        sync.RWMutex
	stats     ScrapingStats
}

// ScrapingStats tracks scraping statistics
type ScrapingStats struct {
	PackagesScraped int
	RequestsMade    int
	Errors          int
	StartTime       time.Time
}

// New creates a new Scraper instance with the given configuration
func New(config *ScrapingConfig) (*Scraper, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create collector with proper configuration for v2
	c := colly.NewCollector(
		colly.UserAgent(config.UserAgent),
		colly.AllowedDomains("pkg.go.dev", "go-colly.org"),
	)

	// Set up rate limiting
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: config.MaxConcurrency,
		Delay:       config.Delay,
	})

	// Set timeout
	c.SetRequestTimeout(config.Timeout)

	// Enable debug if requested
	if config.Debug {
		c.OnRequest(func(r *colly.Request) {
			log.Printf("[DEBUG] Visiting: %s", r.URL.String())
		})
	}

	// Create parser instance
	p := parser.New()

	scraper := &Scraper{
		config:    config,
		collector: c,
		parser:    p,
		stats: ScrapingStats{
			StartTime: time.Now(),
		},
	}

	// Set up event handlers
	scraper.setupEventHandlers()

	return scraper, nil
}

// setupEventHandlers configures the collector's event handlers
func (s *Scraper) setupEventHandlers() {
	// Track requests
	s.collector.OnRequest(func(r *colly.Request) {
		s.mu.Lock()
		s.stats.RequestsMade++
		s.mu.Unlock()

		if s.config.Debug {
			log.Printf("Visiting: %s", r.URL.String())
		}
	})

	// Track errors
	s.collector.OnError(func(r *colly.Response, err error) {
		s.mu.Lock()
		s.stats.Errors++
		s.mu.Unlock()

		log.Printf("Request error for %s: %v", r.Request.URL, err)
	})

	// Log successful responses
	s.collector.OnResponse(func(r *colly.Response) {
		if s.config.Debug {
			log.Printf("Response received from %s: %d", r.Request.URL, r.StatusCode)
		}
	})
}

// ScrapePackageWithRaw scrapes a Go package from pkg.go.dev and returns both structured data and raw HTML
func (s *Scraper) ScrapePackageWithRaw(ctx context.Context, importPath string) (*models.Package, string, error) {
	if strings.TrimSpace(importPath) == "" {
		return nil, "", fmt.Errorf("import path cannot be empty")
	}

	log.Printf("ScrapePackageWithRaw called for %s, TestMode: %v", importPath, s.config.TestMode)
	if s.config.TestMode {
		log.Printf("Returning mock package for %s", importPath)
		mockPkg := s.mockPackage(importPath)
		mockHTML := fmt.Sprintf(`<!DOCTYPE html><html><head><title>%s package - Go Packages</title></head><body><h1>%s</h1><p>%s</p><p>Mock HTML content for testing</p></body></html>`, mockPkg.Name, mockPkg.Name, mockPkg.Description)
		return mockPkg, mockHTML, nil
	}

	// Construct the URL for the package
	url := fmt.Sprintf("https://pkg.go.dev/%s", strings.TrimSpace(importPath))

	var pkg *models.Package
	var rawHTML string
	var scrapeErr error

	// Set up HTML parsing for the package page
	c := s.collector.Clone()

	c.OnHTML("html", func(e *colly.HTMLElement) {
		// Capture raw HTML content
		rawHTML, _ = e.DOM.Html()

		// Parse structured data
		var err error
		pkg, err = s.parser.ParsePackagePage(e)
		if err != nil {
			scrapeErr = fmt.Errorf("failed to parse package page: %w", err)
			return
		}

		// Set the import path from our parameter
		pkg.ImportPath = importPath
		pkg.ScrapedAt = time.Now()

		if s.config.Debug {
			log.Printf("Successfully parsed package: %s", pkg.ImportPath)
		}
	})

	// Visit the package URL
	if err := c.Visit(url); err != nil {
		return nil, "", fmt.Errorf("failed to visit %s: %w", url, err)
	}

	// Wait for the collector to finish
	c.Wait()

	if scrapeErr != nil {
		return nil, "", scrapeErr
	}

	if pkg == nil {
		return nil, "", fmt.Errorf("no package data found for %s", importPath)
	}

	// Update statistics
	s.mu.Lock()
	s.stats.PackagesScraped++
	s.mu.Unlock()

	return pkg, rawHTML, nil
}

// ScrapePackage scrapes a Go package from pkg.go.dev and returns structured data (backward compatibility)
func (s *Scraper) ScrapePackage(ctx context.Context, importPath string) (*models.Package, error) {
	pkg, _, err := s.ScrapePackageWithRaw(ctx, importPath)
	return pkg, err
}

// ScrapePackages scrapes multiple packages concurrently
func (s *Scraper) ScrapePackages(ctx context.Context, importPaths []string) ([]*models.Package, error) {
	if len(importPaths) == 0 {
		return nil, fmt.Errorf("no import paths provided")
	}

	if s.config.TestMode {
		// Sequential processing for tests to avoid concurrency issues
		var packages []*models.Package
		var errors []error
		for _, importPath := range importPaths {
			select {
			case <-ctx.Done():
				break
			default:
			}
			pkg, err := s.ScrapePackage(ctx, importPath)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to scrape %s: %w", importPath, err))
			} else {
				packages = append(packages, pkg)
			}
		}
		if len(errors) > 0 {
			for _, err := range errors {
				log.Printf("Scraping error: %v", err)
			}
			return packages, errors[0]
		}
		return packages, nil
	}

	packages := make([]*models.Package, 0, len(importPaths))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use a channel to limit concurrency
	semaphore := make(chan struct{}, s.config.MaxConcurrency)

	// Collect errors
	errors := make([]error, 0)
	var errMu sync.Mutex

	for _, importPath := range importPaths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-semaphore }()

			// Check context cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			pkg, err := s.ScrapePackage(ctx, path)
			if err != nil {
				errMu.Lock()
				errors = append(errors, fmt.Errorf("failed to scrape %s: %w", path, err))
				errMu.Unlock()
				return
			}

			mu.Lock()
			packages = append(packages, pkg)
			mu.Unlock()
		}(importPath)
	}

	wg.Wait()

	if len(errors) > 0 {
		// Return the first error, but log all errors
		for _, err := range errors {
			log.Printf("Scraping error: %v", err)
		}
		return packages, errors[0]
	}

	return packages, nil
}

// GetStats returns current scraping statistics
func (s *Scraper) GetStats() ScrapingStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}

// Close cleans up the scraper resources
func (s *Scraper) Close() error {
	// Colly doesn't require explicit cleanup, but we can clear internal state
	s.mu.Lock()
	s.stats = ScrapingStats{StartTime: time.Now()}
	s.mu.Unlock()

	return nil
}

// ValidateURL checks if a URL is valid for scraping
func ValidateURL(url string) error {
	if !strings.HasPrefix(url, "https://pkg.go.dev/") {
		return fmt.Errorf("URL must be from pkg.go.dev domain")
	}
	return nil
}

// mockPackage returns a mock package for testing
func (s *Scraper) mockPackage(importPath string) *models.Package {
	return &models.Package{
		Name:        "cobra",
		Description: "A Commander providing a simple interface to create powerful modern CLI interfaces",
		Module:      "github.com/spf13/cobra",
		Version:     "v1.9.1",
		Synopsis:    "Commander library",
		License:     "Apache-2.0",
		Repository:  "https://github.com/spf13/cobra",
		ImportPath:  importPath,
		ScrapedAt:   time.Now(),
		Functions: []models.Function{
			{
				Name:        "Execute",
				Description: "Executes the root command and all subcommands.",
				Signature:   "func Execute() error",
				Receiver:    "",
				Examples:    []models.Example{},
			},
		},
		Types:       []models.Type{},
		Variables:   []models.Variable{},
		Constants:   []models.Constant{},
		Examples:    []models.Example{},
	}
}

// ExtractImportPath extracts the import path from a pkg.go.dev URL
func ExtractImportPath(url string) (string, error) {
	if err := ValidateURL(url); err != nil {
		return "", err
	}

	// Remove the base URL to get the import path
	importPath := strings.TrimPrefix(url, "https://pkg.go.dev/")
	importPath = strings.TrimSuffix(importPath, "/")

	if importPath == "" {
		return "", fmt.Errorf("no import path found in URL")
	}

	return importPath, nil
}
