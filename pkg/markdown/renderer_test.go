package markdown

import (
	"strings"
	"testing"
	"time"

	"github.com/moseye/docinator/internal/models"
)

func TestConvertToMarkdown(t *testing.T) {
	// Create a mock Package with sample data similar to cobra
	mockPkg := &models.Package{
		Name:        "testpkg",
		Description: "A test package for markdown conversion with extensive content to ensure long output. This description is expanded to add more characters and test rendering of longer texts.",
		Synopsis:    "Test Synopsis with additional details for coverage.",
		Module:      "test/module",
		Version:     "v1.0.0",
		ImportPath:  "github.com/test/pkg",
		License:     "MIT",
		Repository:  "https://github.com/test/pkg",
		ScrapedAt:   time.Now(),
		Functions: []models.Function{
			{
				Name:        "TestFunc",
				Description: "A test function for rendering. This description is detailed to increase output length.",
				Signature:   "func TestFunc() string",
				Examples: []models.Example{
					{
						Name:   "TestExample",
						Code:   "fmt.Println(\"Hello World from test example\")",
						Output: "Hello World from test example\n",
					},
					{
						Name:   "SecondExample",
						Code:   "fmt.Println(\"Another example for coverage\")",
						Output: "Another example for coverage\n",
					},
				},
			},
			{
				Name:        "AnotherFunc",
				Description: "Second function to add more content.",
				Signature:   "func AnotherFunc() error",
				Examples: []models.Example{
					{
						Name:   "AnotherExample",
						Code:   "return nil",
						Output: "",
					},
				},
			},
		},
		Types: []models.Type{
			{
				Name:        "TestType",
				Kind:        "struct",
				Definition:  "type TestType struct { Field string }",
				Description: "A test type for rendering with more details.",
				Methods: []models.Function{
					{
						Name:        "Method",
						Description: "A test method with description.",
						Signature:   "func (t *TestType) Method() string",
					},
					{
						Name:        "SecondMethod",
						Description: "Another method for the type.",
						Signature:   "func (t *TestType) SecondMethod()",
					},
				},
				Examples: []models.Example{
					{
						Name:   "TypeExample",
						Code:   "t := &TestType{Field: \"value\"}\nfmt.Println(t.Field)",
						Output: "value\n",
					},
					{
						Name:   "SecondTypeExample",
						Code:   "t.Method()",
						Output: "",
					},
				},
			},
			{
				Name:        "AnotherType",
				Kind:        "interface",
				Definition:  "type AnotherType interface {}",
				Description: "Another test type to extend output.",
				Methods: []models.Function{
					{
						Name:        "InterfaceMethod",
						Description: "Method for interface.",
						Signature:   "func (a AnotherType) InterfaceMethod()",
					},
				},
			},
		},
		Variables: []models.Variable{
			{
				Name:        "TestVar",
				Type:        "string",
				Description: "A test variable with longer description.",
			},
			{
				Name:        "AnotherVar",
				Type:        "int",
				Description: "Second variable.",
			},
		},
		Constants: []models.Constant{
			{
				Name:        "TestConst",
				Type:        "int",
				Value:       "42",
				Description: "A test constant with details.",
			},
			{
				Name:        "AnotherConst",
				Type:        "string",
				Value:       "\"constant value\"",
				Description: "Another constant.",
			},
		},
		Examples: []models.Example{
			{
				Name:   "PackageExample",
				Code:   "fmt.Println(\"Package level example\")",
				Output: "Package level example\n",
			},
		},
	}

	output := ConvertToMarkdown(mockPkg)

	// Acceptance criteria checks
	if len(output) < 1000 {
		t.Errorf("Output should be longer than 1000 chars, got %d", len(output))
	}

	if !strings.Contains(output, "# testpkg") {
		t.Error("Output should contain package name section")
	}

	if !strings.Contains(output, "## Functions") {
		t.Error("Output should contain Functions section")
	}

	if !strings.Contains(output, "TestFunc") {
		t.Error("Output should contain function name")
	}

	if !strings.Contains(output, "## Types") {
		t.Error("Output should contain Types section")
	}

	if !strings.Contains(output, "TestType") {
		t.Error("Output should contain type name")
	}

	if !strings.Contains(output, "```go") {
		t.Error("Output should contain code blocks")
	}

	t.Log("TestConvertToMarkdown passed")
}