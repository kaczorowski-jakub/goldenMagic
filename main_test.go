package main_test

import (
	"fmt"
	"goldenMagic/internal/jsonops"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_duplicate_check(t *testing.T) {
	// Test JSON with existing keys
	testJSON := `{
		"name": "test",
		"version": "1.0.0",
		"scripts": {
			"start": "node index.js",
			"test": "jest"
		},
		"dependencies": {
			"express": "^4.18.0"
		}
	}`

	fmt.Println("=== Testing Duplicate Detection ===")
	fmt.Println("Original JSON:")
	fmt.Println(testJSON)
	fmt.Println()

	// Test 1: Try to add a new key that doesn't exist (should succeed)
	fmt.Println("Test 1: Adding new key 'build' after 'test' (should succeed)")
	newObject1 := `"webpack --mode production"`
	_, err1 := jsonops.InsertItemAfter(testJSON, "test", "build", newObject1)
	require.NoError(t, err1)

	// Test 2: Try to add a key that already exists (should fail)
	fmt.Println("Test 2: Adding existing key 'start' after 'test' (should fail)")
	newObject2 := `"npm start"`
	_, err2 := jsonops.InsertItemAfter(testJSON, "test", "start", newObject2)
	require.Error(t, err2)
	require.Contains(t, err2.Error(), "object with key 'start' already exists")

	// Test 3: Try to add a key at different nesting level (should succeed)
	fmt.Println("Test 3: Adding 'react' after 'express' in dependencies (should succeed)")
	newObject3 := `"^16.0.0"`
	result3, err3 := jsonops.InsertItemAfter(testJSON, "express", "react", newObject3)
	if err3 != nil {
		fmt.Printf("❌ Error: %v\n", err3)
	} else {
		fmt.Println("✅ Success! New JSON:")
		fmt.Println(result3)
	}
}
