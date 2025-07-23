package jsonops

import (
	"fmt"
	"regexp"
	"strings"

	"goldenMagic/internal/fileops"
)

// ReplaceKeyRequest represents a request to replace keys in JSON files
type ReplaceKeyRequest struct {
	OldKey        string   `json:"oldKey"`
	NewKey        string   `json:"newKey"`
	SelectedFiles []string `json:"selectedFiles"`
}

// ReplaceKeyResult represents the result of a key replacement operation
type ReplaceKeyResult struct {
	FilePath         string `json:"filePath"`
	Success          bool   `json:"success"`
	Error            string `json:"error,omitempty"`
	ReplacementCount int    `json:"replacementCount"`
	ModifiedContent  string `json:"modifiedContent"`
}

// ReplaceKeyInFiles replaces old keys with new keys in selected files using string replacement
func ReplaceKeyInFiles(request ReplaceKeyRequest) ([]ReplaceKeyResult, error) {
	if request.OldKey == "" {
		return nil, fmt.Errorf("old key cannot be empty")
	}

	if request.NewKey == "" {
		return nil, fmt.Errorf("new key cannot be empty")
	}

	if request.OldKey == request.NewKey {
		return nil, fmt.Errorf("old key and new key cannot be the same")
	}

	var results []ReplaceKeyResult

	for _, filePath := range request.SelectedFiles {
		result := ReplaceKeyResult{
			FilePath: filePath,
			Success:  false,
		}

		// Read the file content
		content, err := fileops.ReadFile(filePath)
		if err != nil {
			result.Error = fmt.Sprintf("failed to read file: %v", err)
			results = append(results, result)
			continue
		}

		// Perform the key replacement using string replacement
		modifiedContent, replacementCount := replaceKeysInText(string(content), request.OldKey, request.NewKey)

		if replacementCount == 0 {
			result.Error = fmt.Sprintf("no keys found with name '%s'", request.OldKey)
			results = append(results, result)
			continue
		}

		// Write the modified content back to the file
		if err := fileops.WriteFile(filePath, []byte(modifiedContent)); err != nil {
			result.Error = fmt.Sprintf("failed to write file: %v", err)
			results = append(results, result)
			continue
		}

		result.Success = true
		result.ReplacementCount = replacementCount
		result.ModifiedContent = modifiedContent
		results = append(results, result)
	}

	return results, nil
}

// replaceKeysInText replaces JSON keys in text using regex pattern matching
func replaceKeysInText(content, oldKey, newKey string) (string, int) {
	// Create a regex pattern to match JSON keys
	// This pattern matches: "oldKey" followed by optional whitespace and a colon
	pattern := fmt.Sprintf(`"(%s)"\s*:`, regexp.QuoteMeta(oldKey))
	regex := regexp.MustCompile(pattern)

	// Replace all occurrences
	replacementCount := 0
	result := regex.ReplaceAllStringFunc(content, func(match string) string {
		replacementCount++
		// Replace the old key with the new key while preserving the formatting
		return strings.Replace(match, fmt.Sprintf(`"%s"`, oldKey), fmt.Sprintf(`"%s"`, newKey), 1)
	})

	return result, replacementCount
}
