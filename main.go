package main

import (
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"goldenMagic/internal/config"
	"goldenMagic/internal/fileops"
	"goldenMagic/internal/jsonops"
	"goldenMagic/internal/tree"

	"github.com/zserge/lorca"
)

//go:embed frontend
var frontendFiles embed.FS

// App represents the main application
type App struct {
	config    *config.Config
	startTime time.Time
	stats     *AppStats
}

// AppStats tracks application usage statistics
type AppStats struct {
	SearchOperations int
	FilesProcessed   int
	UpdateOperations int
	Errors           int
}

// NewApp creates a new application instance
func NewApp() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	return &App{
		config:    cfg,
		startTime: time.Now(),
		stats:     &AppStats{},
	}, nil
}

// logOperation logs an operation with timing and context
func (a *App) logOperation(operation string, duration time.Duration, err error, details map[string]interface{}) {
	level := "INFO"
	if err != nil {
		level = "ERROR"
		a.stats.Errors++
	}

	log.Printf("[%s] %s completed in %v | Details: %+v | Error: %v",
		level, operation, duration, details, err)
}

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatal("Failed to initialize app:", err)
	}

	log.Printf("🚀 Starting goldenMagic application at %v", app.startTime)
	log.Printf("📁 Configured base paths: %v", app.config.GetBasePaths())

	// Start HTTP server for static files
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go http.Serve(listener, http.FileServer(http.FS(frontendFiles)))
	log.Printf("🌐 HTTP server started on %s", listener.Addr())

	// Create Lorca UI
	ui, err := lorca.New(fmt.Sprintf("http://%s/frontend/", listener.Addr()), "", 1024, 768,
		"--disable-web-security",
		"--disable-features=VizDisplayCompositor",
		"--remote-allow-origins=*")
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	log.Printf("🖥️  UI initialized successfully")

	// Bind Go functions to JavaScript
	ui.Bind("browseFolder", app.BrowseFolder)
	ui.Bind("getJSONFileContent", app.GetJSONFileContent)
	ui.Bind("addJSONItemToFiles", app.AddJSONItemToFiles)
	ui.Bind("addJSONItemAfter", app.AddJSONItemAfter)
	ui.Bind("replaceKeys", app.ReplaceKeys)
	ui.Bind("getBasePaths", app.GetBasePaths)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	select {
	case <-c:
		log.Println("🛑 Interrupt signal received")
	case <-ui.Done():
		log.Println("🔚 UI closed")
	}

	// Print final statistics
	uptime := time.Since(app.startTime)
	log.Printf("📊 Session Statistics:")
	log.Printf("   ⏱️  Uptime: %v", uptime)
	log.Printf("   🔍 Search operations: %d", app.stats.SearchOperations)
	log.Printf("   📄 Files processed: %d", app.stats.FilesProcessed)
	log.Printf("   ✏️  Update operations: %d", app.stats.UpdateOperations)
	log.Printf("   ❌ Errors encountered: %d", app.stats.Errors)

	log.Println("👋 Exiting goldenMagic...")
}

// BrowseFolder searches for files across all configured base paths and returns a unified tree structure
func (a *App) BrowseFolder(extensionFilter, jsonKeyFilter string) (*tree.FileTreeNode, error) {
	start := time.Now()
	a.stats.SearchOperations++

	// Get only valid base paths
	validBasePaths := a.config.GetValidBasePaths()
	if len(validBasePaths) == 0 {
		err := fmt.Errorf("no valid base paths configured")
		a.logOperation("BrowseFolder", time.Since(start), err, map[string]any{
			"extensionFilter": extensionFilter,
			"jsonKeyFilter":   jsonKeyFilter,
		})
		return &tree.FileTreeNode{
			Name:  "No Valid Paths",
			IsDir: true,
			Count: 0,
		}, err
	}

	files, err := fileops.BrowseFolders(validBasePaths, extensionFilter, jsonKeyFilter)
	if err != nil {
		a.logOperation("BrowseFolder", time.Since(start), err, map[string]any{
			"extensionFilter": extensionFilter,
			"jsonKeyFilter":   jsonKeyFilter,
			"basePaths":       validBasePaths,
		})
		return nil, fmt.Errorf("error browsing folders: %v", err)
	}

	result := tree.BuildFileTreeFromMultiplePaths(files, validBasePaths)

	a.logOperation("BrowseFolder", time.Since(start), nil, map[string]any{
		"extensionFilter": extensionFilter,
		"jsonKeyFilter":   jsonKeyFilter,
		"filesFound":      len(files),
		"basePaths":       len(validBasePaths),
	})

	return result, nil
}

// GetJSONFileContent returns the content of a JSON file
func (a *App) GetJSONFileContent(filePath string) (string, error) {
	start := time.Now()

	log.Printf("📖 Loading file content: %s", filePath)

	content, err := fileops.GetJSONFileContent(filePath)

	a.logOperation("GetJSONFileContent", time.Since(start), err, map[string]any{
		"filePath":      filePath,
		"contentLength": len(content),
	})

	if err != nil {
		log.Printf("❌ Failed to load file content: %v", err)
		return "", err
	}

	log.Printf("✅ File content loaded successfully, length: %d", len(content))
	return content, nil
}

// AddJSONItemToFiles adds a JSON item to multiple files
func (a *App) AddJSONItemToFiles(filePaths []string, objectPath, key string, value any) map[string]string {
	results := make(map[string]string)

	for _, filePath := range filePaths {
		// Read existing file
		content, err := fileops.ReadFile(filePath)
		if err != nil {
			results[filePath] = "ERROR: error reading file: " + err.Error()
			continue
		}

		// Insert the JSON key-value pair while preserving structure
		updatedContent, err := jsonops.InsertJSONKeyValue(string(content), objectPath, key, value)
		if err != nil {
			results[filePath] = "ERROR: error inserting JSON: " + err.Error()
			continue
		}

		// Write updated content back to file
		err = fileops.WriteFile(filePath, []byte(updatedContent))
		if err != nil {
			results[filePath] = "ERROR: error writing file: " + err.Error()
			continue
		}

		results[filePath] = "SUCCESS"
	}

	return results
}

// AddJSONObjectAfter adds a complete JSON object after a target object in specified files
func (a *App) AddJSONItemAfter(filePaths []string, targetKey, newObjectKey, newObjectJSON string) map[string]string {
	start := time.Now()
	a.stats.UpdateOperations++

	results := make(map[string]string)

	for _, filePath := range filePaths {
		a.stats.FilesProcessed++

		// Read the file
		content, err := fileops.ReadFile(filePath)
		if err != nil {
			results[filePath] = fmt.Sprintf("ERROR: reading file: %v", err)
			continue
		}

		// Insert the new object after the target
		updatedContent, err := jsonops.InsertItemAfter(string(content), targetKey, newObjectKey, newObjectJSON)
		if err != nil {
			// Check if it's a duplicate key error
			if strings.Contains(err.Error(), "already exists") {
				results[filePath] = fmt.Sprintf("SKIPPED: %v", err)
			} else {
				results[filePath] = fmt.Sprintf("ERROR: inserting object: %v", err)
			}
			continue
		}

		// Write back to file
		err = fileops.WriteFile(filePath, []byte(updatedContent))
		if err != nil {
			results[filePath] = fmt.Sprintf("ERROR: writing file: %v", err)
			continue
		}

		results[filePath] = "SUCCESS"
	}

	successCount := 0
	skippedCount := 0
	errorCount := 0
	for _, result := range results {
		if strings.HasPrefix(result, "SUCCESS") {
			successCount++
		} else if strings.HasPrefix(result, "SKIPPED") {
			skippedCount++
		} else {
			errorCount++
		}
	}

	a.logOperation("AddJSONItemAfter", time.Since(start), nil, map[string]any{
		"targetKey":      targetKey,
		"newObjectKey":   newObjectKey,
		"filesProcessed": len(filePaths),
		"successCount":   successCount,
		"skippedCount":   skippedCount,
		"errorCount":     errorCount,
	})

	return results
}

// GetBasePaths returns all configured base paths
func (a *App) GetBasePaths() ([]string, error) {
	return a.config.GetBasePaths(), nil
}

// GetValidBasePaths returns only the base paths that exist and are accessible
func (a *App) GetValidBasePaths() ([]string, error) {
	return a.config.GetValidBasePaths(), nil
}

// GetBasePathInfo returns detailed information about all base paths
func (a *App) GetBasePathInfo() (map[string]interface{}, error) {
	basePaths := a.config.GetBasePaths()
	validPaths := a.config.GetValidBasePaths()

	info := map[string]interface{}{
		"allPaths":     basePaths,
		"validPaths":   validPaths,
		"totalCount":   len(basePaths),
		"validCount":   len(validPaths),
		"invalidCount": len(basePaths) - len(validPaths),
	}

	// Add validity status for each path
	pathStatus := make(map[string]bool)
	for _, path := range basePaths {
		pathStatus[path] = a.config.IsValidBasePath(path)
	}
	info["pathStatus"] = pathStatus

	return info, nil
}

// ReplaceKeys replaces old keys with new keys in selected files using string replacement
func (a *App) ReplaceKeys(oldKey, newKey string, selectedFiles []string) ([]jsonops.ReplaceKeyResult, error) {
	log.Printf("🔄 Starting key replace operation: oldKey=%s, newKey=%s, files=%d", oldKey, newKey, len(selectedFiles))

	request := jsonops.ReplaceKeyRequest{
		OldKey:        oldKey,
		NewKey:        newKey,
		SelectedFiles: selectedFiles,
	}

	results, err := jsonops.ReplaceKeyInFiles(request)
	if err != nil {
		log.Printf("❌ Replace operation failed: %v", err)
		return nil, err
	}

	successCount := 0
	totalReplacements := 0
	for _, result := range results {
		if result.Success {
			successCount++
			totalReplacements += result.ReplacementCount
		}
	}

	log.Printf("✅ Replace operation completed: %d/%d files successful, %d total replacements",
		successCount, len(selectedFiles), totalReplacements)

	return results, nil
}
