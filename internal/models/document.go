package models

import "time"

type Package struct {
	Name            string     `bson:"name,omitempty"`
	Description     string     `bson:"description,omitempty"`
	Module          string     `bson:"module,omitempty"`
	Version         string     `bson:"version,omitempty"`
	IsLatest        bool       `bson:"is_latest,omitempty"`
	Published       string     `bson:"published,omitempty"`
	Synopsis        string     `bson:"synopsis,omitempty"`
	License         string     `bson:"license,omitempty"`
	LicenseURL      string     `bson:"license_url,omitempty"`
	Repository      string     `bson:"repository,omitempty"`
	ImportPath      string     `bson:"import_path,omitempty"`
	ScrapedAt       time.Time  `bson:"scraped_at,omitempty"`
	Readme          string     `bson:"readme,omitempty"`
	ProcessedReadme string     `bson:"processed_readme,omitempty"`
	Imports         int        `bson:"imports,omitempty"`
	ImportedBy      int        `bson:"imported_by,omitempty"`
	Functions       []Function `bson:"functions,omitempty"`
	Types           []Type     `bson:"types,omitempty"`
	Variables       []Variable `bson:"variables,omitempty"`
	Constants       []Constant `bson:"constants,omitempty"`
	Examples        []Example  `bson:"examples,omitempty"`
}

type Function struct {
	Name        string    `bson:"name,omitempty"`
	Description string    `bson:"description,omitempty"`
	Signature   string    `bson:"signature,omitempty"`
	Receiver    string    `bson:"receiver,omitempty"`
	Deprecated  string    `bson:"deprecated,omitempty"`
	AddedIn     string    `bson:"added_in,omitempty"`
	Examples    []Example `bson:"examples,omitempty"`
}

type Type struct {
	Name        string     `bson:"name,omitempty"`
	Description string     `bson:"description,omitempty"`
	Definition  string     `bson:"definition,omitempty"`
	Kind        string     `bson:"kind,omitempty"`
	Deprecated  string     `bson:"deprecated,omitempty"`
	AddedIn     string     `bson:"added_in,omitempty"`
	Methods     []Function `bson:"methods,omitempty"`
	Examples    []Example  `bson:"examples,omitempty"`
}

type Variable struct {
	Name        string `bson:"name,omitempty"`
	Type        string `bson:"type,omitempty"`
	Description string `bson:"description,omitempty"`
}

type Constant struct {
	Name        string `bson:"name,omitempty"`
	Type        string `bson:"type,omitempty"`
	Value       string `bson:"value,omitempty"`
	Description string `bson:"description,omitempty"`
}

type Example struct {
	Name   string `bson:"name,omitempty"`
	Code   string `bson:"code,omitempty"`
	Output string `bson:"output,omitempty"`
}

type Document struct {
	ID      string   `bson:"_id"`                // import path as primary key, e.g., "github.com/spf13/cobra"
	Package *Package `bson:"package"`            // structured package data
	RawHTML string   `bson:"raw_html,omitempty"` // raw HTML content from the scraped page
}
