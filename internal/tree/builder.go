package tree

import (
	"path/filepath"
	"strings"

	"goldenMagic/internal/fileops"
)

// FileTreeNode represents a node in the file tree
type FileTreeNode struct {
	Name     string             `json:"name"`
	Path     string             `json:"path,omitempty"`
	IsDir    bool               `json:"isDir"`
	Files    []fileops.JSONFile `json:"files,omitempty"`
	Children []*FileTreeNode    `json:"children,omitempty"`
	Count    int                `json:"count"`
	BasePath string             `json:"basePath,omitempty"` // Which base path this node belongs to
}

// BuildFileTreeFromMultiplePaths creates a unified tree structure from files across multiple base paths
func BuildFileTreeFromMultiplePaths(files []fileops.JSONFile, basePaths []string) *FileTreeNode {
	if len(files) == 0 {
		return &FileTreeNode{
			Name:  "No Results",
			IsDir: true,
			Count: 0,
		}
	}

	// Create virtual root node that contains all base paths
	root := &FileTreeNode{
		Name:     "Search Results",
		IsDir:    true,
		Children: make([]*FileTreeNode, 0),
		Count:    len(files),
	}

	// Group files by base path
	filesByBasePath := fileops.GroupFilesByBasePath(files)

	// Create a tree for each base path
	for _, basePath := range basePaths {
		if pathFiles, exists := filesByBasePath[basePath]; exists && len(pathFiles) > 0 {
			basePathTree := BuildFileTree(pathFiles, basePath)
			if basePathTree != nil {
				root.Children = append(root.Children, basePathTree)
			}
		}
	}

	// If only one base path has results, return that tree directly
	if len(root.Children) == 1 {
		return root.Children[0]
	}

	return root
}

// BuildFileTree creates a hierarchical tree structure from a flat list of files
func BuildFileTree(files []fileops.JSONFile, basePath string) *FileTreeNode {
	if len(files) == 0 {
		return &FileTreeNode{
			Name:     filepath.Base(basePath),
			Path:     basePath,
			IsDir:    true,
			Count:    0,
			BasePath: basePath,
		}
	}

	// Create root node
	root := &FileTreeNode{
		Name:     filepath.Base(basePath),
		Path:     basePath,
		IsDir:    true,
		Children: make([]*FileTreeNode, 0),
		Count:    len(files),
		BasePath: basePath,
	}

	// Group files by directory
	dirMap := make(map[string]*FileTreeNode)
	dirMap[basePath] = root

	for _, file := range files {
		dir := filepath.Dir(file.Path)

		// Create directory nodes if they don't exist
		createDirNodes(dirMap, dir, basePath, root)

		// Add file to its directory
		if dirNode, exists := dirMap[dir]; exists {
			dirNode.Files = append(dirNode.Files, file)
		}
	}

	// Calculate counts for all nodes
	calculateCounts(root)

	return root
}

// createDirNodes creates all necessary directory nodes in the path
func createDirNodes(dirMap map[string]*FileTreeNode, targetDir, basePath string, root *FileTreeNode) {
	if targetDir == basePath || dirMap[targetDir] != nil {
		return
	}

	// Get parent directory
	parentDir := filepath.Dir(targetDir)

	// Recursively create parent nodes
	createDirNodes(dirMap, parentDir, basePath, root)

	// Create current directory node
	dirNode := &FileTreeNode{
		Name:     filepath.Base(targetDir),
		Path:     targetDir,
		IsDir:    true,
		Children: make([]*FileTreeNode, 0),
		Files:    make([]fileops.JSONFile, 0),
		BasePath: basePath,
	}

	// Add to parent
	if parentNode, exists := dirMap[parentDir]; exists {
		parentNode.Children = append(parentNode.Children, dirNode)
	}

	dirMap[targetDir] = dirNode
}

// calculateCounts recursively calculates file counts for each directory node
func calculateCounts(node *FileTreeNode) int {
	if !node.IsDir {
		return 0
	}

	count := len(node.Files)

	for _, child := range node.Children {
		count += calculateCounts(child)
	}

	node.Count = count
	return count
}

// FlattenTree converts a tree structure back to a flat list of files
func FlattenTree(node *FileTreeNode) []fileops.JSONFile {
	var files []fileops.JSONFile

	if node == nil {
		return files
	}

	// Add files from current node
	files = append(files, node.Files...)

	// Recursively add files from children
	for _, child := range node.Children {
		files = append(files, FlattenTree(child)...)
	}

	return files
}

// FindNodeByPath finds a node in the tree by its path
func FindNodeByPath(root *FileTreeNode, targetPath string) *FileTreeNode {
	if root == nil {
		return nil
	}

	if root.Path == targetPath {
		return root
	}

	for _, child := range root.Children {
		if found := FindNodeByPath(child, targetPath); found != nil {
			return found
		}
	}

	return nil
}

// GetAllDirectories returns all directory paths in the tree
func GetAllDirectories(node *FileTreeNode) []string {
	var dirs []string

	if node == nil {
		return dirs
	}

	if node.IsDir && node.Path != "" {
		dirs = append(dirs, node.Path)
	}

	for _, child := range node.Children {
		dirs = append(dirs, GetAllDirectories(child)...)
	}

	return dirs
}

// GetAllBasePaths returns all unique base paths in the tree
func GetAllBasePaths(node *FileTreeNode) []string {
	basePathMap := make(map[string]bool)
	collectBasePaths(node, basePathMap)

	var basePaths []string
	for basePath := range basePathMap {
		if basePath != "" {
			basePaths = append(basePaths, basePath)
		}
	}

	return basePaths
}

// collectBasePaths recursively collects base paths from the tree
func collectBasePaths(node *FileTreeNode, basePathMap map[string]bool) {
	if node == nil {
		return
	}

	if node.BasePath != "" {
		basePathMap[node.BasePath] = true
	}

	for _, child := range node.Children {
		collectBasePaths(child, basePathMap)
	}
}

// FilterTreeByExtension filters the tree to only include files with specific extensions
func FilterTreeByExtension(node *FileTreeNode, extensions []string) *FileTreeNode {
	if node == nil {
		return nil
	}

	filteredNode := &FileTreeNode{
		Name:     node.Name,
		Path:     node.Path,
		IsDir:    node.IsDir,
		Files:    make([]fileops.JSONFile, 0),
		Children: make([]*FileTreeNode, 0),
		BasePath: node.BasePath,
	}

	// Filter files
	for _, file := range node.Files {
		for _, ext := range extensions {
			cleanExt := strings.TrimPrefix(ext, "*")
			if strings.HasSuffix(strings.ToLower(file.Name), strings.ToLower(cleanExt)) {
				filteredNode.Files = append(filteredNode.Files, file)
				break
			}
		}
	}

	// Filter children
	for _, child := range node.Children {
		filteredChild := FilterTreeByExtension(child, extensions)
		if filteredChild != nil && (len(filteredChild.Files) > 0 || len(filteredChild.Children) > 0) {
			filteredNode.Children = append(filteredNode.Children, filteredChild)
		}
	}

	// Calculate count
	calculateCounts(filteredNode)

	return filteredNode
}

// FilterTreeByBasePath filters the tree to only include nodes from specific base paths
func FilterTreeByBasePath(node *FileTreeNode, allowedBasePaths []string) *FileTreeNode {
	if node == nil {
		return nil
	}

	// Create allowed paths map for quick lookup
	allowedMap := make(map[string]bool)
	for _, path := range allowedBasePaths {
		allowedMap[path] = true
	}

	// If this node's base path is not allowed, skip it entirely
	if node.BasePath != "" && !allowedMap[node.BasePath] {
		return nil
	}

	filteredNode := &FileTreeNode{
		Name:     node.Name,
		Path:     node.Path,
		IsDir:    node.IsDir,
		Files:    make([]fileops.JSONFile, 0),
		Children: make([]*FileTreeNode, 0),
		BasePath: node.BasePath,
	}

	// Filter files
	for _, file := range node.Files {
		if allowedMap[file.BasePath] {
			filteredNode.Files = append(filteredNode.Files, file)
		}
	}

	// Filter children
	for _, child := range node.Children {
		filteredChild := FilterTreeByBasePath(child, allowedBasePaths)
		if filteredChild != nil {
			filteredNode.Children = append(filteredNode.Children, filteredChild)
		}
	}

	// Calculate count
	calculateCounts(filteredNode)

	return filteredNode
}
