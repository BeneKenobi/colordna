package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ColorScheme represents a color scheme configuration
type ColorScheme struct {
	A          string `yaml:"a"`          // Adenine
	T          string `yaml:"t"`          // Thymine
	G          string `yaml:"g"`          // Guanine
	C          string `yaml:"c"`          // Cytosine
	U          string `yaml:"u"`          // Uracil (RNA)
	N          string `yaml:"n"`          // Unknown/ambiguous nucleotide
	Quality    string `yaml:"quality"`    // Quality score color scheme
	Background bool   `yaml:"background"` // Whether to use background colors
}

// Config represents the application configuration
type Config struct {
	ColorSchemes map[string]ColorScheme `yaml:"color_schemes"`
}

// Default color schemes
var defaultConfig = Config{
	ColorSchemes: map[string]ColorScheme{
		"bright": {
			A:          "\033[91m", // Bright red
			T:          "\033[92m", // Bright green
			G:          "\033[93m", // Bright yellow
			C:          "\033[94m", // Bright blue
			U:          "\033[95m", // Bright magenta
			N:          "\033[90m", // Dark gray
			Quality:    "gradient", // Use gradient for quality scores
			Background: false,      // Font colors only (new default)
		},
		"classic": {
			A:          "\033[41m\033[97m",  // Red background, white text
			T:          "\033[42m\033[30m",  // Green background, black text
			G:          "\033[43m\033[30m",  // Yellow background, black text
			C:          "\033[44m\033[97m",  // Blue background, white text
			U:          "\033[45m\033[97m",  // Magenta background, white text
			N:          "\033[100m\033[97m", // Dark gray background, white text
			Quality:    "gradient",
			Background: true,
		},
		"pastel": {
			A:          "\033[101m\033[30m", // Light red background, black text
			T:          "\033[102m\033[30m", // Light green background, black text
			G:          "\033[103m\033[30m", // Light yellow background, black text
			C:          "\033[104m\033[30m", // Light blue background, black text
			U:          "\033[105m\033[30m", // Light magenta background, black text
			N:          "\033[47m\033[30m",  // Light gray background, black text
			Quality:    "gradient",
			Background: true,
		},
		"monochrome": {
			A:          "\033[1m",  // Bold
			T:          "\033[4m",  // Underline
			G:          "\033[3m",  // Italic
			C:          "\033[2m",  // Dim
			U:          "\033[9m",  // Strikethrough
			N:          "\033[90m", // Dark gray
			Quality:    "mono",
			Background: false,
		},
	},
}

// Load loads the configuration from the specified file, or returns default config if file doesn't exist
func Load(configPath string) (*Config, error) {
	return LoadWithVerbose(configPath, false)
}

// LoadWithVerbose loads the configuration with optional verbose output
func LoadWithVerbose(configPath string, verbose bool) (*Config, error) {
	// If config file doesn't exist, create it with default configuration
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if verbose {
			fmt.Fprintf(os.Stderr, "Config file not found, creating default config at: %s\n", configPath)
		}
		if err := createDefaultConfig(configPath); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Could not create config file, using built-in defaults\n")
			}
			// If we can't create the config file, just return the default config
			return &defaultConfig, nil
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "Default config file created successfully\n")
		}
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Could not read config file, using built-in defaults: %v\n", err)
		}
		return &defaultConfig, nil // Fallback to default on error
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Config file parsed successfully\n")
		fmt.Fprintf(os.Stderr, "Found %d color scheme(s) in config file\n", len(config.ColorSchemes))
	}

	// Merge with defaults for any missing schemes
	mergedSchemes := 0
	for name, scheme := range defaultConfig.ColorSchemes {
		if _, exists := config.ColorSchemes[name]; !exists {
			if config.ColorSchemes == nil {
				config.ColorSchemes = make(map[string]ColorScheme)
			}
			config.ColorSchemes[name] = scheme
			mergedSchemes++
		}
	}

	if verbose && mergedSchemes > 0 {
		fmt.Fprintf(os.Stderr, "Merged %d default color scheme(s)\n", mergedSchemes)
	}

	return &config, nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal default config to YAML
	data, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	// Add header comment
	header := `# colordna Configuration File
# This file contains color schemes for DNA/RNA sequence visualization
# 
# Color format: ANSI escape sequences
# - Font colors: \033[91m (bright red), \033[92m (bright green), etc.
# - Background colors: \033[41m\033[97m (red background + white text)
# - Styles: \033[1m (bold), \033[4m (underline), \033[3m (italic)
#
# You can create custom color schemes by adding new entries under color_schemes.
# The 'bright' scheme is the default and uses only font colors (no backgrounds).

`

	content := header + string(data)

	// Write to file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
