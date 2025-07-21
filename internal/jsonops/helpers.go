package jsonops

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSONParser provides a more robust way to handle JSON operations
type JSONParser struct {
	data map[string]interface{}
}

// NewJSONParser creates a parser from JSON string
func NewJSONParser(jsonStr string) (*JSONParser, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}
	return &JSONParser{data: data}, nil
}

// AddKeyAtPath adds a key-value pair at the specified path
func (jp *JSONParser) AddKeyAtPath(path, key string, value interface{}) error {
	if path == "" {
		// Add at root level
		if _, exists := jp.data[key]; exists {
			return fmt.Errorf("key '%s' already exists at root level", key)
		}
		jp.data[key] = value
		return nil
	}

	// Navigate to the target path and add the key
	return jp.addToPath(jp.data, strings.Split(path, "."), key, value)
}

// addToPath recursively navigates to the target path
func (jp *JSONParser) addToPath(current interface{}, pathParts []string, key string, value interface{}) error {
	if len(pathParts) == 0 {
		// We've reached the target, add the key
		if obj, ok := current.(map[string]interface{}); ok {
			if _, exists := obj[key]; exists {
				return fmt.Errorf("key '%s' already exists", key)
			}
			obj[key] = value
			return nil
		}
		return fmt.Errorf("target is not an object")
	}

	currentKey := pathParts[0]
	remaining := pathParts[1:]

	if obj, ok := current.(map[string]interface{}); ok {
		if next, exists := obj[currentKey]; exists {
			return jp.addToPath(next, remaining, key, value)
		}
		return fmt.Errorf("path not found: %s", currentKey)
	}

	return fmt.Errorf("current element is not an object")
}

// ToIndentedJSON converts back to formatted JSON string
func (jp *JSONParser) ToIndentedJSON() (string, error) {
	result, err := json.MarshalIndent(jp.data, "", "  ")
	return string(result), err
}

// keyExistsInObject checks if a key exists in an object at a specific depth level
func keyExistsInObject(lines []string, key string, startLine, targetDepth int) bool {
	currentDepth := 0

	for i := startLine; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Count braces to track depth
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")
		currentDepth += openBraces - closeBraces

		// If we're at the target depth and find the key
		if currentDepth == targetDepth && strings.Contains(line, `"`+key+`"`) && strings.Contains(line, ":") {
			return true
		}

		// If we've gone back to a shallower depth, stop searching
		if currentDepth < targetDepth {
			break
		}
	}

	return false
}

// keyExistsInArrayObjects checks if a key exists in any object within an array
func keyExistsInArrayObjects(lines []string, key string, arrayStartLine int) bool {
	bracketDepth := 0
	braceDepth := 0
	inArray := false

	for i := arrayStartLine; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Count brackets and braces
		openBrackets := strings.Count(line, "[")
		closeBrackets := strings.Count(line, "]")
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")

		bracketDepth += openBrackets - closeBrackets
		braceDepth += openBraces - closeBraces

		// We're in the array when bracket depth > 0
		if bracketDepth > 0 {
			inArray = true
		}

		// If we're in an object within the array and find the key
		if inArray && braceDepth > 0 && strings.Contains(line, `"`+key+`"`) && strings.Contains(line, ":") {
			return true
		}

		// If we've closed the array, stop searching
		if inArray && bracketDepth == 0 {
			break
		}
	}

	return false
}

// getIndentation extracts the indentation (spaces/tabs) from a line
func getIndentation(line string) string {
	indentation := ""
	for _, char := range line {
		if char == ' ' || char == '\t' {
			indentation += string(char)
		} else {
			break
		}
	}
	return indentation
}

// detectIndentationForObject detects the indentation pattern for objects
func detectIndentationForObject(lines []string, objectStartLine int) string {
	// Look for the first property line after the opening brace
	for i := objectStartLine + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// If we find a property line (contains key-value pair)
		if strings.Contains(trimmed, ":") && strings.Contains(trimmed, `"`) {
			return getIndentation(line)
		}

		// If we hit a closing brace, stop looking
		if strings.Contains(trimmed, "}") {
			break
		}
	}

	// Default to 2 spaces if we can't detect
	baseIndent := getIndentation(lines[objectStartLine])
	return baseIndent + "  "
}

// detectIndentationForArray detects the indentation pattern for arrays
func detectIndentationForArray(lines []string, arrayStartLine int) string {
	// Look for the first element line after the opening bracket
	for i := arrayStartLine + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// If we find any content (object, array, or value)
		if trimmed != "" && !strings.HasPrefix(trimmed, "]") {
			return getIndentation(line)
		}

		// If we hit a closing bracket, stop looking
		if strings.Contains(trimmed, "]") {
			break
		}
	}

	// Default to 2 spaces if we can't detect
	baseIndent := getIndentation(lines[arrayStartLine])
	return baseIndent + "  "
}

// findEndOfValue finds the line index where a JSON value ends
func findEndOfValue(lines []string, startIndex int) int {
	startLine := strings.TrimSpace(lines[startIndex])

	// If the value is on the same line (simple value), return the same index
	if strings.Contains(startLine, ":") {
		colonIndex := strings.Index(startLine, ":")
		valueStart := strings.TrimSpace(startLine[colonIndex+1:])

		// Simple values (string, number, boolean, null)
		if strings.HasPrefix(valueStart, `"`) ||
			strings.HasPrefix(valueStart, "true") ||
			strings.HasPrefix(valueStart, "false") ||
			strings.HasPrefix(valueStart, "null") ||
			(len(valueStart) > 0 && (valueStart[0] >= '0' && valueStart[0] <= '9')) {
			return startIndex
		}

		// Array or object - find the matching closing bracket/brace
		if strings.HasPrefix(valueStart, "[") {
			return findMatchingBracket(lines, startIndex, '[', ']')
		}
		if strings.HasPrefix(valueStart, "{") {
			return findMatchingBracket(lines, startIndex, '{', '}')
		}
	}

	return startIndex
}

// findMatchingBracket finds the line with the matching closing bracket/brace
func findMatchingBracket(lines []string, startIndex int, openChar, closeChar rune) int {
	depth := 0
	inString := false
	escaped := false

	for i := startIndex; i < len(lines); i++ {
		line := lines[i]
		for _, char := range line {
			if escaped {
				escaped = false
				continue
			}

			if char == '\\' && inString {
				escaped = true
				continue
			}

			if char == '"' {
				inString = !inString
				continue
			}

			if !inString {
				if char == openChar {
					depth++
				} else if char == closeChar {
					depth--
					if depth == 0 {
						return i
					}
				}
			}
		}
	}

	return -1
}

// isArrayOfObjects determines if an array contains objects (vs simple values)
func isArrayOfObjects(lines []string, arrayStartLine int) bool {
	bracketDepth := 0

	for i := arrayStartLine; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Count brackets
		openBrackets := strings.Count(line, "[")
		closeBrackets := strings.Count(line, "]")
		bracketDepth += openBrackets - closeBrackets

		// Look for opening brace indicating an object
		if bracketDepth > 0 && strings.Contains(line, "{") {
			return true
		}

		// If we've closed the array without finding objects, it's not an array of objects
		if bracketDepth == 0 && i > arrayStartLine {
			break
		}
	}

	return false
}

// validateJSON validates if a string is valid JSON
func validateJSON(jsonStr string) error {
	var temp interface{}
	return json.Unmarshal([]byte(jsonStr), &temp)
}
