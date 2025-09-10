package docinator

import (
	"log"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "docinator",
	Short: "Documentation Web Scraper",
	Long: `A CLI tool for scraping documentation from Go packages on pkg.go.dev
and converting it to markdown format.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringP("output", "o", "", "output directory (default stdout)")
	rootCmd.PersistentFlags().Bool("test-mode", false, "enable test mode for mock data")
	if err := rootCmd.MarkPersistentFlagDirname("output"); err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(scrapeCmd)
}