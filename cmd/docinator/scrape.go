package docinator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/moseye/docinator/internal/models"
	mongostore "github.com/moseye/docinator/internal/storage/mongo"
	"github.com/moseye/docinator/pkg/markdown"
	"github.com/moseye/docinator/pkg/raw"
	"github.com/moseye/docinator/pkg/scraper"
	"github.com/spf13/cobra"
)

var scrapeCmd = &cobra.Command{
	Use:   "scrape [packages...]",
	Short: "Scrape documentation from Go packages",
	Long: `Scrape the documentation from one or more Go packages on pkg.go.dev,
parse the content, and generate markdown files.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := rootCmd.PersistentFlags().GetBool("verbose")
		testMode, _ := rootCmd.PersistentFlags().GetBool("test-mode")
		outputDir, _ := rootCmd.PersistentFlags().GetString("output")
		log.Printf("TestMode: %v", testMode)
		log.Printf("Starting scrape command with args: %v, verbose: %v, outputDir: %v", args, verbose, outputDir)

		config := &scraper.ScrapingConfig{
			Debug:    verbose,
			TestMode: testMode,
		}
		s, err := scraper.New(config)
		if err != nil {
			log.Fatalf("Failed to create scraper: %v", err)
		}
		defer s.Close()
		log.Printf("Scraper created successfully")

		ctx := cmd.Context()

		// Initialize MongoDB store (disabled if MONGODB_URI is not set)
		store, err := mongostore.NewFromEnv(ctx)
		if err != nil {
			log.Printf("MongoDB store initialization error (disabled): %v", err)
			store = nil
		}
		if store != nil && store.Enabled() {
			defer func() {
				if err := store.Close(ctx); err != nil {
					log.Printf("MongoDB disconnect error: %v", err)
				}
			}()
		}

		// Scrape packages with both structured data and raw HTML
		var pkgs []*models.Package
		var rawHTMLs []string
		var scrapeErrors []error

		for _, importPath := range args {
			// 1) Check MongoDB cache first
			if store != nil && store.Enabled() {
				doc, err := store.GetByID(ctx, importPath)
				if err != nil {
					log.Printf("MongoDB lookup error for %s: %v", importPath, err)
				} else if doc != nil && doc.Package != nil {
					pkgs = append(pkgs, doc.Package)
					rawHTMLs = append(rawHTMLs, doc.RawHTML)
					if verbose {
						log.Printf("Loaded from MongoDB cache: %s", importPath)
					}
					continue
				}
			}

			// 2) Not cached → scrape
			pkg, rawHTML, err := s.ScrapePackageWithRaw(ctx, importPath)
			if err != nil {
				scrapeErrors = append(scrapeErrors, fmt.Errorf("failed to scrape %s: %w", importPath, err))
				continue
			}
			pkgs = append(pkgs, pkg)
			rawHTMLs = append(rawHTMLs, rawHTML)

			// 3) Persist to MongoDB (upsert) for future runs
			if store != nil && store.Enabled() {
				id := importPath
				if pkg != nil && pkg.ImportPath != "" {
					id = pkg.ImportPath
				}
				doc := &models.Document{
					ID:      id,
					Package: pkg,
					RawHTML: rawHTML,
				}
				if err := store.Upsert(ctx, doc); err != nil {
					log.Printf("MongoDB upsert failed for %s: %v", id, err)
				} else if verbose {
					log.Printf("Upserted into MongoDB: %s", id)
				}
			}
		}

		if len(scrapeErrors) > 0 {
			for _, err := range scrapeErrors {
				log.Printf("Scraping error: %v", err)
			}
			if len(pkgs) == 0 {
				log.Fatalf("All scraping attempts failed")
			}
		}

		log.Printf("Successfully scraped %d packages", len(pkgs))

		if outputDir == "" {
			// Output to stdout (markdown only for readability)
			for _, pkg := range pkgs {
				log.Printf("Generating markdown for package: %s", pkg.ImportPath)
				cmd.Print(markdown.PackageToMarkdown(pkg))
			}
		} else {
			// Output to files - both markdown and raw versions
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Fatalf("Failed to create output dir: %v", err)
			}

			for i, pkg := range pkgs {
				log.Printf("Generating both formats for package: %s", pkg.ImportPath)

				// Generate markdown file
				markdownFilename := fmt.Sprintf("%s/%s.md", outputDir, pkg.ImportPath)
				markdownContent := markdown.PackageToMarkdown(pkg)

				markdownDir := filepath.Dir(markdownFilename)
				if err := os.MkdirAll(markdownDir, 0755); err != nil {
					log.Printf("Failed to create markdown dir %s: %v", markdownDir, err)
				}

				if err := os.WriteFile(markdownFilename, []byte(markdownContent), 0644); err != nil {
					log.Printf("Failed to write markdown file %s: %v", markdownFilename, err)
				} else if verbose {
					log.Printf("Wrote markdown: %s", markdownFilename)
				}

				// Generate raw HTML file
				rawFilename := fmt.Sprintf("%s/%s_raw.txt", outputDir, pkg.ImportPath)
				rawContent := raw.PackageToRaw(pkg, rawHTMLs[i])

				rawDir := filepath.Dir(rawFilename)
				if err := os.MkdirAll(rawDir, 0755); err != nil {
					log.Printf("Failed to create raw dir %s: %v", rawDir, err)
				}

				if err := os.WriteFile(rawFilename, []byte(rawContent), 0644); err != nil {
					log.Printf("Failed to write raw file %s: %v", rawFilename, err)
				} else if verbose {
					log.Printf("Wrote raw version: %s", rawFilename)
				}
			}
		}

		if verbose {
			stats := s.GetStats()
			log.Printf("Scraped %d packages, %d requests, %d errors", stats.PackagesScraped, stats.RequestsMade, stats.Errors)
		}
	},
}
