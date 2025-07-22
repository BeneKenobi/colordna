# colordna

A Go command-line tool that colorizes DNA/RNA sequences and quality scores for better visualization in the terminal.

## Features

- **Multiple file format support**: FASTA, FASTQ, SAM, VCF with automatic format detection
- **Custom color schemes**: Define your own color schemes or use built-in ones
- **Quality score coloring**: Gradient coloring for Phred quality scores in FASTQ/SAM files
- **Pipe support**: Works with standard input for streaming data processing

## Installation

### From Source

```bash
git clone https://github.com/benekenobi/colordna.git
cd colordna
go build -o colordna
sudo mv colordna /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/benekenobi/colordna@latest
```

### Development

```bash
# Clone the repository
git clone https://github.com/benekenobi/colordna.git
cd colordna

# Download dependencies
go mod download

# Build
go build -o colordna

# Run tests
go test ./...

# Format code
go fmt ./...

# Vet code
go vet ./...

# Install locally
go install
```

## Usage

### Basic Usage

```bash
# Colorize a FASTA file
colordna sequences.fasta

# Colorize a FASTQ file with quality scores
colordna reads.fastq

# Use a specific color scheme
colordna --scheme classic sequences.fasta

# Process multiple files
colordna file1.fasta file2.fastq file3.sam

# Use with pipes
cat sequences.fasta | colordna
samtools view alignment.bam | colordna
```

### Command Options

```
  -s, --scheme string   Color scheme to use (default "bright")
      --config string   Config file (default "~/.colordna.yaml")  
  -v, --verbose         Verbose output
  -h, --help           Show help
```

## Color Schemes

### Built-in Schemes

1. **bright** (default): Font-only coloring with bright colors, no background
   - A: Bright red
   - T: Bright green
   - G: Bright yellow
   - C: Bright blue
   - U: Bright magenta
   - N: Dark gray

2. **classic**: Background coloring similar to original dnacol
   - Colored backgrounds with contrasting text

3. **pastel**: Light background colors with black text

4. **monochrome**: Text styling only (bold, italic, underline, etc.)

### Custom Color Schemes

Edit `~/.colordna.yaml` to add custom color schemes:

```yaml
color_schemes:
  my_custom:
    a: "\033[31m"      # Red
    t: "\033[32m"      # Green  
    g: "\033[33m"      # Yellow
    c: "\033[34m"      # Blue
    u: "\033[35m"      # Magenta
    "n": "\033[90m"    # Dark gray
    quality: gradient   # gradient, mono, or none
    background: false   # Use background colors
```

### Color Codes

Use ANSI escape sequences for colors:

#### Font Colors
- `\033[31m` - Red
- `\033[32m` - Green
- `\033[33m` - Yellow
- `\033[34m` - Blue
- `\033[35m` - Magenta
- `\033[36m` - Cyan
- `\033[91m` - Bright red
- `\033[92m` - Bright green
- etc.

#### Background Colors
- `\033[41m` - Red background
- `\033[42m` - Green background
- etc.

#### Text Styles
- `\033[1m` - Bold
- `\033[2m` - Dim
- `\033[3m` - Italic
- `\033[4m` - Underline

## Acknowledgments

Inspired by the [dnacol](https://github.com/koelling/dnacol) Python project by [koelling](https://github.com/koelling).
