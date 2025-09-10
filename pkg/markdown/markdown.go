package markdown

import (
	"fmt"
	"strings"

	"github.com/moseye/docinator/internal/models"
)

// PackageToMarkdown converts a Package struct to a professional markdown formatted string matching pkg.go.dev style.
func PackageToMarkdown(pkg *models.Package) string {
	var b strings.Builder

	// Professional header with import path (expected format)
	header := fmt.Sprintf("# %s package - %s", pkg.Name, pkg.ImportPath)
	b.WriteString(header + "\n\n")

	// Package metadata section
	b.WriteString("## Package Documentation\n\n")

	// Import Path
	if pkg.ImportPath != "" {
		b.WriteString(fmt.Sprintf("**Import Path:** `%s`\n\n", pkg.ImportPath))
	}

	// Module
	if pkg.Module != "" {
		b.WriteString(fmt.Sprintf("**Module:** %s\n\n", pkg.Module))
	}

	// Version with status
	if pkg.Version != "" {
		versionText := pkg.Version
		if pkg.IsLatest {
			versionText += " (Latest)"
		}
		b.WriteString(fmt.Sprintf("**Version:** %s\n\n", versionText))
	}

	// Published date
	if pkg.Published != "" {
		b.WriteString(fmt.Sprintf("**Published:** %s\n\n", pkg.Published))
	}

	// Imports
	if pkg.Imports > 0 {
		b.WriteString(fmt.Sprintf("**Imports:** %d\n\n", pkg.Imports))
	}

	// Imported By (comma formatting for readability)
	if pkg.ImportedBy > 0 {
		b.WriteString(fmt.Sprintf("**Imported By:** %s\n\n", formatNumber(pkg.ImportedBy)))
	}

	// License with link
	if pkg.License != "" && pkg.LicenseURL != "" {
		b.WriteString(fmt.Sprintf("**License:** [%s](%s)\n\n", pkg.License, pkg.LicenseURL))
	} else if pkg.License != "" {
		b.WriteString(fmt.Sprintf("**License:** %s\n\n", pkg.License))
	}

	// Repository link
	if pkg.Repository != "" {
		// Display a clean label (strip scheme) but keep full URL for the link
		label := pkg.Repository
		if strings.HasPrefix(label, "https://") {
			label = strings.TrimPrefix(label, "https://")
		} else if strings.HasPrefix(label, "http://") {
			label = strings.TrimPrefix(label, "http://")
		}
		b.WriteString(fmt.Sprintf("**Repository:** [%s](%s)\n\n", label, pkg.Repository))
	}

	// Overview/Synopsis
	if pkg.Synopsis != "" {
		b.WriteString("## Overview\n\n")
		b.WriteString(pkg.Synopsis + "\n\n")
	} else if pkg.Description != "" {
		b.WriteString("## Overview\n\n")
		b.WriteString(pkg.Description + "\n\n")
	}

	// README section with processed markdown
	b.WriteString("## README\n\n")
	if pkg.ProcessedReadme != "" {
		b.WriteString(pkg.ProcessedReadme)
	} else if pkg.Readme != "" {
		// Fallback to raw HTML if not processed
		b.WriteString(pkg.Readme)
	}
	b.WriteString("\n\n")

	// Documentation Index
	b.WriteString("## Documentation\n\n")
	b.WriteString("### Index\n\n")

	// Index entries for constants, variables, functions, types
	if len(pkg.Constants) > 0 {
		b.WriteString("#### Constants\n")
		for _, c := range pkg.Constants {
			b.WriteString(fmt.Sprintf("- [`%s`](#pkg-constants)\n", c.Name))
		}
		b.WriteString("\n")
	}

	if len(pkg.Variables) > 0 {
		b.WriteString("#### Variables\n")
		for _, v := range pkg.Variables {
			b.WriteString(fmt.Sprintf("- [`%s`](#pkg-variables)\n", v.Name))
		}
		b.WriteString("\n")
	}

	if len(pkg.Functions) > 0 {
		b.WriteString("#### Functions\n")
		for _, f := range pkg.Functions {
			// Use exact id-based anchor produced by pkg.go.dev (case-sensitive)
			b.WriteString(fmt.Sprintf("- [`%s`](#%s)\n", f.Name, f.Name))
		}
		b.WriteString("\n")
	}

	if len(pkg.Types) > 0 {
		b.WriteString("#### Types\n")
		for _, t := range pkg.Types {
			// Use exact id-based anchor for types
			b.WriteString(fmt.Sprintf("- [`%s`](#%s)\n", t.Name, t.Name))
		}
		b.WriteString("\n")
	}

	// Constants section
	if len(pkg.Constants) > 0 {
		b.WriteString("### Constants\n\n")
		for _, c := range pkg.Constants {
			b.WriteString(fmt.Sprintf("#### %s\n\n", c.Name))
			// Prefer rendering constant declaration as fenced code if multi-line or looks like code
			if c.Value != "" {
				if strings.Contains(c.Value, "\n") || strings.Contains(c.Value, "const ") {
					b.WriteString("```go\n")
					b.WriteString(c.Value)
					b.WriteString("\n```\n\n")
				} else {
					b.WriteString(fmt.Sprintf("**Value:** `%s`\n\n", c.Value))
				}
			}
			if c.Type != "" {
				b.WriteString(fmt.Sprintf("**Type:** `%s`\n\n", c.Type))
			}
			if c.Description != "" {
				b.WriteString(fmt.Sprintf("%s\n\n", c.Description))
			}
		}
	}

	// Variables section
	if len(pkg.Variables) > 0 {
		b.WriteString("### Variables\n\n")
		for _, v := range pkg.Variables {
			b.WriteString(fmt.Sprintf("#### %s\n\n", v.Name))
			// Prefer rendering variable declaration as fenced code if multi-line or looks like code
			if v.Type != "" {
				if strings.Contains(v.Type, "\n") || strings.Contains(v.Type, "var ") {
					b.WriteString("```go\n")
					b.WriteString(v.Type)
					b.WriteString("\n```\n\n")
				} else {
					b.WriteString(fmt.Sprintf("**Type:** `%s`\n\n", v.Type))
				}
			}
			if v.Description != "" {
				b.WriteString(fmt.Sprintf("%s\n\n", v.Description))
			}
		}
	}

	// Functions section
	if len(pkg.Functions) > 0 {
		b.WriteString("### Functions\n\n")
		for _, f := range pkg.Functions {
			b.WriteString(fmt.Sprintf("#### %s\n\n", f.Name))
			if f.Signature != "" {
				b.WriteString("```go\n")
				b.WriteString(f.Signature)
				b.WriteString("\n```\n\n")
			}
			if f.Description != "" {
				b.WriteString(f.Description)
				b.WriteString("\n")
			}
			// Since / Deprecated tags
			if f.AddedIn != "" {
				b.WriteString(fmt.Sprintf("_Since: %s_\n", f.AddedIn))
			}
			if f.Deprecated != "" {
				b.WriteString("**deprecated**\n")
			}
			b.WriteString("\n")
			addExamples(&b, f.Examples)
		}
	}

	// Types section
	if len(pkg.Types) > 0 {
		b.WriteString("### Types\n\n")
		for _, t := range pkg.Types {
			b.WriteString(fmt.Sprintf("#### %s\n\n", t.Name))
			if t.Definition != "" {
				b.WriteString("```go\n")
				b.WriteString(t.Definition)
				b.WriteString("\n```\n\n")
			}
			if t.Kind != "" {
				b.WriteString(fmt.Sprintf("**Kind:** %s\n\n", t.Kind))
			}
			if t.Description != "" {
				b.WriteString(t.Description)
				b.WriteString("\n")
			}
			// Since / Deprecated tags
			if t.AddedIn != "" {
				b.WriteString(fmt.Sprintf("_Since: %s_\n", t.AddedIn))
			}
			if t.Deprecated != "" {
				b.WriteString("**deprecated**\n")
			}
			b.WriteString("\n")
			// Methods
			if len(t.Methods) > 0 {
				b.WriteString("##### Methods\n\n")
				for _, m := range t.Methods {
					b.WriteString(fmt.Sprintf("###### %s\n\n", m.Name))
					if m.Signature != "" {
						b.WriteString("```go\n")
						b.WriteString(m.Signature)
						b.WriteString("\n```\n\n")
					}
					if m.Description != "" {
						b.WriteString(m.Description)
						b.WriteString("\n")
					}
					if m.AddedIn != "" {
						b.WriteString(fmt.Sprintf("_Since: %s_\n", m.AddedIn))
					}
					if m.Deprecated != "" {
						b.WriteString("**deprecated**\n")
					}
					b.WriteString("\n")
					addExamples(&b, m.Examples)
				}
			}
			addExamples(&b, t.Examples)
		}
	}

	// Package-level examples
	if len(pkg.Examples) > 0 {
		b.WriteString("### Examples\n\n")
		addExamples(&b, pkg.Examples)
	}

	// Footer with scraped timestamp
	b.WriteString(fmt.Sprintf("\n*Scraped at: %s*\n", pkg.ScrapedAt.Format("2006-01-02 15:04:05")))

	return b.String()
}

// formatNumber formats large numbers with commas
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return addCommas(n)
}

// addCommas adds commas to large numbers
func addCommas(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

// addExamples appends example markdown to the builder
func addExamples(b *strings.Builder, examples []models.Example) {
	if len(examples) == 0 {
		return
	}
	for _, ex := range examples {
		if ex.Name != "" {
			b.WriteString(fmt.Sprintf("###### %s\n\n", ex.Name))
		}
		if ex.Code != "" {
			b.WriteString("```go\n")
			b.WriteString(ex.Code)
			b.WriteString("\n```\n\n")
		}
		if ex.Output != "" {
			b.WriteString("**Output:**\n")
			b.WriteString("```\n")
			b.WriteString(ex.Output)
			b.WriteString("\n```\n\n")
		}
	}
}
