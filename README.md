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

### 🚀 **Mass JSON Operations**
- **Bulk JSON Editing**: Add properties to all filtered files at once
- **Insert After Object**: Add complete JSON objects after specified target objects with duplicate detection
- **Replace Keys**: Rename JSON keys across multiple files using text replacement
- **Context-Aware Paths**: Smart object path detection and auto-completion
- **Structure Preservation**: Maintains original file formatting and key order
- **Progress Tracking**: Real-time feedback on bulk operations
- **Error Handling**: Detailed success/failure reporting per file
- **Duplicate Prevention**: Automatically prevents adding duplicate keys to maintain JSON integrity

### 🎯 **Smart Filtering**
- **File Extension Filter**: Support for *.json, *.golden, and custom extensions
- **JSON Key Filter**: Find files containing specific nested properties
- **Combined Filtering**: Use both filters simultaneously for precise results

## 📦 Installation Options

### Option 1: Pre-built Executables (Recommended)

Download the appropriate executable from the `exec/` directory:

#### Windows
- **File**: `goldenMagic.exe`
- **Usage**: Double-click to run or execute from command prompt
- **Requirements**: Windows 10 or later, Chrome/Chromium browser

#### macOS Intel (x86_64)
- **File**: `goldenMagic-macos-amd64`
- **For**: Intel-based Macs
- **Setup**:
  ```bash
  chmod +x goldenMagic-macos-amd64
  ./goldenMagic-macos-amd64
  ```

#### macOS Apple Silicon (ARM64)
- **File**: `goldenMagic-macos-arm64`
- **For**: Apple Silicon Macs (M1, M2, M3+)
- **Setup**:
  ```bash
  chmod +x goldenMagic-macos-arm64
  ./goldenMagic-macos-arm64
  ```

**Note**: macOS may show a security warning for unsigned applications. Right-click the file and select "Open" to bypass this, or allow it in System Preferences > Security & Privacy.

### Option 2: Build from Source

#### Prerequisites
- Go 1.21 or higher
- Chrome/Chromium browser (required by Lorca)

#### Steps
1. Clone or download this repository
2. Navigate to the project directory
3. Install dependencies:
   ```bash
   go mod tidy
   ```
4. Run or build:
   ```bash
   # Run directly
   go run main.go
   
   # Or build executable
   go build -o goldenMagic.exe .
   ```

## ⚙️ Configuration Setup

### Creating config.env

Create a `config.env` file in the same directory as your executable (or project root if building from source):

```env
# JSON File Manager Configuration
# Specify one or more base paths separated by semicolons (Windows) or colons (macOS/Linux)

# Single path example:
JSON_MANAGER_BASE_PATH=C:\Your\Project\Directory\

# Multiple paths example (Windows):
JSON_MANAGER_BASE_PATH=C:\Project1\;C:\Project2\;D:\TestData\

# Multiple paths example (macOS/Linux):
JSON_MANAGER_BASE_PATH=/Users/username/project1:/Users/username/project2:/opt/testdata

# Additional configuration options:
# JSON_MANAGER_MAX_FILE_SIZE=10485760  # Max file size in bytes (default: 10MB)
# JSON_MANAGER_TIMEOUT=30              # Operation timeout in seconds (default: 30)
```

### Configuration Details

| Setting | Description | Default | Example |
|---------|-------------|---------|---------|
| `JSON_MANAGER_BASE_PATH` | Base directories to search for JSON files | None (required) | `C:\Projects\` |
| `JSON_MANAGER_MAX_FILE_SIZE` | Maximum file size to process (bytes) | 10485760 (10MB) | `5242880` |
| `JSON_MANAGER_TIMEOUT` | Operation timeout in seconds | 30 | `60` |

### Path Configuration Tips

- **Windows**: Use backslashes `\` and separate multiple paths with semicolons `;`
- **macOS/Linux**: Use forward slashes `/` and separate multiple paths with colons `:`
- **Relative Paths**: Supported, relative to executable location
- **Network Paths**: Supported on Windows (e.g., `\\server\share\`)
- **Validation**: Invalid paths are automatically filtered out and logged

## 🚀 Quick Start

1. **Download** the appropriate executable for your platform from `exec/`
2. **Create** `config.env` in the same directory as the executable
3. **Configure** your base paths in `config.env`
4. **Run** the executable - it will open in your default browser
5. **Search** for JSON files using the filters
6. **Edit** files individually or perform bulk operations

## 📖 How to Use

### 1. **File Discovery**
- **Base Paths**: Automatically loaded from `config.env` (multiple paths supported)
- **File Extension Filter**: Choose `*.json`, `*.golden`, or enter custom extensions
- **JSON Key Filter**: Enter a key name to find files containing that property (searches deep)
- **Search**: Click "🔍 Search Files" to discover files across all configured paths

### 2. **Browse Results**
- **Tree View**: Hierarchical display of folders and files from all base paths
- **File Count**: Shows total matching files across all locations
- **Base Path Indicators**: Icons show which base path each file comes from
- **View Content**: Click file names to see beautifully formatted JSON with syntax highlighting

### 3. **Mass JSON Operations**
- **Access**: Use the "Add to Selected", "Add after Selected", or "Replace Key" buttons
- **Object Path**: Auto-filled from your JSON key filter (supports context paths)
- **Add Properties**: Specify key name and JSON value to add to all filtered files
- **Duplicate Detection**: Automatically prevents adding existing keys
- **Context-Aware**: Use simple paths like "address" instead of full paths like "user.profile.address"
- **Structure Preservation**: Original file formatting and key order maintained

## 🎯 Advanced Usage Examples

### Example 1: Find and Update User Configurations
```
1. JSON Key Filter: "config"
2. Search → Shows all files with "config" objects across all base paths
3. Add to Selected → Auto-fills "config" as object path
4. Add: "version": "2.0" → Updates all config objects
```

### Example 2: Update Nested Address Information  
```
1. JSON Key Filter: "address"  
2. Search → Finds files with address objects anywhere in all configured directories
3. Add to Selected → Auto-fills "address" as context path
4. Add: "country": "USA" → Adds to all address objects regardless of nesting
```

### Example 3: Custom File Types Across Multiple Projects
```
1. Configure multiple base paths in config.env
2. File Extension: "*.golden"
3. JSON Key Filter: "testData"
4. Search → Finds .golden files containing testData across all projects
5. Add to Selected → Add test metadata to all matching files
```

### Example 4: Replace Keys Across Multiple Codebases
```
1. JSON Key Filter: "firstName"
2. Search → Shows all files containing "firstName" key across all base paths
3. Replace Key → Auto-fills "firstName" as old key
4. New Key: "first_name" → Renames all occurrences using text replacement
```

## 📝 JSON Value Format

When adding values in mass JSON operations, use proper JSON formatting:

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
- **Duplicate Check**: Will skip if "settings" key already exists in the same object

### Complex Object Insertion
- **Target Object Key**: `"database"`
- **New Object Key**: `"cache"`
- **New Object JSON**: `{"redis": {"host": "localhost", "port": 6379}, "ttl": 3600}`
- **Result**: Inserts a complete cache configuration object after database config
- **Duplicate Check**: Prevents adding "cache" if it already exists at the same level

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
- **Insert After**: Won't add objects with keys that already exist at the same level
- **Root Level**: Won't add keys that already exist at the root
- **Nested Objects**: Won't add keys that already exist in target objects  
- **Array Objects**: Won't add keys if ANY object in the array already has that key

### ⚠️ **Feedback Messages:**
- `"SKIPPED: object with key 'settings' already exists"`
- `"SKIPPED: key 'status' already exists at root level"`
- `"SKIPPED: key 'isActive' already exists in object 'user'"`
- `"SKIPPED: key 'priority' already exists in one or more objects within array 'tasks'"`

### 🔄 **Allowed Operations:**
- **Array Values**: Duplicate values are allowed in value arrays
- **Different Contexts**: Same key name can exist in different objects/arrays
- **Replace Keys**: Text replacement operations don't check for duplicates

## 🏗️ Project Structure

```
goldenMagic/
├── exec/                           # Pre-built executables
│   ├── goldenMagic.exe            # Windows executable
│   ├── goldenMagic-macos-amd64    # macOS Intel executable  
│   └── goldenMagic-macos-arm64    # macOS Apple Silicon executable
├── main.go                        # Core Go application
├── go.mod                         # Go module dependencies
├── config.env                     # Environment configuration (create this)
├── internal/                      # Internal Go packages
│   ├── config/                    # Configuration management
│   ├── fileops/                   # File operations
│   ├── jsonops/                   # JSON manipulation
│   └── tree/                      # Tree structure building
├── frontend/                      # Web interface files
│   ├── index.html                # Main web interface
│   ├── css/
│   │   └── styles.css            # Application styling
│   └── js/
│       └── app.js                # Frontend JavaScript logic
└── README.md                     # This documentation
```

## 🔧 Technical Architecture

- **Backend**: Go with Lorca framework for cross-platform desktop application
- **Frontend**: Modern HTML5, CSS3, and JavaScript with embedded file serving
- **JSON Processing**: Advanced string-based manipulation preserving file structure
- **Deep Search**: Recursive JSON key discovery at any nesting level
- **File Operations**: Efficient tree-based folder scanning with filtering
- **Mass Operations**: Bulk file processing with individual error tracking and duplicate prevention
- **Multi-Path Support**: Simultaneous searching across multiple base directories

## 📦 Dependencies

```go
require (
    github.com/joho/godotenv v1.5.1    // Environment file loading
    github.com/zserge/lorca v0.1.10    // Cross-platform desktop UI
)
```

## 🔧 Troubleshooting

| Issue | Solution |
|-------|----------|
| **Application won't start** | Ensure Chrome/Chromium is installed |
| **Config not loading** | Check `config.env` file exists in same directory as executable |
| **No files found** | Verify base paths in `config.env` exist and contain target files |
| **Permission denied** | Run executable with appropriate permissions |
| **Mass operation fails** | Check JSON value syntax and object path validity |
| **macOS security warning** | Right-click executable and select "Open", or allow in Security & Privacy |
| **Path not found** | Use absolute paths in `config.env` or check path separators (`;` for Windows, `:` for macOS/Linux) |

## 🎯 Use Cases

- **Multi-Project Management**: Search and update JSON files across multiple codebases simultaneously
- **Test Data Management**: Update test configurations across multiple projects
- **Configuration Updates**: Bulk modify application settings with duplicate prevention
- **Data Migration**: Add new fields to existing JSON datasets safely
- **Quality Assurance**: Find and verify specific data structures across projects
- **Development Tools**: Manage JSON-based project configurations at scale

## License

This project is open source and available under the MIT License.

## Contributing

Feel free to submit issues, feature requests, or pull requests to improve the application. 