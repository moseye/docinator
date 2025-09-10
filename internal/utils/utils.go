package utils

import (
	"regexp"
	"strconv"
	"strings"
)

// ConvertHTMLToMarkdown provides a simple, dependency-free HTML → Markdown conversion.
// This is a lightweight best-effort converter intended for README blocks scraped from pkg.go.dev.
func ConvertHTMLToMarkdown(html string) string {
	if strings.TrimSpace(html) == "" {
		return html
	}

	// Normalize newlines
	html = strings.ReplaceAll(html, "\r\n", "\n")

	// Unescape a few common entities early to avoid interfering with regex groups
	html = unescapeCommonEntities(html)

	// 1) Fenced code blocks: <pre><code ...>...</code></pre> → ```go ... ```
	// Try to capture language if present in class attribute (e.g., class="language-go")
	preCodeRe := regexp.MustCompile(`(?is)<pre>\s*<code([^>]*)>(.*?)</code>\s*</pre>`)
	html = preCodeRe.ReplaceAllStringFunc(html, func(m string) string {
		matches := preCodeRe.FindStringSubmatch(m)
		attrs := matches[1]
		code := matches[2]
		lang := ""
		if strings.Contains(strings.ToLower(attrs), "language-go") {
			lang = "go"
		}
		// Make sure inner HTML entities are unescaped inside code blocks
		code = unescapeCommonEntities(code)
		// Ensure code content does not contain stray backtick fences; leave as-is best-effort
		if lang != "" {
			return "```" + lang + "\n" + code + "\n```\n\n"
		}
		return "```\n" + code + "\n```\n\n"
	})

	// 2) Headings with attributes: <h1 ...>Text</h1> → # Text (supports h1..h6)
	for level := 1; level <= 6; level++ {
		lit := strconv.Itoa(level)
		tag := regexp.MustCompile(`(?is)<h` + lit + `[^>]*>(.*?)</h` + lit + `>`)
		prefix := strings.Repeat("#", level) + " "
		html = tag.ReplaceAllString(html, prefix+`$1`+"\n\n")
	}

	// 3) Anchors: <a ... href="URL" ...>TEXT</a> -> [TEXT](URL)
	anchorRe := regexp.MustCompile(`(?is)<a[^>]*\shref="([^"]+)"[^>]*>(.*?)</a>`)
	html = anchorRe.ReplaceAllString(html, `[$2]($1)`)

	// 4) Images: <img ... alt="ALT" ... src="SRC" .../> -> ![ALT](SRC)
	// alt then src
	imgAltSrcRe := regexp.MustCompile(`(?is)<img[^>]*\salt="([^"]*)"[^>]*\ssrc="([^"]+)"[^>]*/?>`)
	html = imgAltSrcRe.ReplaceAllString(html, `![$1]($2)`)
	// src then alt
	imgSrcAltRe := regexp.MustCompile(`(?is)<img[^>]*\ssrc="([^"]+)"[^>]*\salt="([^"]*)"[^>]*/?>`)
	html = imgSrcAltRe.ReplaceAllString(html, `![$2]($1)`)
	// Fallback: src only
	imgSrcOnlyRe := regexp.MustCompile(`(?is)<img[^>]*\ssrc="([^"]+)"[^>]*/?>`)
	html = imgSrcOnlyRe.ReplaceAllString(html, `![]($1)`)

	// 5) Inline code: <code>text</code> -> `text` (avoid inside already converted fenced blocks)
	inlineCodeRe := regexp.MustCompile(`(?is)<code>(.*?)</code>`)
	html = inlineCodeRe.ReplaceAllString(html, "`$1`")

	// 6) Paragraphs, lists, blockquotes, emphasis, breaks, hr
	replacements := map[string]string{
		"<p>":           "\n",
		"</p>":          "\n\n",
		"<br>":          "\n",
		"<br/>":         "\n",
		"<strong>":      "**",
		"</strong>":     "**",
		"<b>":           "**",
		"</b>":          "**",
		"<em>":          "*",
		"</em>":         "*",
		"<i>":           "*",
		"</i>":          "*",
		"<ul>":          "\n",
		"</ul>":         "\n",
		"<ol>":          "\n",
		"</ol>":         "\n",
		"<li>":          "- ",
		"</li>":         "\n",
		"<blockquote>":  "> ",
		"</blockquote>": "\n",
		"<hr>":          "\n---\n",
		"<hr/>":         "\n---\n",
	}
	for from, to := range replacements {
		html = strings.ReplaceAll(html, from, to)
	}

	// 7) Remove any remaining tags conservatively
	html = stripAllTags(html)

	// 8) Fix common fence issues where a lone backtick line appears instead of closing triple fence
	html = strings.ReplaceAll(html, "\n`\n", "\n```\n")
	html = strings.ReplaceAll(html, "\n`\n\n", "\n```\n\n")

	// 9) Final entity unescape and whitespace cleanup
	html = unescapeCommonEntities(html)
	for strings.Contains(html, "\n\n\n") {
		html = strings.ReplaceAll(html, "\n\n\n", "\n\n")
	}
	html = strings.TrimSpace(html)
	return html
}

// LooksLikeHTML returns true if the input string appears to contain HTML tags.
func LooksLikeHTML(s string) bool {
	if s == "" {
		return false
	}
	// Heuristic: contains a tag-like pattern and a closing angle bracket
	return strings.Contains(s, "<") && strings.Contains(s, ">")
}

// stripAllTags removes any residual <...> tags.
func stripAllTags(s string) string {
	for {
		start := strings.Index(s, "<")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], ">")
		if end == -1 {
			break
		}
		s = s[:start] + s[start+end+1:]
	}
	return s
}

// unescapeCommonEntities replaces a subset of common HTML entities for readability.
func unescapeCommonEntities(s string) string {
	repl := map[string]string{
		"&quot;": `"`,
		"&#34;":  `"`,
		"&apos;": `'`,
		"&#39;":  `'`,
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&nbsp;": " ",
	}
	for k, v := range repl {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}
