package parser

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Format represents a file format
type Format int

const (
	FormatUnknown Format = iota
	FormatFASTA
	FormatFASTQ
	FormatSAM
	FormatVCF
)

var (
	// Regular expressions for sequence detection
	dnaRegex     = regexp.MustCompile(`^[ATGCN]*$`)
	rnaRegex     = regexp.MustCompile(`^[AUGCN]*$`)
	proteinRegex = regexp.MustCompile(`^[ACDEFGHIKLMNPQRSTVWY]*$`)
	qualityRegex = regexp.MustCompile(`^[!-~]*$`) // Printable ASCII characters for quality scores
)

// DetectFormatFromFilename detects file format based on filename extension
func DetectFormatFromFilename(filename string) Format {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".fasta", ".fa", ".fna", ".ffn", ".faa", ".frn":
		return FormatFASTA
	case ".fastq", ".fq":
		return FormatFASTQ
	case ".sam":
		return FormatSAM
	case ".vcf":
		return FormatVCF
	default:
		return FormatUnknown
	}
}

// DetectFormatFromContent detects file format based on content patterns
func DetectFormatFromContent(lines []string) Format {
	if len(lines) == 0 {
		return FormatUnknown
	}

	// Check for VCF
	for _, line := range lines {
		if strings.HasPrefix(line, "##fileformat=VCF") {
			return FormatVCF
		}
	}

	// Check for SAM
	for _, line := range lines {
		if strings.HasPrefix(line, "@HD") || strings.HasPrefix(line, "@SQ") || strings.HasPrefix(line, "@RG") {
			return FormatSAM
		}
		// SAM data line pattern (has multiple tab-separated fields)
		if !strings.HasPrefix(line, "@") && strings.Count(line, "\t") >= 10 {
			return FormatSAM
		}
	}

	// Check for FASTQ vs FASTA
	fastaHeaders := 0
	fastqHeaders := 0

	for i, line := range lines {
		if strings.HasPrefix(line, ">") {
			fastaHeaders++
		} else if strings.HasPrefix(line, "@") && i < len(lines)-3 {
			// FASTQ header should be followed by sequence, +, quality
			if i+3 < len(lines) {
				seqLine := lines[i+1]
				plusLine := lines[i+2]
				qualLine := lines[i+3]

				if strings.HasPrefix(plusLine, "+") &&
					IsSequenceLine(seqLine) &&
					IsQualityLine(qualLine) &&
					len(seqLine) == len(qualLine) {
					fastqHeaders++
				}
			}
		}
	}

	if fastqHeaders > 0 {
		return FormatFASTQ
	} else if fastaHeaders > 0 {
		return FormatFASTA
	}

	return FormatUnknown
}

// IsSequenceLine checks if a line contains a DNA/RNA/protein sequence
func IsSequenceLine(line string) bool {
	if len(line) == 0 {
		return false
	}

	// Convert to uppercase for checking
	upper := strings.ToUpper(line)

	// Check if it's DNA
	if dnaRegex.MatchString(upper) {
		return true
	}

	// Check if it's RNA
	if rnaRegex.MatchString(upper) {
		return true
	}

	// Check if it's protein (longer sequences more likely to be protein)
	if len(upper) > 10 && proteinRegex.MatchString(upper) {
		return true
	}

	return false
}

// IsQualityLine checks if a line contains quality scores (FASTQ format)
func IsQualityLine(line string) bool {
	if len(line) == 0 {
		return false
	}

	// Quality scores should be printable ASCII characters
	return qualityRegex.MatchString(line)
}

// IsDNASequence checks specifically for DNA sequences
func IsDNASequence(sequence string) bool {
	if len(sequence) == 0 {
		return false
	}
	upper := strings.ToUpper(sequence)
	return dnaRegex.MatchString(upper)
}

// IsRNASequence checks specifically for RNA sequences
func IsRNASequence(sequence string) bool {
	if len(sequence) == 0 {
		return false
	}
	upper := strings.ToUpper(sequence)
	return rnaRegex.MatchString(upper)
}

// IsProteinSequence checks specifically for protein sequences
func IsProteinSequence(sequence string) bool {
	if len(sequence) == 0 {
		return false
	}
	upper := strings.ToUpper(sequence)
	return proteinRegex.MatchString(upper)
}
