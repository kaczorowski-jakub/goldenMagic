package fileops

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MaxFileSize defines the maximum file size to process (10MB)
const MaxFileSize = 10 * 1024 * 1024

// JSONFile represents a JSON file with its metadata
type JSONFile struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	BasePath string `json:"basePath"` // Which base path this file belongs to
	Size     int64  `json:"size"`     // File size in bytes
}

// GetJSONFileContent returns the content of a JSON file with size validation
func GetJSONFileContent(filePath string) (string, error) {
	// Check file size first
	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("error getting file info: %v", err)
	}

	if info.Size() > MaxFileSize {
		return "", fmt.Errorf("file too large (%d bytes, max %d bytes)", info.Size(), MaxFileSize)
	}

	content, err := ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	// Validate JSON format
	if err := validateJSON(string(content)); err != nil {
		return "", fmt.Errorf("invalid JSON content: %v", err)
	}

	return string(content), nil
}

// validateJSON validates if a string is valid JSON
func validateJSON(jsonStr string) error {
	var temp interface{}
	return json.Unmarshal([]byte(jsonStr), &temp)
}

// ReadFile reads a file with better error handling
func ReadFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Use io.ReadAll for better memory management
	return io.ReadAll(file)
}

// WriteFile writes data to a file with atomic operations
func WriteFile(filePath string, data []byte) error {
	// Write to temporary file first
	tempFile := filePath + ".tmp"

	file, err := os.Create(tempFile)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	closeErr := file.Close()

	if err != nil {
		os.Remove(tempFile) // Clean up temp file
		return err
	}

	if closeErr != nil {
		os.Remove(tempFile) // Clean up temp file
		return closeErr
	}

	// Atomically replace the original file
	return os.Rename(tempFile, filePath)
}

// ContainsKeyDeep recursively searches for a key in JSON content
func ContainsKeyDeep(content []byte, searchKey string) bool {
	var data any
	if err := json.Unmarshal(content, &data); err != nil {
		return false
	}

	return containsKeyRecursive(data, searchKey)
}

// containsKeyRecursive recursively searches for a key in any JSON structure
func containsKeyRecursive(data any, searchKey string) bool {
	switch v := data.(type) {
	case map[string]any:
		// Check if the key exists directly
		if _, exists := v[searchKey]; exists {
			return true
		}
		// Recursively check nested objects
		for _, value := range v {
			if containsKeyRecursive(value, searchKey) {
				return true
			}
		}
	case []any:
		// Recursively check array elements
		for _, item := range v {
			if containsKeyRecursive(item, searchKey) {
				return true
			}
		}
	}
	return false
}

// BrowseFolders recursively searches for files across multiple base paths
func BrowseFolders(basePaths []string, extensionFilter, jsonKeyFilter string) ([]JSONFile, error) {
	var allFiles []JSONFile

	for _, basePath := range basePaths {
		files, err := BrowseFolder(basePath, extensionFilter, jsonKeyFilter)
		if err != nil {
			// Log the error but continue with other paths
			continue
		}
		allFiles = append(allFiles, files...)
	}

	return allFiles, nil
}

// BrowseFolder recursively searches for files matching the extension filter and JSON key filter
func BrowseFolder(folderPath, extensionFilter, jsonKeyFilter string) ([]JSONFile, error) {
	var files []JSONFile

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Apply extension filter
		if extensionFilter != "" {
			// Remove the * if present
			filter := strings.TrimPrefix(extensionFilter, "*")
			if !strings.HasSuffix(strings.ToLower(info.Name()), strings.ToLower(filter)) {
				return nil
			}
		}

		// Apply JSON key filter (only for JSON-like files)
		if jsonKeyFilter != "" {
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				// Skip files we can't read
				return nil
			}

			// Check if file contains the specified JSON key
			if !ContainsKeyDeep(content, jsonKeyFilter) {
				return nil
			}
		}

		files = append(files, JSONFile{
			Name:     info.Name(),
			Path:     path,
			BasePath: folderPath,
			Size:     info.Size(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return files, nil
}

// GroupFilesByBasePath groups files by their base path
func GroupFilesByBasePath(files []JSONFile) map[string][]JSONFile {
	grouped := make(map[string][]JSONFile)

	for _, file := range files {
		grouped[file.BasePath] = append(grouped[file.BasePath], file)
	}

	return grouped
}

// GetUniqueBasePaths returns all unique base paths from a list of files
func GetUniqueBasePaths(files []JSONFile) []string {
	basePathMap := make(map[string]bool)

	for _, file := range files {
		basePathMap[file.BasePath] = true
	}

	var basePaths []string
	for basePath := range basePathMap {
		basePaths = append(basePaths, basePath)
	}

	return basePaths
}

// FilterFilesByBasePath filters files to only include those from specific base paths
func FilterFilesByBasePath(files []JSONFile, allowedBasePaths []string) []JSONFile {
	allowedMap := make(map[string]bool)
	for _, path := range allowedBasePaths {
		allowedMap[path] = true
	}

	var filtered []JSONFile
	for _, file := range files {
		if allowedMap[file.BasePath] {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// CopyFile copies a file from source to destination
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}
