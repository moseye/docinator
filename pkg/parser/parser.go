package parser

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/moseye/docinator/internal/models"
	"github.com/moseye/docinator/internal/utils"
)

// Parser handles HTML parsing operations for pkg.go.dev pages
type Parser struct {
	// Future: Add any parser configuration here
}

// New creates a new Parser instance
func New() *Parser {
	return &Parser{}
}

// ParsePackagePage parses a pkg.go.dev package page and extracts structured data
func (p *Parser) ParsePackagePage(e *colly.HTMLElement) (*models.Package, error) {
	doc := e.DOM
	pkg := &models.Package{}

	// Extract metadata
	// Package Name from title heading
	if el := doc.Find("h1.UnitHeader-titleHeading"); el.Length() > 0 {
		pkg.Name = strings.TrimSpace(el.Text())
		log.Printf("Set package name to: %s", pkg.Name)
	}

	// Import Path from breadcrumb current
	if el := doc.Find(".UnitHeader-breadcrumbCurrent"); el.Length() > 0 {
		text := strings.TrimSpace(el.Text())
		if text != "" {
			pkg.ImportPath = text
			log.Printf("Set import path to: %s", pkg.ImportPath)
		}
	}

	// Version from aria-label (more reliable)
	if el := doc.Find("a[aria-label^='Version: ']"); el.Length() > 0 {
		ariaLabel := el.AttrOr("aria-label", "")
		if strings.HasPrefix(ariaLabel, "Version: ") {
			pkg.Version = strings.TrimPrefix(ariaLabel, "Version: ")
			log.Printf("Set version to: %s", pkg.Version)
		}
	}

	// IsLatest from badge (support multiple possible class names)
	if doc.Find(".DetailsHeader-badge--latest, .UnitHeader-badge--latest, .DetailsHeader-span--latest").Length() > 0 {
		if strings.Contains(doc.Find(".DetailsHeader-badge--latest, .UnitHeader-badge--latest, .DetailsHeader-span--latest").Text(), "Latest") {
			pkg.IsLatest = true
			log.Printf("Package is latest version")
		}
	}

	// Published from span with data-test-id
	if el := doc.Find("[data-test-id='UnitHeader-commitTime']"); el.Length() > 0 {
		text := strings.TrimSpace(el.Text())
		if strings.HasPrefix(text, "Published: ") {
			pkg.Published = strings.TrimSpace(strings.TrimPrefix(text, "Published: "))
			log.Printf("Set published to: %s", pkg.Published)
		}
	}

	// Extract license information (set License and LicenseURL)
	e.ForEach("a[data-test-id='UnitHeader-license'], [data-test-id='UnitHeader-licenses'] a, .UnitHeader-license a", func(_ int, el *colly.HTMLElement) {
		licenseText := strings.TrimSpace(el.Text)
		licenseHref := el.Attr("href")
		if licenseText == "" && el.DOM != nil {
			licenseText = strings.TrimSpace(el.DOM.Text())
		}
		if licenseText != "" {
			pkg.License = licenseText
			if licenseHref != "" {
				// Normalize to absolute pkg.go.dev URL if relative
				if strings.HasPrefix(licenseHref, "/") {
					pkg.LicenseURL = "https://pkg.go.dev" + licenseHref
				} else {
					pkg.LicenseURL = licenseHref
				}
			}
			log.Printf("Set license to: %s, URL: %s", pkg.License, pkg.LicenseURL)
		}
	})

	// Imports
	if el := doc.Find("[data-test-id='UnitHeader-imports'] a"); el.Length() > 0 {
		aria := el.AttrOr("aria-label", "")
		// Fallback to text content if aria-label missing
		value := aria
		if value == "" {
			value = strings.TrimSpace(el.Text())
		}
		// Expect formats like "Imports: 17"
		if strings.HasPrefix(value, "Imports: ") {
			countStr := strings.TrimSpace(strings.TrimPrefix(value, "Imports: "))
			countStr = strings.ReplaceAll(countStr, ",", "")
			if num, err := strconv.Atoi(countStr); err == nil {
				pkg.Imports = num
				log.Printf("Set imports to: %d", pkg.Imports)
			}
		}
	}

	// Imported By
	if el := doc.Find("[data-test-id='UnitHeader-importedby'] a"); el.Length() > 0 {
		aria := el.AttrOr("aria-label", "")
		// Fallback to text content if aria-label missing
		value := aria
		if value == "" {
			value = strings.TrimSpace(el.Text())
		}
		// Expect formats like "Imported By: 177,680"
		prefixes := []string{"Imported By: ", "Imported by: "}
		for _, pfx := range prefixes {
			if strings.HasPrefix(value, pfx) {
				countStr := strings.TrimSpace(strings.TrimPrefix(value, pfx))
				countStr = strings.ReplaceAll(countStr, ",", "")
				if num, err := strconv.Atoi(countStr); err == nil {
					pkg.ImportedBy = num
					log.Printf("Set imported by to: %d", pkg.ImportedBy)
				}
				break
			}
		}
	}

	// Extract repository URL
	e.ForEach(".UnitMeta-repo a", func(_ int, el *colly.HTMLElement) {
		pkg.Repository = strings.TrimSpace(el.Attr("href"))
	})

	// Synopsis / Description (prefer overview paragraph)
	if el := doc.Find(".Documentation-overview p"); el.Length() > 0 {
		pkg.Description = strings.TrimSpace(el.First().Text())
		log.Printf("Set synopsis/description to: %s", pkg.Description)
	}

	// README HTML
	if el := doc.Find(".UnitReadme-content .Overview-readmeContent"); el.Length() > 0 {
		html, err := el.Html()
		if err == nil {
			pkg.Readme = html
			pkg.ProcessedReadme = utils.ConvertHTMLToMarkdown(html)
			log.Printf("Extracted and converted README")
		}
	}

	// Constants: iterate declaration blocks and extract pre + adjacent description
	doc.Find(".Documentation-constants .Documentation-declaration").Each(func(i int, s *goquery.Selection) {
		pre := s.Find("pre").First()
		code := strings.TrimSpace(pre.Text())
		if code == "" {
			return
		}
		// Derive a name from the first span id within this block, fallback to block index
		name := strings.TrimSpace(pre.Find("span[id][data-kind='constant']").First().AttrOr("id", "const-block-"+strconv.Itoa(i+1)))
		// Try to capture the description paragraph immediately following the declaration
		descSel := s.NextAllFiltered("p").First()
		desc := strings.TrimSpace(descSel.Text())
		constant := models.Constant{Name: name, Value: code, Description: desc}
		pkg.Constants = append(pkg.Constants, constant)
		log.Printf("Added constant block: %s", name)
	})
	// Variables: iterate declaration blocks and extract pre + adjacent description
	doc.Find(".Documentation-variables .Documentation-declaration").Each(func(i int, s *goquery.Selection) {
		pre := s.Find("pre").First()
		code := strings.TrimSpace(pre.Text())
		if code == "" {
			return
		}
		// Derive a name from the first span id within this block, fallback to block index
		name := strings.TrimSpace(pre.Find("span[id][data-kind='variable']").First().AttrOr("id", "var-block-"+strconv.Itoa(i+1)))
		// Try to capture the description paragraph immediately following the declaration
		descSel := s.NextAllFiltered("p").First()
		desc := strings.TrimSpace(descSel.Text())
		variable := models.Variable{Name: name, Type: code, Description: desc}
		pkg.Variables = append(pkg.Variables, variable)
		log.Printf("Added variable block: %s", name)
	})
	// Functions
	doc.Find(".Documentation-functions .Documentation-function").Each(func(i int, s *goquery.Selection) {

		header := s.Find("h4").First()

		id := header.AttrOr("id", "")

		// Prefer signature from declaration <pre> (reliable)
		sig := strings.TrimSpace(s.Find(".Documentation-declaration pre").First().Text())

		if sig != "" {

			// AddedIn version (just the version token, e.g., v1.1.2, if available)
			addedIn := strings.TrimSpace(s.Find(".Documentation-sinceVersionVersion").First().Text())
			if addedIn == "" {
				// fallback to entire sinceVersion text
				addedIn = strings.TrimSpace(s.Find(".Documentation-sinceVersion").First().Text())
			}

			// Description is the first paragraph under the function block
			desc := strings.TrimSpace(s.Find("p").First().Text())

			deprecated := ""
			if s.Find(".Documentation-deprecatedTag").Length() > 0 {
				deprecated = "deprecated"
			}

			function := models.Function{Name: id, Signature: sig, Description: desc, Deprecated: deprecated, AddedIn: addedIn}

			pkg.Functions = append(pkg.Functions, function)

			log.Printf("Added function: %s", id)

		}

	})

	// Types
	doc.Find(".Documentation-types .Documentation-type").Each(func(i int, s *goquery.Selection) {

		header := s.Find("h4").First()

		id := header.AttrOr("id", "")

		// Type definition from declaration <pre>
		def := strings.TrimSpace(s.Find(".Documentation-declaration pre").First().Text())

		if id != "" && def != "" {

			// AddedIn version
			addedIn := strings.TrimSpace(s.Find(".Documentation-sinceVersionVersion").First().Text())
			if addedIn == "" {
				addedIn = strings.TrimSpace(s.Find(".Documentation-sinceVersion").First().Text())
			}

			// Description: pick first paragraph in the type block (after declaration)
			desc := strings.TrimSpace(s.Find("p").First().Text())

			deprecated := ""
			if s.Find(".Documentation-deprecatedTag").Length() > 0 {
				deprecated = "deprecated"
			}

			typeInfo := models.Type{Name: id, Definition: def, Kind: "type", Description: desc, Deprecated: deprecated, AddedIn: addedIn}

			// Methods
			s.Find(".Documentation-typeMethod").Each(func(j int, methodSel *goquery.Selection) {

				// Method header id, e.g., Command.AddCommand
				mh := methodSel.Find("h4").First()
				mName := mh.AttrOr("id", "")
				if mName == "" {
					mName = strings.TrimSpace(mh.Text())
				}

				// Signature from declaration pre
				mSig := strings.TrimSpace(methodSel.Find(".Documentation-declaration pre").First().Text())

				// Description (first paragraph within the method block)
				mDesc := strings.TrimSpace(methodSel.Find("p").First().Text())

				// Since version
				mAddedIn := strings.TrimSpace(methodSel.Find(".Documentation-sinceVersionVersion").First().Text())
				if mAddedIn == "" {
					mAddedIn = strings.TrimSpace(methodSel.Find(".Documentation-sinceVersion").First().Text())
				}

				// Deprecated tag
				mDeprecated := ""
				if methodSel.Find(".Documentation-deprecatedTag").Length() > 0 {
					mDeprecated = "deprecated"
				}

				if mSig != "" || mName != "" {
					method := models.Function{Name: mName, Signature: mSig, Description: mDesc, Deprecated: mDeprecated, AddedIn: mAddedIn}
					typeInfo.Methods = append(typeInfo.Methods, method)
				}
			})

			pkg.Types = append(pkg.Types, typeInfo)

			log.Printf("Added type: %s", id)

		}

	})

	return pkg, nil
}
