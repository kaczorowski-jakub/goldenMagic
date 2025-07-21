package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/zserge/lorca"
)

//go:embed frontend
var fs embed.FS

// JSONFile represents a file with its path and name
type JSONFile struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// FileTreeNode represents a node in the file tree
type FileTreeNode struct {
	Name     string          `json:"name"`
	Path     string          `json:"path"`
	IsDir    bool            `json:"isDir"`
	Files    []JSONFile      `json:"files,omitempty"`
	Children []*FileTreeNode `json:"children,omitempty"`
	Count    int             `json:"count,omitempty"`
	AllFiles []JSONFile      `json:"allFiles,omitempty"`
}

// App represents the main application
type App struct {
	ui lorca.UI
}

// NewApp creates a new application instance
func NewApp() *App {
	return &App{}
}

// BrowseFolder scans a folder for files matching the extension filter and JSON key filter, returns a tree structure
func (a *App) BrowseFolder(folderPath, extensionFilter, jsonKeyFilter string) (*FileTreeNode, error) {

	allFiles := []JSONFile{}
	nodeMap := make(map[string]*FileTreeNode)

	// Parse the single extension filter
	var extension string
	if extensionFilter != "" {
		ext := strings.TrimSpace(extensionFilter)
		// Remove * if present and ensure it starts with .
		ext = strings.ReplaceAll(ext, "*", "")
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extension = strings.ToLower(ext)
	}

	folderPath, err := filepath.Abs(folderPath)
	if err != nil {
		return nil, err
	}

	// Create root node
	rootNode := &FileTreeNode{
		Name:     filepath.Base(folderPath),
		Path:     folderPath,
		IsDir:    true,
		Files:    []JSONFile{},
		Children: []*FileTreeNode{},
	}
	nodeMap[folderPath] = rootNode

	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Create directory node if it doesn't exist
			if _, exists := nodeMap[path]; !exists {
				dirNode := &FileTreeNode{
					Name:     info.Name(),
					Path:     path,
					IsDir:    true,
					Files:    []JSONFile{},
					Children: []*FileTreeNode{},
				}
				nodeMap[path] = dirNode

				// Add to parent directory
				parentPath := filepath.Dir(path)
				if parentNode, exists := nodeMap[parentPath]; exists {
					parentNode.Children = append(parentNode.Children, dirNode)
				}
			}
		} else {
			fileName := strings.ToLower(info.Name())

			// Check if file matches the extension filter
			matchesExtension := extension == "" || strings.HasSuffix(fileName, extension)

			if matchesExtension {
				// Check JSON key filter if specified
				matchesJsonKey := true
				if jsonKeyFilter != "" {
					matchesJsonKey = false
					// Read and parse the file to check for the key
					if content, err := os.ReadFile(path); err == nil {
						var jsonContent any
						if err := json.Unmarshal(content, &jsonContent); err == nil {
							// Check if the specified key exists anywhere in the JSON (deep search)
							if containsKeyDeep(jsonContent, jsonKeyFilter) {
								matchesJsonKey = true
							}
						}
					}
				}

				if matchesJsonKey {
					jsonFile := JSONFile{
						Path: path,
						Name: info.Name(),
					}

					// Add file to its parent directory node
					parentPath := filepath.Dir(path)
					if parentNode, exists := nodeMap[parentPath]; exists {
						parentNode.Files = append(parentNode.Files, jsonFile)
						parentNode.Count = len(parentNode.Files)
						allFiles = append(allFiles, jsonFile)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	rootNode.Count = len(allFiles)
	rootNode.AllFiles = allFiles
	return rootNode, nil
}

// GetJSONFileContent retrieves the content of a specific JSON file
func (a *App) GetJSONFileContent(filePath string) (map[string]any, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var jsonContent map[string]any
	if err := json.Unmarshal(content, &jsonContent); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return jsonContent, nil
}

// AddJSONItemToFiles adds a new JSON item to multiple files at once
func (a *App) AddJSONItemToFiles(filePaths []string, objectPath, key string, value any) map[string]string {
	results := make(map[string]string)

	for _, filePath := range filePaths {
		err := a.addJSONItemToSingleFile(filePath, objectPath, key, value)
		if err != nil {
			results[filePath] = "ERROR: " + err.Error()
		} else {
			results[filePath] = "SUCCESS"
		}
	}

	return results
}

// addJSONItemToSingleFile adds a new JSON item to a single file while preserving structure
func (a *App) addJSONItemToSingleFile(filePath, objectPath, key string, value any) error {
	// Read existing file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Parse JSON to verify it's valid and find insertion point
	var jsonContent map[string]any
	if err := json.Unmarshal(content, &jsonContent); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	// Verify the object path exists (if specified)
	if objectPath != "" {
		_, err := navigateToObject(jsonContent, objectPath)
		if err != nil {
			return fmt.Errorf("error navigating to object path '%s': %v", objectPath, err)
		}
	}

	// Convert value to JSON string
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error marshaling value: %v", err)
	}

	// Perform string-based insertion to preserve original formatting and order
	updatedContent, err := insertJSONKeyValue(string(content), objectPath, key, string(valueJSON))
	if err != nil {
		return fmt.Errorf("error inserting JSON key-value: %v", err)
	}

	if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

// insertJSONKeyValue inserts a key-value pair into JSON string while preserving structure
func insertJSONKeyValue(jsonStr, objectPath, key, valueJSON string) (string, error) {
	if objectPath == "" {
		// Insert at root level
		return insertAtRoot(jsonStr, key, valueJSON)
	}

	// For context paths, find the target object key and insert there
	pathParts := strings.Split(objectPath, ".")
	targetKey := pathParts[len(pathParts)-1]

	return insertAtContextPath(jsonStr, targetKey, key, valueJSON)
}

// insertAtRoot inserts a key-value pair at the root level
func insertAtRoot(jsonStr, key, valueJSON string) (string, error) {
	lines := strings.Split(jsonStr, "\n")

	// Find the last property in the root object
	insertIndex := -1
	currentDepth := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track depth by counting braces
		if strings.Contains(trimmed, "{") {
			currentDepth++
		}
		if strings.Contains(trimmed, "}") {
			if currentDepth == 1 {
				// Found the closing brace of root object
				insertIndex = i
				break
			}
			currentDepth--
		}

		// If we're at root level and this line has a property, it's a potential insertion point
		if currentDepth == 1 && strings.Contains(trimmed, ":") && (strings.HasSuffix(trimmed, ",") || !strings.HasSuffix(trimmed, "{")) {
			insertIndex = i + 1
		}
	}

	return insertLineAtIndex(lines, insertIndex, key, valueJSON)
}

// insertAtContextPath inserts a key-value pair in the first object that contains the target key
func insertAtContextPath(jsonStr, targetKey, key, valueJSON string) (string, error) {
	lines := strings.Split(jsonStr, "\n")

	// Find the target object by searching for the target key
	insertIndex := -1
	currentDepth := 0
	inTargetObject := false
	targetObjectDepth := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track depth by counting braces
		if strings.Contains(trimmed, "{") {
			currentDepth++
		}

		// Check if this line contains our target key
		if !inTargetObject && strings.Contains(trimmed, `"`+targetKey+`"`) && strings.Contains(trimmed, ":") {
			// Check if the value after the colon is an object (contains {)
			if strings.Contains(trimmed, "{") {
				inTargetObject = true
				targetObjectDepth = currentDepth
			}
		}

		if strings.Contains(trimmed, "}") {
			if inTargetObject && currentDepth == targetObjectDepth {
				// Found the closing brace of our target object
				insertIndex = i
				break
			}
			currentDepth--
		}

		// If we're in the target object and this line has a property, it's a potential insertion point
		if inTargetObject && currentDepth == targetObjectDepth && strings.Contains(trimmed, ":") {
			insertIndex = i + 1
		}
	}

	if insertIndex == -1 {
		return "", fmt.Errorf("could not find target object with key '%s'", targetKey)
	}

	return insertLineAtIndex(lines, insertIndex, key, valueJSON)
}

// insertLineAtIndex inserts a new JSON line at the specified index
func insertLineAtIndex(lines []string, insertIndex int, key, valueJSON string) (string, error) {
	if insertIndex == -1 {
		return "", fmt.Errorf("could not find insertion point in JSON")
	}

	// Determine indentation by looking at existing lines
	indent := "  " // default
	if insertIndex > 0 {
		prevLine := lines[insertIndex-1]
		leadingSpaces := len(prevLine) - len(strings.TrimLeft(prevLine, " \t"))
		if leadingSpaces > 0 {
			indent = prevLine[:leadingSpaces]
		}
	}

	// Check if we need a comma before our new line
	if insertIndex > 0 {
		prevLine := strings.TrimSpace(lines[insertIndex-1])
		if prevLine != "" && !strings.HasSuffix(prevLine, "{") && !strings.HasSuffix(prevLine, ",") {
			// Add comma to previous line
			lines[insertIndex-1] = lines[insertIndex-1] + ","
		}
	}

	// Create the new line
	newLine := indent + `"` + key + `": ` + valueJSON

	// Check if the next line is a closing brace (no comma needed)
	if insertIndex < len(lines) {
		nextLine := strings.TrimSpace(lines[insertIndex])
		if !strings.HasPrefix(nextLine, "}") {
			newLine += ","
		}
	}

	// Insert the new line
	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:insertIndex]...)
	result = append(result, newLine)
	result = append(result, lines[insertIndex:]...)

	return strings.Join(result, "\n"), nil
}

// navigateToObject navigates to a nested object using dot notation or finds it by context
func navigateToObject(jsonData map[string]any, path string) (map[string]any, error) {
	if path == "" {
		return jsonData, nil
	}

	parts := strings.Split(path, ".")

	// Try absolute path first
	if obj, err := navigateAbsolutePath(jsonData, parts); err == nil {
		return obj, nil
	}

	// If absolute path fails, try to find by context (search for the path anywhere in the structure)
	if obj, err := findObjectByContext(jsonData, parts); err == nil {
		return obj, nil
	}

	return nil, fmt.Errorf("could not find object at path '%s' (tried absolute and context search)", path)
}

// navigateAbsolutePath navigates using absolute path from root
func navigateAbsolutePath(jsonData map[string]any, parts []string) (map[string]any, error) {
	current := jsonData

	for i, part := range parts {
		if value, exists := current[part]; exists {
			if i == len(parts)-1 {
				// Last part - this should be the target object
				if obj, ok := value.(map[string]any); ok {
					return obj, nil
				}
				return nil, fmt.Errorf("path does not point to an object")
			} else {
				// Intermediate part - continue navigation
				if obj, ok := value.(map[string]any); ok {
					current = obj
				} else {
					return nil, fmt.Errorf("path part '%s' is not an object", part)
				}
			}
		} else {
			return nil, fmt.Errorf("path part '%s' does not exist", part)
		}
	}

	return current, nil
}

// findObjectByContext searches for an object by context path (can be anywhere in the structure)
func findObjectByContext(jsonData map[string]any, pathParts []string) (map[string]any, error) {
	targetKey := pathParts[len(pathParts)-1] // The final key we're looking for

	// Search recursively for the target key
	found := findObjectWithKey(jsonData, targetKey)
	if len(found) > 0 {
		// Return the first match (in a real scenario, you might want to handle multiple matches differently)
		return found[0], nil
	}

	return nil, fmt.Errorf("could not find object with key '%s' in context", targetKey)
}

// findObjectWithKey recursively searches for objects that contain a specific key
func findObjectWithKey(data any, targetKey string) []map[string]any {
	var results []map[string]any

	switch v := data.(type) {
	case map[string]any:
		// Check if this object contains the target key and that key points to an object
		if value, exists := v[targetKey]; exists {
			if obj, ok := value.(map[string]any); ok {
				results = append(results, obj)
			}
		}

		// Recursively search in nested objects
		for _, value := range v {
			results = append(results, findObjectWithKey(value, targetKey)...)
		}

	case []any:
		// Search in array elements
		for _, item := range v {
			results = append(results, findObjectWithKey(item, targetKey)...)
		}
	}

	return results
}

// GetBasePath returns the base path from environment variable or current directory
func (a *App) GetBasePath() (string, error) {
	basePath := os.Getenv("JSON_MANAGER_BASE_PATH")
	if basePath == "" {
		var err error
		basePath, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	return basePath, nil
}

// containsKeyDeep recursively searches for a key in any nested JSON structure
func containsKeyDeep(jsonData any, targetKey string) bool {
	switch data := jsonData.(type) {
	case map[string]any:
		// Check if the key exists at this level
		if _, exists := data[targetKey]; exists {
			return true
		}
		// Recursively check all values
		for _, value := range data {
			if containsKeyDeep(value, targetKey) {
				return true
			}
		}
	case []any:
		// For arrays, check each element
		for _, item := range data {
			if containsKeyDeep(item, targetKey) {
				return true
			}
		}
	}
	return false
}

func main() {
	// Load environment variables from config.env file
	err := godotenv.Load("config.env")
	if err != nil {
		log.Printf("Warning: Could not load config.env file: %v", err)
		log.Printf("Using system environment variables or defaults")
	}

	app := NewApp()

	// Create Lorca UI with Chrome args for better compatibility
	ui, err := lorca.New("", "", 1200, 800, "--disable-web-security", "--disable-features=VizDisplayCompositor", "--remote-allow-origins=*")
	if err != nil {
		log.Printf("Failed to start with Chrome: %v", err)
		os.Exit(1)
	}

	defer ui.Close()

	app.ui = ui

	// Bind Go functions to JavaScript
	ui.Bind("browseFolder", app.BrowseFolder)
	ui.Bind("getJSONFileContent", app.GetJSONFileContent)
	ui.Bind("addJSONItemToFiles", app.AddJSONItemToFiles)
	ui.Bind("getBasePath", app.GetBasePath)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go http.Serve(ln, http.FileServer(http.FS(fs)))
	ui.Load(fmt.Sprintf("http://%s/frontend", ln.Addr()))

	/*
			// Load the HTML file and CSS content
			htmlContent, err := os.ReadFile("frontend/index.html")
			if err != nil {
				log.Fatal("Error reading frontend/index.html:", err)
			}

			// Load CSS content
			cssContent, err := os.ReadFile("frontend/css/styles.css")
			if err != nil {
				log.Fatal("Error reading frontend/css/styles.css:", err)
			}

			// Embed CSS directly into HTML
			htmlWithCSS := strings.Replace(string(htmlContent),
				`<link rel="stylesheet" href="css/styles.css">`,
				`<style>`+string(cssContent)+`</style>`, 1)

			ui.Load("data:text/html," + url.PathEscape(htmlWithCSS))


		// Wait until the interrupt signal arrives or browser window is closed
		<-ui.Done()
	*/

	// Wait until the interrupt signal arrives or browser window is closed
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	case <-ui.Done():
	}

	log.Println("exiting...")
}
