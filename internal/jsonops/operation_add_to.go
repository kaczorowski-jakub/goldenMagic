package jsonops

import (
	"encoding/json"
	"fmt"
	"strings"
)

// InsertJSONKeyValue inserts a key-value pair into JSON string while preserving structure.
//
// Parameters:
//   - jsonStr: The JSON string to modify
//   - objectPath: The path to the target object (empty string for root level)
//   - key: The key name to insert
//   - value: The value to associate with the key
//
// Returns:
//   - Modified JSON string with the new key-value pair
//   - Error if the operation fails
//
// Example:
//
//	result, err := InsertJSONKeyValue(`{"name": "test"}`, "", "id", 123)
//	// Result: `{"id": 123, "name": "test"}`
func InsertJSONKeyValue(jsonStr, objectPath, key string, value any) (string, error) {
	// Convert value to JSON string
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("error marshaling value: %v", err)
	}

	// Choose insertion method based on object path
	if objectPath == "" {
		return insertAtRoot(jsonStr, key, string(valueJSON))
	} else {
		return insertAtContextPath(jsonStr, objectPath, key, string(valueJSON))
	}
}

// insertAtRoot inserts a key-value pair at the root level of the JSON object.
// This function handles duplicate key checking and proper indentation.
func insertAtRoot(jsonStr, key, valueJSON string) (string, error) {
	lines := strings.Split(jsonStr, "\n")

	// Check if key already exists at root level
	if keyExistsInObject(lines, key, 0, 1) {
		return jsonStr, fmt.Errorf("key '%s' already exists at root level", key)
	}

	// Find the first line with opening brace
	openBraceIndex := -1
	for i, line := range lines {
		if strings.Contains(strings.TrimSpace(line), "{") {
			openBraceIndex = i
			break
		}
	}

	if openBraceIndex == -1 {
		return jsonStr, fmt.Errorf("no opening brace found")
	}

	// Get proper indentation for root level properties
	indent := detectIndentationForObject(lines, openBraceIndex)

	// Create the new property line
	newPropertyLine := indent + `"` + key + `": ` + valueJSON + ","

	// Insert right after the opening brace
	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:openBraceIndex+1]...)
	result = append(result, newPropertyLine)
	result = append(result, lines[openBraceIndex+1:]...)

	return strings.Join(result, "\n"), nil
}

// insertAtContextPath inserts a key-value pair at a specific context path.
// This function determines whether the target is an object or array and delegates accordingly.
func insertAtContextPath(jsonStr, objectPath, key, valueJSON string) (string, error) {
	lines := strings.Split(jsonStr, "\n")

	// Find the target path
	targetLineIndex := -1
	for i, line := range lines {
		if strings.Contains(line, `"`+objectPath+`"`) && strings.Contains(line, ":") {
			targetLineIndex = i
			break
		}
	}

	if targetLineIndex == -1 {
		return jsonStr, fmt.Errorf("path '%s' not found", objectPath)
	}

	// Determine if the target is an array or object
	targetLine := strings.TrimSpace(lines[targetLineIndex])
	colonIndex := strings.Index(targetLine, ":")
	if colonIndex == -1 {
		return jsonStr, fmt.Errorf("invalid target line format")
	}

	valueStart := strings.TrimSpace(targetLine[colonIndex+1:])

	if strings.HasPrefix(valueStart, "[") {
		// Handle array
		if isArrayOfObjects(lines, targetLineIndex) {
			return insertIntoArrayObjects(lines, targetLineIndex, key, valueJSON)
		} else {
			return insertIntoArrayValues(lines, targetLineIndex, valueJSON)
		}
	} else if strings.HasPrefix(valueStart, "{") {
		// Handle object - check for duplicate key first
		if keyExistsInObject(lines, key, targetLineIndex, 2) {
			return jsonStr, fmt.Errorf("key '%s' already exists in target object", key)
		}
		return insertIntoObject(lines, targetLineIndex, key, valueJSON)
	}

	return jsonStr, fmt.Errorf("target path '%s' is not an object or array", objectPath)
}

// insertIntoObject inserts a key-value pair into an object
func insertIntoObject(lines []string, objectLineIndex int, key, valueJSON string) (string, error) {
	// Find the opening brace for this object
	objectLine := lines[objectLineIndex]
	if !strings.Contains(objectLine, "{") {
		// Object starts on next line, find it
		for i := objectLineIndex + 1; i < len(lines); i++ {
			if strings.Contains(strings.TrimSpace(lines[i]), "{") {
				objectLineIndex = i
				break
			}
		}
	}

	// Get proper indentation
	indent := detectIndentationForObject(lines, objectLineIndex)

	// Create the new property line
	newPropertyLine := indent + `"` + key + `": ` + valueJSON + ","

	// Insert right after the opening brace
	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:objectLineIndex+1]...)
	result = append(result, newPropertyLine)
	result = append(result, lines[objectLineIndex+1:]...)

	return strings.Join(result, "\n"), nil
}

// insertIntoArrayObjects inserts a key-value pair at the beginning of each array element
func insertIntoArrayObjects(lines []string, arrayLineIndex int, key, valueJSON string) (string, error) {
	// Check if key already exists in any array object
	if keyExistsInArrayObjects(lines, key, arrayLineIndex) {
		return strings.Join(lines, "\n"), fmt.Errorf("key '%s' already exists in one or more array objects", key)
	}

	result := make([]string, 0, len(lines))
	bracketDepth := 0
	braceDepth := 0
	inTargetArray := false
	i := 0

	// Copy lines until we reach the target array
	for i < len(lines) {
		line := lines[i]
		result = append(result, line)

		if i == arrayLineIndex {
			inTargetArray = true
		}

		if inTargetArray {
			trimmed := strings.TrimSpace(line)
			openBrackets := strings.Count(trimmed, "[")
			closeBrackets := strings.Count(trimmed, "]")
			openBraces := strings.Count(trimmed, "{")
			closeBraces := strings.Count(trimmed, "}")

			bracketDepth += openBrackets - closeBrackets
			braceDepth += openBraces - closeBraces

			// If we just opened an object inside the array, insert our property
			if bracketDepth > 0 && openBraces > 0 && strings.Contains(trimmed, "{") {
				indent := detectIndentationForObject(lines, i)
				newPropertyLine := indent + `"` + key + `": ` + valueJSON + ","
				result = append(result, newPropertyLine)
			}

			// If we've closed the target array, we're done
			if bracketDepth == 0 && i > arrayLineIndex {
				break
			}
		}

		i++
	}

	// Append remaining lines
	result = append(result, lines[i+1:]...)

	return strings.Join(result, "\n"), nil
}

// insertIntoArrayValues inserts a value at the beginning of an array
func insertIntoArrayValues(lines []string, arrayLineIndex int, valueJSON string) (string, error) {
	// Find the opening bracket
	bracketLineIndex := arrayLineIndex
	arrayLine := lines[arrayLineIndex]
	if !strings.Contains(arrayLine, "[") {
		// Array starts on next line, find it
		for i := arrayLineIndex + 1; i < len(lines); i++ {
			if strings.Contains(strings.TrimSpace(lines[i]), "[") {
				bracketLineIndex = i
				break
			}
		}
	}

	// Get proper indentation
	indent := detectIndentationForArray(lines, bracketLineIndex)

	// Create the new value line
	newValueLine := indent + valueJSON + ","

	// Insert right after the opening bracket
	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:bracketLineIndex+1]...)
	result = append(result, newValueLine)
	result = append(result, lines[bracketLineIndex+1:]...)

	return strings.Join(result, "\n"), nil
}
