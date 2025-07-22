package cmd

import (
	"fmt"
	"strings"

	"github.com/benekenobi/colordna/internal/colorer"
	"github.com/benekenobi/colordna/internal/config"
	"github.com/spf13/cobra"
)

// previewCmd represents the preview command
var previewCmd = &cobra.Command{
	Use:   "preview [scheme-name]",
	Short: "Preview a color scheme",
	Long: `Preview shows how DNA/RNA sequences and quality scores will look
with the specified color scheme. If no scheme is specified, it shows
all available schemes.`,
	RunE: runPreview,
}

func runPreview(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If specific scheme requested
	if len(args) > 0 {
		schemeName := args[0]
		scheme, exists := cfg.ColorSchemes[schemeName]
		if !exists {
			return fmt.Errorf("color scheme '%s' not found", schemeName)
		}

		fmt.Printf("Color scheme: %s\n", schemeName)
		fmt.Println(strings.Repeat("-", 40))
		showSchemePreview(scheme)
		return nil
	}

	// Show all schemes
	fmt.Println("Available color schemes:")
	fmt.Println(strings.Repeat("=", 50))

	for name, scheme := range cfg.ColorSchemes {
		fmt.Printf("\nScheme: %s", name)
		if name == colorScheme {
			fmt.Print(" (current)")
		}
		fmt.Println()
		fmt.Println(strings.Repeat("-", 40))
		showSchemePreview(scheme)
	}

	return nil
}

func showSchemePreview(scheme config.ColorScheme) {
	colorizer := colorer.New(scheme)

	// DNA sequence preview
	dnaSeq := "ATGCGATCGATCGTAG"
	fmt.Printf("DNA:     %s\n", colorizer.ColorizeSequence(dnaSeq))

	// RNA sequence preview
	rnaSeq := "AUGCGAUCGAUCGUAG"
	fmt.Printf("RNA:     %s\n", colorizer.ColorizeSequence(rnaSeq))

	// Quality score preview
	if scheme.Quality == "gradient" {
		quality := "!\"#)*+./:9?EFIJK"
		fmt.Printf("Quality: %s\n", colorizer.ColorizeQuality(quality))
		fmt.Printf("         %s\n", "Poor -> Good Quality")
	} else if scheme.Quality == "mono" {
		quality := "!\"#)*+./:9?EFIJK"
		fmt.Printf("Quality: %s\n", colorizer.ColorizeQuality(quality))
		fmt.Printf("         %s\n", "Dim -> Bold Quality")
	}
}

func init() {
	rootCmd.AddCommand(previewCmd)
}
