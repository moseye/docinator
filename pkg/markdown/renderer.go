package markdown

import (
	"bytes"
	"text/template"

	"github.com/moseye/docinator/internal/models"
)

// ConvertToMarkdown renders a Package to a Markdown string.
func ConvertToMarkdown(pkg *models.Package) string {
	tmplStr := "# {{.Name}}\n\n{{.Description}}\n\n{{if .Synopsis}}**Synopsis:** {{.Synopsis}}{{end}}\n\n{{if .Module}}**Module:** {{.Module}}{{if .Version}} ({{.Version}}){{end}}{{end}}\n\n{{if .ImportPath}}**Import Path:** {{.ImportPath}}{{end}}\n\n{{if .License}}**License:** {{.License}}{{end}}\n\n{{if .Repository}}**Repository:** {{.Repository}}{{end}}\n\n## Functions\n\n{{range .Functions}}\n### {{.Name}}{{if .Receiver}} ({{.Receiver}}){{end}}\n{{.Signature}}\n\n{{.Description}}\n\n{{if .Examples}}\n#### Examples\n{{range .Examples}}\n**{{.Name}}**\n\n```go\n{{.Code}}\n```\n\n**Output:**\n{{.Output}}\n{{end}}\n{{end}}\n{{end}}\n\n## Types\n\n{{range .Types}}\n### {{.Name}} ({{.Kind}})\n{{.Definition}}\n\n{{.Description}}\n\n{{if .Methods}}\n#### Methods\n{{range .Methods}}\n- **{{.Name}}** ({{.Signature}})\n\n  {{.Description}}\n\n  {{if .Examples}}\n  **Examples:**\n  {{range .Examples}}\n  **{{.Name}}**\n\n  ```go\n  {{.Code}}\n  ```\n\n  **Output:**\n  {{.Output}}\n  {{end}}\n  {{end}}\n  {{end}}\n  {{end}}\n\n{{if .Examples}}\n#### Examples\n{{range .Examples}}\n**{{.Name}}**\n\n```go\n{{.Code}}\n```\n\n**Output:**\n{{.Output}}\n{{end}}\n{{end}}\n{{end}}\n\n## Variables\n\n{{range .Variables}}\n### {{.Name}} {{.Type}}\n{{.Description}}\n{{end}}\n\n## Constants\n\n{{range .Constants}}\n### {{.Name}} {{.Type}} = {{.Value}}\n{{.Description}}\n{{end}}"

	tmpl, err := template.New("markdown").Parse(tmplStr)
	if err != nil {
		return "" // Basic error handling as per scope
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, pkg); err != nil {
		return ""
	}

	return buf.String()
}