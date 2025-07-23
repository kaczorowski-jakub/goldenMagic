# ✨ goldenMagic - Advanced JSON File Explorer & Editor

A powerful desktop application built with Go and the Lorca framework that provides advanced JSON file exploration, filtering, and mass editing capabilities with a beautiful web-based interface.

## ✨ Key Features

### 🔍 **Advanced File Discovery**
- **Multi-Path Search**: Search across multiple directories simultaneously
- **Tree Structure Display**: Hierarchical view of folders and files
- **Dual Filtering System**: Filter by file extension AND JSON content
- **Deep JSON Key Search**: Find files containing specific keys at any nesting level
- **Real-time Search**: Instant results as you type

### 📊 **Beautiful JSON Viewer**
- **Syntax Highlighting**: Color-coded JSON with line numbers
- **Collapsible Tree View**: Organized folder structure
- **Copy to Clipboard**: One-click JSON content copying
- **File Statistics**: Properties count, lines, and file size
- **Structure Preservation**: Maintains original formatting and key order

### 🚀 **Add JSON Item Operations**
- **Bulk JSON Editing**: Add properties to all filtered files at once
- **Insert After Object**: Add complete JSON objects after specified target objects
- **Replace Keys**: Rename JSON keys across multiple files using text replacement
- **Context-Aware Paths**: Smart object path detection and auto-completion
- **Structure Preservation**: Maintains original file formatting and key order
- **Progress Tracking**: Real-time feedback on bulk operations
- **Error Handling**: Detailed success/failure reporting per file

### 🎯 **Smart Filtering**
- **File Extension Filter**: Support for *.json, *.golden, and custom extensions
- **JSON Key Filter**: Find files containing specific nested properties
- **Combined Filtering**: Use both filters simultaneously for precise results

## Prerequisites

- Go 1.21 or higher
- Chrome/Chromium browser (required by Lorca)

## Installation

1. Clone or download this repository
2. Navigate to the project directory
3. Install dependencies:
   ```bash
   go mod tidy
   ```

## 🚀 Quick Start

### Configuration
Create a `config.env` file in the project root:
```env
# JSON File Manager Configuration
JSON_MANAGER_BASE_PATH=C:\Your\Preferred\Base\Path\
```

### Running the Application
```bash
go run main.go
```

The application will open in a new window with a modern web interface.

## 📖 How to Use

### 1. **File Discovery**
- **Base Path**: Automatically loaded from `config.env`
- **File Extension Filter**: Choose `*.json`, `*.golden`, or enter custom extensions
- **JSON Key Filter**: Enter a key name to find files containing that property (searches deep)
- **Search**: Click "🔍 Search Files" to discover files

### 2. **Browse Results**
- **Tree View**: Hierarchical display of folders and files
- **File Count**: Shows total matching files
- **View Content**: Click "View" to see beautifully formatted JSON with syntax highlighting

### 3. **Add JSON Item Operations**
- **Access**: Use the "Add to Selected", "Add after Selected", or "Replace Key" buttons
- **Object Path**: Auto-filled from your JSON key filter (supports context paths)
- **Add Properties**: Specify key name and JSON value to add to all filtered files
- **Context-Aware**: Use simple paths like "address" instead of full paths like "user.profile.address"
- **Structure Preservation**: Original file formatting and key order maintained

## 🎯 Advanced Usage Examples

### Example 1: Find and Update User Configurations
```
1. JSON Key Filter: "config"
2. Search → Shows all files with "config" objects
3. Add to Selected → Auto-fills "config" as object path
4. Add: "version": "2.0" → Updates all config objects
```

### Example 2: Update Nested Address Information  
```
1. JSON Key Filter: "address"  
2. Search → Finds files with address objects anywhere
3. Add to Selected → Auto-fills "address" as context path
4. Add: "country": "USA" → Adds to all address objects regardless of nesting
```

### Example 3: Custom File Types
```
1. File Extension: "*.golden"
2. JSON Key Filter: "testData"
3. Search → Finds .golden files containing testData
4. Add to Selected → Add test metadata to all matching files
```

### Example 4: Replace Keys Across Files
```
1. JSON Key Filter: "firstName"
2. Search → Shows all files containing "firstName" key
3. Replace Key → Auto-fills "firstName" as old key
4. New Key: "first_name" → Renames all occurrences using text replacement
```

## 📝 JSON Value Format

When adding values in add JSON item operations, use proper JSON formatting:

| Type | Example |
|------|---------|
| **String** | `"Hello World"` |
| **Number** | `123` or `45.67` |
| **Boolean** | `true` or `false` |
| **Array** | `[1, 2, 3]` or `["a", "b", "c"]` |
| **Object** | `{"name": "John", "age": 30}` |
| **Null** | `null` |

## 💡 Add JSON Item Examples

### Simple Property Addition
- **Object Path**: `""` (root level)
- **Key**: `"lastUpdated"`
- **Value**: `"2024-01-15T10:30:00Z"`

### Nested Object Update
- **Object Path**: `"user"` (context path)
- **Key**: `"status"`
- **Value**: `"active"`

### Array of Objects Update
- **Object Path**: `"users"` (array containing objects)
- **Key**: `"isActive"`
- **Value**: `true`
- **Result**: Adds `"isActive": true` as the first property in each user object
- **Note**: Will fail if any object already contains the key

### Array of Values Update
- **Object Path**: `"tags"` (array containing strings/numbers)
- **Key**: Not used for value arrays
- **Value**: `"urgent"`
- **Result**: Adds `"urgent"` as the first element in the tags array
- **Note**: Allows duplicate values (no validation)

### Complex Data Addition
- **Object Path**: `"metadata"`
- **Key**: `"tags"`
- **Value**: `["production", "verified", "v2.0"]`

## ➕ Insert After Object Examples

### Simple Object Insertion
- **Target Object Key**: `"user"` (existing object to insert after)
- **New Object Key**: `"settings"`
- **New Object JSON**: `{"theme": "dark", "notifications": true}`
- **Result**: Adds complete settings object after the user object

### Complex Object Insertion
- **Target Object Key**: `"database"`
- **New Object Key**: `"cache"`
- **New Object JSON**: `{"redis": {"host": "localhost", "port": 6379}, "ttl": 3600}`
- **Result**: Inserts a complete cache configuration object after database config

## 🔄 Replace Key Examples

### Simple Key Replacement
- **Old Key Name**: `"firstName"` (auto-filled from JSON key filter)
- **New Key Name**: `"first_name"`
- **Result**: Renames all `"firstName":` to `"first_name":` using text replacement

### Batch Key Standardization
- **Old Key Name**: `"user_id"`
- **New Key Name**: `"userId"`
- **Result**: Updates naming convention across all selected files

### Configuration Key Updates
- **Old Key Name**: `"db_host"`
- **New Key Name**: `"database_host"`
- **Result**: Modernizes configuration key names while preserving all formatting

## 🔒 Duplicate Key Prevention

The application automatically prevents duplicate keys to maintain JSON integrity:

### ✅ **Protected Operations:**
- **Root Level**: Won't add keys that already exist at the root
- **Nested Objects**: Won't add keys that already exist in target objects  
- **Array Objects**: Won't add keys if ANY object in the array already has that key

### ⚠️ **Error Messages:**
- `"key 'status' already exists at root level"`
- `"key 'isActive' already exists in object 'user'"`
- `"key 'priority' already exists in one or more objects within array 'tasks'"`

### 🔄 **Allowed Operations:**
- **Array Values**: Duplicate values are allowed in value arrays
- **Different Contexts**: Same key name can exist in different objects/arrays

## 🏗️ Project Structure

```
goldenMagic/
├── main.go                 # Core Go application with JSON processing logic
├── go.mod                  # Go module dependencies
├── config.env              # Environment configuration file
├── frontend/
│   ├── index.html         # Main web interface
│   ├── css/
│   │   └── styles.css     # Application styling
│   └── js/
│       └── app.js         # Frontend JavaScript logic
└── README.md              # Documentation
```

## 🔧 Technical Architecture

- **Backend**: Go with Lorca framework for cross-platform desktop application
- **Frontend**: Modern HTML5, CSS3, and JavaScript with embedded file serving
- **JSON Processing**: Advanced string-based manipulation preserving file structure
- **Deep Search**: Recursive JSON key discovery at any nesting level
- **File Operations**: Efficient tree-based folder scanning with filtering
- **Add JSON Item Operations**: Bulk file processing with individual error tracking

## 📦 Dependencies

```go
require (
    github.com/joho/godotenv v1.5.1    // Environment file loading
    github.com/zserge/lorca v0.1.10    // Cross-platform desktop UI
)
```

## 🏗️ Building

### Development
```bash
go run main.go
```

### Production Build
```bash
go build -o golden-magic.exe main.go
```

## 🔧 Troubleshooting

| Issue | Solution |
|-------|----------|
| **Application won't start** | Ensure Chrome/Chromium is installed |
| **Config not loading** | Check `config.env` file exists and has correct path |
| **No files found** | Verify base path exists and contains target files |
| **Add JSON item fails** | Check JSON value syntax and object path validity |
| **Permission errors** | Ensure read/write access to target directories |

## 🎯 Use Cases

- **Test Data Management**: Update test configurations across multiple files
- **Configuration Updates**: Bulk modify application settings
- **Data Migration**: Add new fields to existing JSON datasets  
- **Quality Assurance**: Find and verify specific data structures
- **Development Tools**: Manage JSON-based project configurations

## License

This project is open source and available under the MIT License.

## Contributing

Feel free to submit issues, feature requests, or pull requests to improve the application. 