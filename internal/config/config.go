package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	BasePaths []string
}

// ConfigError represents configuration-related errors
type ConfigError struct {
	Field   string
	Message string
	Cause   error
}

func (e *ConfigError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("config error in %s: %s (caused by: %v)", e.Field, e.Message, e.Cause)
	}
	return fmt.Sprintf("config error in %s: %s", e.Field, e.Message)
}

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig() (*Config, error) {
	// Try to load .env file (ignore error if file doesn't exist)
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("Warning: Could not load config.env file: %v", err)
	}

	basePaths, err := getBasePaths()
	if err != nil {
		return nil, &ConfigError{
			Field:   "BasePaths",
			Message: "failed to parse base paths",
			Cause:   err,
		}
	}

	config := &Config{
		BasePaths: basePaths,
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.BasePaths) == 0 {
		return &ConfigError{
			Field:   "BasePaths",
			Message: "no base paths configured",
		}
	}

	validCount := 0
	for _, path := range c.BasePaths {
		if c.IsValidBasePath(path) {
			validCount++
		}
	}

	if validCount == 0 {
		return &ConfigError{
			Field:   "BasePaths",
			Message: "no valid base paths found",
		}
	}

	return nil
}

// getBasePaths returns the base paths from environment variable or current directory
func getBasePaths() ([]string, error) {
	basePathsStr := os.Getenv("JSON_MANAGER_BASE_PATHS")
	if basePathsStr == "" {
		// Fallback to current working directory
		if cwd, err := os.Getwd(); err == nil {
			log.Printf("No base paths configured, using current directory: %s", cwd)
			return []string{cwd}, nil
		}
		// Ultimate fallback
		log.Printf("Warning: Using relative path '.' as fallback")
		return []string{"."}, nil
	}

	// Split by semicolon or comma to support multiple paths
	var paths []string
	separators := []string{";", ","}

	for _, sep := range separators {
		if strings.Contains(basePathsStr, sep) {
			paths = strings.Split(basePathsStr, sep)
			break
		}
	}

	// If no separator found, treat as single path
	if len(paths) == 0 {
		paths = []string{basePathsStr}
	}

	// Clean and validate paths
	var cleanPaths []string
	var errors []string

	for i, path := range paths {
		cleanPath := strings.TrimSpace(path)
		if cleanPath == "" {
			errors = append(errors, fmt.Sprintf("path %d is empty", i+1))
			continue
		}

		// Convert to absolute path if possible
		if absPath, err := filepath.Abs(cleanPath); err == nil {
			cleanPaths = append(cleanPaths, absPath)
		} else {
			log.Printf("Warning: Could not convert path to absolute: %s (error: %v)", cleanPath, err)
			cleanPaths = append(cleanPaths, cleanPath)
		}
	}

	if len(errors) > 0 {
		return cleanPaths, fmt.Errorf("path validation errors: %s", strings.Join(errors, ", "))
	}

	if len(cleanPaths) == 0 {
		return nil, fmt.Errorf("no valid paths found after processing")
	}

	log.Printf("Loaded %d base paths: %v", len(cleanPaths), cleanPaths)
	return cleanPaths, nil
}

// GetBasePaths returns all configured base paths
func (c *Config) GetBasePaths() []string {
	return c.BasePaths
}

// GetFirstBasePath returns the first base path (for backwards compatibility)
func (c *Config) GetFirstBasePath() string {
	if len(c.BasePaths) > 0 {
		return c.BasePaths[0]
	}
	return "."
}

// ResolveFullPath resolves a relative path against the first base path
func (c *Config) ResolveFullPath(relativePath string) string {
	if filepath.IsAbs(relativePath) {
		return relativePath
	}

	firstBasePath := c.GetFirstBasePath()
	return filepath.Join(firstBasePath, relativePath)
}

// ResolveFullPathInBasePath resolves a relative path against a specific base path
func (c *Config) ResolveFullPathInBasePath(relativePath, basePath string) string {
	if filepath.IsAbs(relativePath) {
		return relativePath
	}
	return filepath.Join(basePath, relativePath)
}

// IsValidPath checks if a path exists and is accessible
func (c *Config) IsValidPath(path string) bool {
	fullPath := c.ResolveFullPath(path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// IsValidBasePath checks if a base path exists and is accessible
func (c *Config) IsValidBasePath(basePath string) bool {
	_, err := os.Stat(basePath)
	return err == nil
}

// GetValidBasePaths returns only the base paths that exist and are accessible
func (c *Config) GetValidBasePaths() []string {
	var validPaths []string
	for _, path := range c.BasePaths {
		if c.IsValidBasePath(path) {
			validPaths = append(validPaths, path)
		}
	}
	return validPaths
}

// GetConfigDir returns the directory containing configuration files
func GetConfigDir() string {
	if configDir := os.Getenv("CONFIG_DIR"); configDir != "" {
		return configDir
	}

	// Default to current directory
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}

	return "."
}
