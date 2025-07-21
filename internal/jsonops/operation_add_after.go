package jsonops

import (
	"encoding/json"
	"fmt"
	"strings"
)

// InsertItemAfter adds a JSON object after all occurrences of a target key in the JSON string
func InsertItemAfter(jsonStr, targetKey, newObjectKey, newObjectJSON string) (string, error) {
	lines := strings.Split(jsonStr, "\n")

	// Find all occurrences of the target key
	var targetLineIndices []int
	for i, line := range lines {
		// Look for the target key (accounting for quotes and colon)
		if strings.Contains(line, `"`+targetKey+`"`) && strings.Contains(line, ":") {
			targetLineIndices = append(targetLineIndices, i)
		}
	}

	if len(targetLineIndices) == 0 {
		return "", fmt.Errorf("target key '%s' not found", targetKey)
	}

	// Validate the new object JSON once
	var newObj interface{}
	if err := json.Unmarshal([]byte(newObjectJSON), &newObj); err != nil {
		return "", fmt.Errorf("invalid JSON for new object: %v", err)
	}

	// Convert to properly indented JSON template
	formattedJSON, err := json.MarshalIndent(newObj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error formatting new object: %v", err)
	}

	// Process each target occurrence from last to first to avoid index shifting
	result := lines
	for i := len(targetLineIndices) - 1; i >= 0; i-- {
		targetLineIndex := targetLineIndices[i]

		// Find the end of this target object/value
		insertIndex := findEndOfValue(result, targetLineIndex)
		if insertIndex == -1 {
			continue // Skip this occurrence if we can't find the end
		}

		// Detect indentation from the target line
		targetLine := result[targetLineIndex]
		indentation := ""
		for _, char := range targetLine {
			if char == ' ' || char == '\t' {
				indentation += string(char)
			} else {
				break
			}
		}

		// Split the formatted JSON into lines and apply base indentation
		newObjLines := strings.Split(string(formattedJSON), "\n")
		for j, line := range newObjLines {
			if j == 0 {
				// First line gets the key and value
				newObjLines[j] = indentation + `"` + newObjectKey + `": ` + line
			} else {
				// Subsequent lines get additional indentation
				newObjLines[j] = indentation + "  " + line
			}
		}

		// Add comma to the target line if it doesn't have one and we're not at the end
		if insertIndex < len(result)-1 && !strings.HasSuffix(strings.TrimSpace(result[insertIndex]), ",") {
			result[insertIndex] = strings.TrimRight(result[insertIndex], " \t") + ","
		}

		// Add comma to our new object if needed (not the last item)
		lastNewLine := newObjLines[len(newObjLines)-1]
		if insertIndex < len(result)-1 && !strings.HasSuffix(strings.TrimSpace(lastNewLine), ",") {
			// Check if the next non-empty line after insert position contains closing brace
			nextNonEmptyIndex := insertIndex + 1
			for nextNonEmptyIndex < len(result) && strings.TrimSpace(result[nextNonEmptyIndex]) == "" {
				nextNonEmptyIndex++
			}

			if nextNonEmptyIndex < len(result) && !strings.Contains(strings.TrimSpace(result[nextNonEmptyIndex]), "}") {
				newObjLines[len(newObjLines)-1] = lastNewLine + ","
			}
		}

		// Insert the new object after the target
		newResult := make([]string, 0, len(result)+len(newObjLines))
		newResult = append(newResult, result[:insertIndex+1]...)
		newResult = append(newResult, newObjLines...)
		newResult = append(newResult, result[insertIndex+1:]...)
		result = newResult
	}

	return strings.Join(result, "\n"), nil
}
