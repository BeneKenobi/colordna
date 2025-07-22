package colorer

import (
	"fmt"
	"strings"

	"github.com/benekenobi/colordna/internal/config"
	"github.com/benekenobi/colordna/internal/parser"
)

const (
	resetCode = "\033[0m"
)

// Colorer handles the coloring of sequences and quality scores
type Colorer struct {
	scheme config.ColorScheme
}

// New creates a new Colorer with the given color scheme
func New(scheme config.ColorScheme) *Colorer {
	return &Colorer{scheme: scheme}
}

// ColorizeSequence colorizes a DNA/RNA/protein sequence
func (c *Colorer) ColorizeSequence(sequence string) string {
	if len(sequence) == 0 {
		return sequence
	}

	var result strings.Builder
	result.Grow(len(sequence) * 10) // Pre-allocate for efficiency

	for _, char := range strings.ToUpper(sequence) {
		color := c.getColorForNucleotide(char)
		if color != "" {
			result.WriteString(color)
			result.WriteRune(char)
			result.WriteString(resetCode)
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// ColorizeQuality colorizes quality scores in FASTQ format
func (c *Colorer) ColorizeQuality(quality string) string {
	if len(quality) == 0 {
		return quality
	}

	if c.scheme.Quality == "gradient" {
		return c.colorizeQualityGradient(quality)
	} else if c.scheme.Quality == "mono" {
		return c.colorizeQualityMono(quality)
	}

	// Default: no coloring
	return quality
}

// ColorizeSAM colorizes the sequence column in SAM format
func (c *Colorer) ColorizeSAM(line string) string {
	fields := strings.Split(line, "\t")
	if len(fields) < 11 {
		return line // Not enough fields for SAM format
	}

	// Field 9 (index 9) contains the sequence
	sequence := fields[9]
	if sequence != "*" && parser.IsSequenceLine(sequence) {
		fields[9] = c.ColorizeSequence(sequence)
	}

	// Field 10 (index 10) contains the quality scores
	if len(fields) > 10 {
		quality := fields[10]
		if quality != "*" && parser.IsQualityLine(quality) {
			fields[10] = c.ColorizeQuality(quality)
		}
	}

	return strings.Join(fields, "\t")
}

// ColorizeVCF colorizes relevant fields in VCF format
func (c *Colorer) ColorizeVCF(line string) string {
	fields := strings.Split(line, "\t")
	if len(fields) < 5 {
		return line // Not enough fields for VCF format
	}

	// Field 3 (index 3) contains the reference allele
	if ref := fields[3]; ref != "." && parser.IsSequenceLine(ref) {
		fields[3] = c.ColorizeSequence(ref)
	}

	// Field 4 (index 4) contains the alternative allele(s)
	if alt := fields[4]; alt != "." {
		alternatives := strings.Split(alt, ",")
		for i, allele := range alternatives {
			if parser.IsSequenceLine(allele) {
				alternatives[i] = c.ColorizeSequence(allele)
			}
		}
		fields[4] = strings.Join(alternatives, ",")
	}

	return strings.Join(fields, "\t")
}

// getColorForNucleotide returns the ANSI color code for a nucleotide
func (c *Colorer) getColorForNucleotide(nucleotide rune) string {
	switch nucleotide {
	case 'A':
		return c.scheme.A
	case 'T':
		return c.scheme.T
	case 'G':
		return c.scheme.G
	case 'C':
		return c.scheme.C
	case 'U':
		return c.scheme.U
	case 'N':
		return c.scheme.N
	default:
		// For other characters (like ambiguous nucleotides), use N color or no color
		if c.scheme.N != "" {
			return c.scheme.N
		}
		return ""
	}
}

// colorizeQualityGradient applies a gradient color to quality scores
func (c *Colorer) colorizeQualityGradient(quality string) string {
	var result strings.Builder
	result.Grow(len(quality) * 10)

	for _, char := range quality {
		// Convert quality character to Phred score
		phred := int(char) - 33 // Standard Phred+33 encoding

		color := c.getQualityColor(phred)
		if color != "" {
			result.WriteString(color)
			result.WriteRune(char)
			result.WriteString(resetCode)
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// colorizeQualityMono applies monochrome styling to quality scores
func (c *Colorer) colorizeQualityMono(quality string) string {
	var result strings.Builder
	result.Grow(len(quality) * 10)

	for _, char := range quality {
		phred := int(char) - 33

		var style string
		if phred >= 30 {
			style = "\033[1m" // Bold for high quality
		} else if phred >= 20 {
			style = "\033[0m" // Normal for medium quality
		} else {
			style = "\033[2m" // Dim for low quality
		}

		result.WriteString(style)
		result.WriteRune(char)
		result.WriteString(resetCode)
	}

	return result.String()
}

// getQualityColor returns color based on Phred quality score
func (c *Colorer) getQualityColor(phred int) string {
	// Quality color gradient from red (low) to green (high)
	switch {
	case phred >= 40:
		return "\033[92m" // Bright green - excellent quality
	case phred >= 30:
		return "\033[32m" // Green - good quality
	case phred >= 20:
		return "\033[93m" // Yellow - acceptable quality
	case phred >= 10:
		return "\033[91m" // Red - poor quality
	default:
		return "\033[31m" // Dark red - very poor quality
	}
}

// ColorizeText colorizes any text that contains DNA/RNA sequences
func (c *Colorer) ColorizeText(text string) string {
	// This method can be used to colorize sequences found within arbitrary text
	words := strings.Fields(text)
	for i, word := range words {
		// Remove common punctuation for sequence detection
		cleanWord := strings.Trim(word, ".,;:!?()[]{}\"'")
		if len(cleanWord) >= 3 && parser.IsSequenceLine(cleanWord) {
			// Replace the sequence part while preserving punctuation
			prefix := word[:strings.Index(word, cleanWord)]
			suffix := word[strings.Index(word, cleanWord)+len(cleanWord):]
			words[i] = prefix + c.ColorizeSequence(cleanWord) + suffix
		}
	}
	return strings.Join(words, " ")
}

// HasColor checks if the colorer has any colors configured
func (c *Colorer) HasColor() bool {
	return c.scheme.A != "" || c.scheme.T != "" || c.scheme.G != "" ||
		c.scheme.C != "" || c.scheme.U != "" || c.scheme.N != ""
}

// GetScheme returns the current color scheme
func (c *Colorer) GetScheme() config.ColorScheme {
	return c.scheme
}

// Preview returns a preview of the color scheme
func (c *Colorer) Preview() string {
	sample := "ATGCUN"
	result := fmt.Sprintf("DNA/RNA: %s\n", c.ColorizeSequence(sample))

	if c.scheme.Quality == "gradient" {
		qualitySample := "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJ"
		result += fmt.Sprintf("Quality: %s\n", c.ColorizeQuality(qualitySample))
	}

	return result
}
