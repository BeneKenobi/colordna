package cmd

import (
	"fmt"
	"sort"

	"github.com/benekenobi/colordna/internal/config"
	"github.com/spf13/cobra"
)

// schemesCmd represents the schemes command
var schemesCmd = &cobra.Command{
	Use:   "schemes",
	Short: "List available color schemes",
	Long: `List all available color schemes from the configuration file.
Use 'colordna preview [scheme-name]' to see how a specific scheme looks.`,
	RunE: runSchemes,
}

func runSchemes(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Available color schemes:")
	fmt.Println()

	// Sort scheme names for consistent output
	var names []string
	for name := range cfg.ColorSchemes {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		scheme := cfg.ColorSchemes[name]
		status := ""
		if name == colorScheme {
			status = " (default)"
		}

		backgroundInfo := ""
		if scheme.Background {
			backgroundInfo = " [background colors]"
		} else {
			backgroundInfo = " [font colors only]"
		}

		fmt.Printf("  %s%s%s\n", name, status, backgroundInfo)
	}

	fmt.Println()
	fmt.Printf("Current default: %s\n", colorScheme)
	fmt.Println("Use 'colordna preview [scheme-name]' to see how a scheme looks.")
	fmt.Println("Use 'colordna --scheme [scheme-name]' to use a different scheme.")

	return nil
}

func init() {
	rootCmd.AddCommand(schemesCmd)
}
