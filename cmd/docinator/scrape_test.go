package docinator

import (
	"bytes"
	"testing"
)

func TestScrapeCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "Basic scrape command",
			args: []string{"github.com/spf13/cobra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			rootCmd.SetArgs(append([]string{"scrape", "--test-mode"}, tt.args...) )
			scrapeCmd.SetOut(&buf)
			err := rootCmd.Execute()
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if len(buf.Bytes()) == 0 {
				t.Errorf("Expected output, got empty")
			}
		})
	}
}