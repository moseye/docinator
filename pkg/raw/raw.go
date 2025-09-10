package raw

import (
	"fmt"
	"strings"

	"github.com/moseye/docinator/internal/models"
)

// PackageToRaw converts a Package struct to raw text format containing the original scraped content.
func PackageToRaw(pkg *models.Package, rawHTML string) string {
	var b strings.Builder

	// Header with package info
	b.WriteString(fmt.Sprintf("=== RAW WEB SCRAPE DATA ===\n"))
	b.WriteString(fmt.Sprintf("Package: %s\n", pkg.Name))
	b.WriteString(fmt.Sprintf("Import Path: %s\n", pkg.ImportPath))
	b.WriteString(fmt.Sprintf("Scraped At: %s\n", pkg.ScrapedAt.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("Source URL: https://pkg.go.dev/%s\n", pkg.ImportPath))
	b.WriteString("================================\n\n")

	// Raw HTML content
	if rawHTML != "" {
		b.WriteString("=== RAW HTML CONTENT ===\n")
		b.WriteString(rawHTML)
		b.WriteString("\n=========================\n")
	} else {
		b.WriteString("=== NO RAW CONTENT AVAILABLE ===\n")
		b.WriteString("Raw HTML content was not captured during scraping.\n")
		b.WriteString("===================================\n")
	}

	return b.String()
}