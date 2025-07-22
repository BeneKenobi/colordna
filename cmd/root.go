package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/benekenobi/colordna/internal/colorer"
	"github.com/benekenobi/colordna/internal/config"
	"github.com/benekenobi/colordna/internal/parser"
	"github.com/spf13/cobra"
)

var (
	colorScheme string
	configFile  string
	verbose     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "colordna [file...]",
	Short: "Color DNA/RNA sequences and quality scores in terminal output",
	Long: `colordna is a command-line tool that colorizes DNA/RNA sequences and quality scores
for better visualization in the terminal. It supports multiple file formats including
FASTA, FASTQ, SAM, and VCF with automatic format detection.

Features:
- Automatic file format detection (FASTA, FASTQ, SAM, VCF)
- Multiple color schemes with customizable colors
- Support for both sequence and quality score coloring
- Pipe support for streaming data
- Custom color scheme configuration

Examples:
  colordna sequences.fasta
  colordna --scheme bright sequences.fastq
  cat file.sam | colordna
  colordna file1.fasta file2.fastq`,
	RunE:               runColordna,
	DisableFlagParsing: false,
	DisableAutoGenTag:  true,
	Args:               cobra.ArbitraryArgs, // Allow any number of file arguments
}

func runColordna(cmd *cobra.Command, args []string) error {
	// Load configuration
	if verbose {
		fmt.Fprintf(os.Stderr, "Loading configuration from: %s\n", configFile)
	}
	cfg, err := config.LoadWithVerbose(configFile, verbose)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "Configuration loaded successfully\n")
		fmt.Fprintf(os.Stderr, "Available color schemes: ")
		schemes := make([]string, 0, len(cfg.ColorSchemes))
		for name := range cfg.ColorSchemes {
			schemes = append(schemes, name)
		}
		fmt.Fprintf(os.Stderr, "%s\n", strings.Join(schemes, ", "))
	}

	// Get color scheme
	scheme, exists := cfg.ColorSchemes[colorScheme]
	if !exists {
		return fmt.Errorf("color scheme '%s' not found", colorScheme)
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "Using color scheme: %s\n", colorScheme)
	}

	colorizer := colorer.New(scheme)

	// If no files specified, read from stdin
	if len(args) == 0 {
		if verbose {
			fmt.Fprintf(os.Stderr, "Reading from standard input\n")
		}
		return processReader(os.Stdin, colorizer, "")
	}

	// Process each file
	if verbose {
		fmt.Fprintf(os.Stderr, "Processing %d file(s)\n", len(args))
	}
	for i, filename := range args {
		if verbose {
			fmt.Fprintf(os.Stderr, "[%d/%d] Processing file: %s\n", i+1, len(args), filename)
		}
		if err := processFile(filename, colorizer); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", filename, err)
			}
			continue
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "[%d/%d] Completed: %s\n", i+1, len(args), filename)
		}
	}

	return nil
}

func processFile(filename string, colorizer *colorer.Colorer) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return processReader(file, colorizer, filename)
}

func processReader(reader io.Reader, colorizer *colorer.Colorer, filename string) error {
	scanner := bufio.NewScanner(reader)

	// Detect format from first few lines if filename is provided
	var format parser.Format
	var lines []string

	// Read first few lines to detect format
	for i := 0; i < 10 && scanner.Scan(); i++ {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	// Detect format
	if filename != "" {
		format = parser.DetectFormatFromFilename(filename)
		if verbose {
			if format != parser.FormatUnknown {
				fmt.Fprintf(os.Stderr, "Format detected from filename: %s\n", formatToString(format))
			} else {
				fmt.Fprintf(os.Stderr, "Could not detect format from filename, analyzing content...\n")
			}
		}
	}
	if format == parser.FormatUnknown && len(lines) > 0 {
		format = parser.DetectFormatFromContent(lines)
	}
	if format == parser.FormatUnknown {
		return fmt.Errorf("could not detect format from content")
	} else if verbose {
		fmt.Fprintf(os.Stderr, "Format detected from content: %s\n", formatToString(format))
	}

	lineCount := 0
	sequenceCount := 0

	// Process the buffered lines first
	for _, line := range lines {
		processLine(line, format, colorizer)
		lineCount++
		if isSequenceCountableLine(line, format) {
			sequenceCount++
		}
	}

	// Continue processing the rest of the input
	for scanner.Scan() {
		processLine(scanner.Text(), format, colorizer)
		lineCount++
		if isSequenceCountableLine(scanner.Text(), format) {
			sequenceCount++
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Processed %d lines, %d sequences\n", lineCount, sequenceCount)
	}

	return scanner.Err()
}

func processLine(line string, format parser.Format, colorizer *colorer.Colorer) {
	switch format {
	case parser.FormatFASTA:
		if strings.HasPrefix(line, ">") {
			// Header line - print as is
			fmt.Println(line)
		} else {
			// Sequence line - colorize
			fmt.Println(colorizer.ColorizeSequence(line))
		}
	case parser.FormatFASTQ:
		if strings.HasPrefix(line, "@") || strings.HasPrefix(line, "+") {
			// Header lines - print as is
			fmt.Println(line)
		} else if parser.IsSequenceLine(line) {
			// Sequence line - colorize
			fmt.Println(colorizer.ColorizeSequence(line))
		} else if parser.IsQualityLine(line) {
			// Quality line - colorize
			fmt.Println(colorizer.ColorizeQuality(line))
		} else {
			fmt.Println(line)
		}
	case parser.FormatSAM:
		if strings.HasPrefix(line, "@") {
			// Header line - print as is
			fmt.Println(line)
		} else {
			// Data line - colorize sequence column
			fmt.Println(colorizer.ColorizeSAM(line))
		}
	case parser.FormatVCF:
		if strings.HasPrefix(line, "#") {
			// Header line - print as is
			fmt.Println(line)
		} else {
			// Data line - colorize relevant columns
			fmt.Println(colorizer.ColorizeVCF(line))
		}
	default:
		// Unknown format - just print as is
		fmt.Println(line)
	}
}

// formatToString converts a parser.Format to a human-readable string
func formatToString(format parser.Format) string {
	switch format {
	case parser.FormatFASTA:
		return "FASTA"
	case parser.FormatFASTQ:
		return "FASTQ"
	case parser.FormatSAM:
		return "SAM"
	case parser.FormatVCF:
		return "VCF"
	default:
		return "Unknown"
	}
}

// isSequenceCountableLine determines if a line represents a sequence entry for counting purposes
func isSequenceCountableLine(line string, format parser.Format) bool {
	switch format {
	case parser.FormatFASTA:
		return strings.HasPrefix(line, ">")
	case parser.FormatFASTQ:
		return strings.HasPrefix(line, "@") && !strings.HasPrefix(line, "+")
	case parser.FormatSAM:
		return !strings.HasPrefix(line, "@") && strings.Count(line, "\t") >= 10
	case parser.FormatVCF:
		return !strings.HasPrefix(line, "#")
	default:
		return false
	}
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	homeDir, _ := os.UserHomeDir()
	defaultConfig := filepath.Join(homeDir, ".colordna.yaml")

	rootCmd.PersistentFlags().StringVar(&configFile, "config", defaultConfig, "config file")
	rootCmd.PersistentFlags().StringVarP(&colorScheme, "scheme", "s", "bright", "color scheme to use")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
