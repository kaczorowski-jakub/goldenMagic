// Global variables
let currentFileTree = null;
let allFiles = [];
let searchTimeout = null; // For debouncing

// Debounce utility function
function debounce(func, wait) {
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(searchTimeout);
            func(...args);
        };
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(later, wait);
    };
}

// Enhanced error handling
function handleError(error, context = 'Operation') {
    console.error(`${context} error:`, error);
    
    let errorMessage = error.message || 'Unknown error occurred';
    
    // Provide more user-friendly error messages
    if (errorMessage.includes('file too large')) {
        errorMessage = '📁 File is too large to process (max 10MB)';
    } else if (errorMessage.includes('invalid JSON')) {
        errorMessage = '❌ Invalid JSON format detected';
    } else if (errorMessage.includes('no valid base paths')) {
        errorMessage = '📂 No valid directories found in configuration';
    } else if (errorMessage.includes('target key') && errorMessage.includes('not found')) {
        errorMessage = '🔍 Target key not found in selected files';
    }
    
    showMessage(`❌ ${context}: ${errorMessage}`, 'error');
}

// Initialize the application
async function initializeApp() {
    try {
        // Load and display base paths
        await loadBasePaths();
        
        // Set up event listeners with debouncing
        setupEventListeners();
        
        showMessage('✅ Application initialized successfully', 'success');
    } catch (error) {
        handleError(error, 'Initialization failed');
    }
}

// Enhanced search with debouncing
const debouncedSearch = debounce(async () => {
    try {
        await searchFiles();
    } catch (error) {
        handleError(error, 'Search failed');
    }
}, 500); // 500ms delay

// Load and display base paths
async function loadBasePaths() {
    try {
        const basePaths = await window.getBasePaths();
        displayBasePaths(basePaths);
    } catch (error) {
        handleError(error, 'Loading base paths failed');
    }
}

// Display base paths in the UI
function displayBasePaths(basePaths) {
    const pathsContainer = document.getElementById('base-paths-list');
    const pathsCount = document.getElementById('paths-count');
    
    if (!pathsContainer || !pathsCount) {
        console.error('Base paths container elements not found');
        return;
    }
    
    // Update count
    pathsCount.textContent = basePaths.length;
    
    // Clear existing content
    pathsContainer.innerHTML = '';
    
    if (basePaths.length === 0) {
        pathsContainer.innerHTML = '<div class="no-paths">No base paths configured</div>';
        return;
    }
    
    // Create path elements
    basePaths.forEach((path, index) => {
        const pathElement = document.createElement('div');
        pathElement.className = 'base-path-item';
        pathElement.innerHTML = `
            <div class="path-info">
                <span class="path-number">${index + 1}.</span>
                <span class="path-text" title="${path}">${path}</span>
            </div>
        `;
        pathsContainer.appendChild(pathElement);
    });
}

// Set up event listeners
function setupEventListeners() {
    // Search button
    const searchBtn = document.getElementById('searchBtn');
    if (searchBtn) {
        searchBtn.addEventListener('click', searchFiles);
    }
    
    // Enter key in filter inputs
    const extensionFilter = document.getElementById('fileExtension');
    const jsonKeyFilter = document.getElementById('jsonKeyFilter');
    
    if (extensionFilter) {
        extensionFilter.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                searchFiles();
            }
        });
    }
    
    if (jsonKeyFilter) {
        jsonKeyFilter.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                searchFiles();
            }
        });
    }
}

// Enhanced search function with better error handling
async function searchFiles() {
    const extensionFilter = document.getElementById('fileExtension').value.trim();
    const jsonKeyFilter = document.getElementById('jsonKeyFilter').value.trim();
    
    // Show loading state
    const searchBtn = document.getElementById('searchBtn');
    const originalText = searchBtn.textContent;
    searchBtn.textContent = '🔍 Searching...';
    searchBtn.disabled = true;
    
    try {
        showMessage('🔍 Searching files...', 'info');
        
        const fileTree = await window.browseFolder(extensionFilter, jsonKeyFilter);
        
        if (!fileTree) {
            throw new Error('No results returned from search');
        }
        
        currentFileTree = fileTree;
        allFiles = flattenFileTree(fileTree);
        
        displayFileTree(fileTree);
        
        const count = fileTree.count || 0;
        if (count === 0) {
            showMessage('📂 No files found matching your criteria', 'warning');
        } else {
            showMessage(`✅ Found ${count} file${count !== 1 ? 's' : ''} matching your criteria`, 'success');
        }
        
    } catch (error) {
        handleError(error, 'Search operation failed');
        // Clear results on error
        document.getElementById('file-tree').innerHTML = '<div class="no-results">Search failed. Please try again.</div>';
    } finally {
        // Reset button state
        searchBtn.textContent = originalText;
        searchBtn.disabled = false;
    }
}

// Display file tree with multiple paths support
function displayFileTree(tree) {
    const resultsContainer = document.getElementById('results');
    
    if (!tree || tree.count === 0) {
        resultsContainer.innerHTML = `
            <div class="no-results">
                <h3>No Results Found</h3>
                <p>Try adjusting your search filters or check your base paths configuration.</p>
            </div>
        `;
        return;
    }
    
    // Create tree header with controls
    const headerHTML = `
        <div class="tree-header">
            <h3>${tree.name} (${tree.count} files)</h3>
            <div class="header-controls">
                <div class="selection-controls">
                    <label class="checkbox-label">
                        <input type="checkbox" id="select-all-files" onchange="toggleAllFiles()">
                        Select All (<span id="selected-count">0</span>)
                    </label>
                </div>
                <div class="action-buttons">
                    <button id="add-json-item-to-btn" class="action-btn add-operation" onclick="toggleAddJSONItemToForm()">
                        ➕ Add to Selected
                    </button>
                    <button id="insert-after-btn" class="action-btn insert-operation" onclick="toggleInsertAfterForm()">
                        📝 Add after Selected
                    </button>
                    <button id="replace-key-btn" class="action-btn replace-operation" onclick="toggleReplaceKeyForm()">
                        🔄 Replace Key
                    </button>
                </div>
            </div>
        </div>
        <div id="add-json-item-to-form" class="add-json-item-to-form" style="display: none;">
            <h4>➕ Add Property to Selected Files</h4>
            <div class="form-row">
                <input type="text" id="add-json-object-path" placeholder="Object path (e.g., user.address or leave empty for root)" />
                <input type="text" id="add-json-key" placeholder="Key name" />
            </div>
            <p class="form-help">💡 This will add the property as the FIRST item in the target objects.</p>
            <textarea id="add-json-value" placeholder="Value (JSON format, e.g., &quot;string&quot;, 123, {&quot;nested&quot;: true})"></textarea>
            <div class="add-json-item-to-buttons">
                <button id="perform-add-json-item-to" class="btn btn-primary">➕ Add to Selected Files</button>
                <button id="cancel-add-json-item-to" class="btn">Cancel</button>
            </div>
        </div>
        <div id="insert-after-form" class="add-json-item-to-form" style="display: none;">
            <h4>📝 Add Object After Target</h4>
            <div class="form-row">
                <input type="text" id="target-object-key" placeholder="Target object key (to insert after)" />
                <input type="text" id="new-object-key" placeholder="New object key name" />
            </div>
            <p class="form-help">💡 This will add a complete JSON object AFTER ALL OCCURRENCES of the specified target key. If the target key appears multiple times in a file, the new object will be added after each occurrence.</p>
            <textarea id="new-object-json" placeholder="New object JSON (e.g., {&quot;name&quot;: &quot;value&quot;, &quot;nested&quot;: {&quot;key&quot;: true}})"></textarea>
            <div class="add-json-item-to-buttons">
                <button id="perform-insert-after" class="btn btn-primary">📝 Add After Target</button>
                <button id="cancel-insert-after" class="btn">Cancel</button>
            </div>
        </div>
        <div id="replace-key-form" class="add-json-item-to-form" style="display: none;">
            <h4>🔄 Replace Key in Selected Files</h4>
            <div class="form-row">
                <input type="text" id="old-key-name" placeholder="Old key name (to be replaced)" />
                <input type="text" id="new-key-name" placeholder="New key name (replacement)" />
            </div>
            <p class="form-help">💡 This will rename all occurrences of the old key to the new key in the selected files using simple text replacement.</p>
            <div class="add-json-item-to-buttons">
                <button id="perform-replace-key" class="btn btn-primary">🔄 Replace Key</button>
                <button id="cancel-replace-key" class="btn">Cancel</button>
            </div>
        </div>
    `;
    
    // Create tree content
    const treeHTML = renderTreeNode(tree, 0);
    
    resultsContainer.innerHTML = headerHTML + '<div class="tree-container">' + treeHTML + '</div>';
    
    // Set up form event listeners after HTML is added
    setupFormEventListeners();
    
    // Set up inline copy button listeners
    setupInlineCopyListeners();
    
    // Update selection count
    updateSelectionCount();
}

// Set up inline copy button listeners using event delegation
function setupInlineCopyListeners() {
    // Remove any existing listeners to avoid duplicates
    document.removeEventListener('click', handleInlineCopyClick);
    
    // Add event listener for inline copy buttons
    document.addEventListener('click', handleInlineCopyClick);
}

// Handle inline copy button clicks
function handleInlineCopyClick(event) {
    if (event.target.matches('.inline-copy-btn[data-file-path]') || 
        event.target.closest('.inline-copy-btn[data-file-path]')) {
        
        const button = event.target.matches('.inline-copy-btn[data-file-path]') ? 
                      event.target : 
                      event.target.closest('.inline-copy-btn[data-file-path]');
        
        const filePath = button.getAttribute('data-file-path');
        if (filePath) {
            event.preventDefault();
            event.stopPropagation();
            copyJsonToClipboard(filePath);
        }
    }
}

// Set up form event listeners
function setupFormEventListeners() {
    const performAddJSONItemToBtn = document.getElementById('perform-add-json-item-to');
    if (performAddJSONItemToBtn) {
        performAddJSONItemToBtn.addEventListener('click', performAddJSONItemTo);
    }
    
    const performInsertAfterBtn = document.getElementById('perform-insert-after');
    if (performInsertAfterBtn) {
        performInsertAfterBtn.addEventListener('click', performInsertAfter);
    }
    
    const cancelAddJSONItemToBtn = document.getElementById('cancel-add-json-item-to');
    if (cancelAddJSONItemToBtn) {
        cancelAddJSONItemToBtn.addEventListener('click', toggleAddJSONItemToForm);
    }
    
    const cancelInsertAfterBtn = document.getElementById('cancel-insert-after');
    if (cancelInsertAfterBtn) {
        cancelInsertAfterBtn.addEventListener('click', toggleInsertAfterForm);
    }
    
    const performReplaceKeyBtn = document.getElementById('perform-replace-key');
    if (performReplaceKeyBtn) {
        performReplaceKeyBtn.addEventListener('click', performReplaceKey);
    }
    
    const cancelReplaceKeyBtn = document.getElementById('cancel-replace-key');
    if (cancelReplaceKeyBtn) {
        cancelReplaceKeyBtn.addEventListener('click', toggleReplaceKeyForm);
    }
}

// Render a tree node (supports multiple base paths)
function renderTreeNode(node, depth) {
    const indent = '  '.repeat(depth);
    let html = '';
    
    if (node.isDir) {
        // Directory node
        const hasFiles = node.files && node.files.length > 0;
        const hasChildren = node.children && node.children.length > 0;
        const isExpanded = depth < 2; // Auto-expand first two levels
        
        html += `
            <div class="tree-node directory" style="margin-left: ${depth * 20}px">
                <div class="directory-header" onclick="toggleDirectory(this)">
                    <span class="toggle-icon ${isExpanded ? 'expanded' : ''}">${isExpanded ? '▼' : '▶'}</span>
                    <span class="directory-name">📁 ${node.name}</span>
                    <span class="file-count">(${node.count} files)</span>
                    ${node.basePath ? '<span class="base-path-indicator" title="Base Path: ' + node.basePath + '">🏠</span>' : ''}
                </div>
                <div class="directory-content" style="display: ${isExpanded ? 'block' : 'none'}">
        `;
        
        // Render files in this directory
        if (hasFiles) {
            node.files.forEach(file => {
                const fileId = 'file-' + btoa(file.path).replace(/[^a-zA-Z0-9]/g, ''); // Create safe ID
                html += `
                    <div class="tree-node file" style="margin-left: ${(depth + 1) * 20}px">
                        <div class="file-item">
                            <label class="file-checkbox">
                                <input type="checkbox" value="${file.path}" onchange="updateSelectionCount()">
                            </label>
                            <span class="file-name" onclick='loadFileContentInline(${JSON.stringify(file.path)}, "${fileId}")' title="Click to view content" style="cursor: pointer;">
                                📄 ${file.name}
                            </span>
                            <span class="file-path" title="${file.path}">${file.path}</span>
                            ${file.basePath ? '<span class="file-base-path" title="From: ' + file.basePath + '">📂</span>' : ''}
                        </div>
                        <div id="${fileId}" class="inline-file-content" style="display: none; margin-left: 20px; margin-top: 10px; border-left: 3px solid #3b82f6; padding-left: 15px; background: #f8fafc;"></div>
                    </div>
                `;
            });
        }
        
        // Render subdirectories
        if (hasChildren) {
            node.children.forEach(child => {
                html += renderTreeNode(child, depth + 1);
            });
        }
        
        html += `
                </div>
            </div>
        `;
    }
    
    return html;
}

// Flatten file tree to get all files
function flattenFileTree(node) {
    let files = [];
    
    if (!node) return files;
    
    // Add files from current node
    if (node.files) {
        files = files.concat(node.files);
    }
    
    // Recursively add files from children
    if (node.children) {
        node.children.forEach(child => {
            files = files.concat(flattenFileTree(child));
        });
    }
    
    return files;
}

// Toggle directory expansion
function toggleDirectory(element) {
    const toggleIcon = element.querySelector('.toggle-icon');
    const content = element.parentElement.querySelector('.directory-content');
    
    if (content.style.display === 'none') {
        content.style.display = 'block';
        toggleIcon.textContent = '▼';
        toggleIcon.classList.add('expanded');
    } else {
        content.style.display = 'none';
        toggleIcon.textContent = '▶';
        toggleIcon.classList.remove('expanded');
    }
}

// Toggle all files selection
function toggleAllFiles() {
    const selectAllCheckbox = document.getElementById('select-all-files');
    const fileCheckboxes = document.querySelectorAll('.file-checkbox input[type="checkbox"]');
    
    fileCheckboxes.forEach(checkbox => {
        checkbox.checked = selectAllCheckbox.checked;
    });
    
    updateSelectionCount();
}

// Update selection count display
function updateSelectionCount() {
    const selectedCheckboxes = document.querySelectorAll('.file-checkbox input[type="checkbox"]:checked');
    const selectedCount = selectedCheckboxes.length;
    const totalCount = document.querySelectorAll('.file-checkbox input[type="checkbox"]').length;
    
    const selectedCountElement = document.getElementById('selected-count');
    if (selectedCountElement) {
        selectedCountElement.textContent = selectedCount;
    }
    
    const selectAllCheckbox = document.getElementById('select-all-files');
    if (selectAllCheckbox) {
        selectAllCheckbox.checked = selectedCount === totalCount && totalCount > 0;
        selectAllCheckbox.indeterminate = selectedCount > 0 && selectedCount < totalCount;
    }
}

// Get selected files
function getSelectedFiles() {
    const selectedCheckboxes = document.querySelectorAll('.file-checkbox input[type="checkbox"]:checked');
    return Array.from(selectedCheckboxes).map(checkbox => {
        const filePath = checkbox.value;
        const fileName = checkbox.closest('.file-item').querySelector('.file-name').textContent.replace('📄 ', '');
        return { path: filePath, name: fileName };
    });
}

// Load and display file content inline below the file item
async function loadFileContentInline(filePath, containerId) {
    const container = document.getElementById(containerId);
    if (!container) {
        console.error('Inline container not found:', containerId);
        return;
    }
    
    try {
        // Toggle visibility - if already shown, hide it
        if (container.style.display === 'block') {
            container.style.display = 'none';
            return;
        }
        
        showMessage('📖 Loading file content...', 'info');
        
        // Check if the function is available
        if (typeof window.getJSONFileContent !== 'function') {
            throw new Error('getJSONFileContent function not available. Please ensure the application is properly loaded.');
        }
        
        console.log('Loading file:', filePath);
        
        const content = await window.getJSONFileContent(filePath);
        console.log('File content loaded successfully, length:', content.length);
        
        // Display content inline
        displayFileContentInline(filePath, content, container);
        container.style.display = 'block';
        
        showMessage('✅ File loaded successfully', 'success');
        
    } catch (error) {
        console.error('File loading error:', error);
        
        // Show error in the inline container
        container.innerHTML = `
            <div style="color: #dc2626; padding: 10px; background: #fee2e2; border-radius: 6px;">
                ❌ Error loading file: ${String(error.message || error)}
            </div>
        `;
        container.style.display = 'block';
        
        showMessage(`❌ Cannot load file: ${String(error.message || error)}`, 'error');
    }
}

// Legacy function for backward compatibility (if needed elsewhere)
async function loadFileContent(filePath) {
    try {
        showMessage('📖 Loading file content...', 'info');
        
        // Check if the function is available
        if (typeof window.getJSONFileContent !== 'function') {
            throw new Error('getJSONFileContent function not available. Please ensure the application is properly loaded.');
        }
        
        console.log('Loading file:', filePath);
        console.log('Path type:', typeof filePath);
        console.log('Path length:', filePath.length);
        console.log('Raw path characters:', filePath.split('').map(c => c.charCodeAt(0)));
        
        const content = await window.getJSONFileContent(filePath);
        console.log('File content loaded successfully, length:', content.length);
        
        displayFileContent(filePath, content);
        showMessage('✅ File loaded successfully', 'success');
        
    } catch (error) {
        console.error('File loading error:', error);
        console.error('Error details:', {
            message: error.message,
            filePath: filePath,
            functionAvailable: typeof window.getJSONFileContent
        });
        
        // Provide more specific error messages
        let errorMessage = String(error.message || error);
        if (errorMessage.includes('file too large')) {
            errorMessage = 'File is too large to load (max 10MB)';
        } else if (errorMessage.includes('invalid JSON')) {
            errorMessage = 'File contains invalid JSON format';
        } else if (errorMessage.includes('no such file')) {
            errorMessage = 'File not found or access denied';
        } else if (errorMessage.includes('not available')) {
            errorMessage = 'Application not properly initialized. Please refresh the page.';
        }
        
        showMessage(`❌ Cannot load file: ${errorMessage}`, 'error');
    }
}

// Display file content inline in the provided container
function displayFileContentInline(filePath, content, container) {
    const formattedContent = formatJsonContent(content);
    const fileName = filePath.split(/[\\\/]/).pop();
    
    container.innerHTML = `
        <div style="background: white; border-radius: 8px; padding: 15px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; border-bottom: 1px solid #e5e7eb; padding-bottom: 10px;">
                <h4 style="margin: 0; color: #374151;">📄 ${fileName}</h4>
                <button class="inline-copy-btn" data-file-path="${filePath}" 
                        style="padding: 4px 8px; background: #3b82f6; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;">
                    📋 Copy
                </button>
            </div>
            <div style="max-height: 400px; overflow-y: auto; font-family: 'Courier New', monospace; font-size: 12px;">
                ${formattedContent}
            </div>
        </div>
    `;
}

// Display file content with syntax highlighting (legacy function for separate container)
function displayFileContent(filePath, content) {
    const contentContainer = document.getElementById('file-content');
    
    if (!contentContainer) {
        console.error('File content container not found');
        return;
    }
    
    const formattedContent = formatJsonContent(content);
    
    // Show the container
    contentContainer.style.display = 'block';
    
    contentContainer.innerHTML = `
        <div class="file-content-header">
            <h3>📄 ${filePath.split(/[\\\/]/).pop()}</h3>
            <div class="file-actions">
                <button onclick="copyJsonToClipboard(${JSON.stringify(filePath)})" class="copy-btn">
                    📋 Copy to Clipboard
                </button>
            </div>
        </div>
        <div class="file-path-info">
            <strong>Path:</strong> ${filePath}
        </div>
        <div class="json-content">
            ${formattedContent}
        </div>
    `;
    
    // Scroll to content
    contentContainer.scrollIntoView({ behavior: 'smooth' });
    
    showMessage('✅ File content loaded successfully', 'success');
}

// Format JSON content with syntax highlighting and line numbers
function formatJsonContent(jsonString) {
    try {
        const parsed = JSON.parse(jsonString);
        const formatted = JSON.stringify(parsed, null, 2);
        const lines = formatted.split('\n');
        
        let result = '<div class="json-lines">';
        lines.forEach((line, index) => {
            const lineNumber = index + 1;
            const highlightedLine = highlightJsonSyntax(line);
            result += `<div class="json-line">`;
            result += `<span class="line-number">${lineNumber}</span>`;
            result += `<span class="line-content">${highlightedLine}</span>`;
            result += `</div>`;
        });
        result += '</div>';
        
        return result;
    } catch (error) {
        return `<pre class="json-error">Invalid JSON: ${error.message}</pre>`;
    }
}

// Highlight JSON syntax
function highlightJsonSyntax(line) {
    return line
        .replace(/(".*?")(\s*:)/g, '<span class="json-key">$1</span>$2')
        .replace(/:\s*(".*?")/g, ': <span class="json-string">$1</span>')
        .replace(/:\s*(true|false)/g, ': <span class="json-boolean">$1</span>')
        .replace(/:\s*(null)/g, ': <span class="json-null">$1</span>')
        .replace(/:\s*(-?\d+\.?\d*)/g, ': <span class="json-number">$1</span>')
        .replace(/([{}[\],])/g, '<span class="json-punctuation">$1</span>');
}

// Copy JSON content to clipboard
async function copyJsonToClipboard(filePath) {
    try {
        const content = await window.getJSONFileContent(filePath);
        await navigator.clipboard.writeText(content);
        showMessage('✅ Content copied to clipboard', 'success');
        
        // Find any copy buttons for this file and briefly change their text
        const copyButtons = document.querySelectorAll(`[data-file-path="${filePath}"]`);
        copyButtons.forEach(button => {
            if (button.textContent.includes('Copy')) {
                const originalText = button.textContent;
                button.textContent = '✅ Copied!';
                button.style.background = '#10b981';
                
                setTimeout(() => {
                    button.textContent = originalText;
                    button.style.background = '#3b82f6';
                }, 2000);
            }
        });
        
    } catch (error) {
        showMessage('❌ Failed to copy content: ' + error.message, 'error');
    }
}

// Add JSON item functionality
function toggleAddJSONItemToForm() {
    const form = document.getElementById('add-json-item-to-form');
    const isVisible = form.style.display === 'block';
    
    if (isVisible) {
        form.style.display = 'none';
    } else {
        form.style.display = 'block';
        
        // Auto-populate object path from JSON key filter
        const jsonKeyFilter = document.getElementById('jsonKeyFilter').value.trim();
        const objectPathInput = document.getElementById('add-json-object-path');
        if (jsonKeyFilter && objectPathInput) {
            objectPathInput.value = jsonKeyFilter;
        }
        
        // Focus on the first input
        const firstInput = form.querySelector('input');
        if (firstInput) {
            firstInput.focus();
        }
        
        // Hide insert after form if open
        const insertAfterForm = document.getElementById('insert-after-form');
        if (insertAfterForm && insertAfterForm.style.display === 'block') {
            insertAfterForm.style.display = 'none';
        }
        
        // Hide replace key form if open
        const replaceKeyForm = document.getElementById('replace-key-form');
        if (replaceKeyForm && replaceKeyForm.style.display === 'block') {
            replaceKeyForm.style.display = 'none';
        }
    }
}

async function performAddJSONItemTo() {
    const objectPath = document.getElementById('add-json-object-path').value.trim();
    const key = document.getElementById('add-json-key').value.trim();
    const valueStr = document.getElementById('add-json-value').value.trim();

    if (!key || !valueStr) {
        showMessage('❌ Please enter both key and value', 'error');
        return;
    }

    try {
        // Parse the value as JSON
        const value = JSON.parse(valueStr);
        
        // Get selected file paths
        const selectedFiles = getSelectedFiles();
        const filePaths = selectedFiles.map(file => file.path);
        
        if (filePaths.length === 0) {
            showMessage('❌ No files selected for update', 'error');
            return;
        }

        // Show progress message
        showMessage(`➕ Adding property to ${filePaths.length} files across multiple paths...`, 'info');
        
        // Call the backend function
        const results = await window.addJSONItemToFiles(filePaths, objectPath, key, value);
        
        // Process results
        let successCount = 0;
        let errorCount = 0;
        const errors = [];
        
        for (const [filePath, result] of Object.entries(results)) {
            if (result === 'SUCCESS') {
                successCount++;
            } else {
                errorCount++;
                errors.push(`${filePath}: ${result}`);
            }
        }
        
        // Show results
        if (errorCount === 0) {
            showMessage(`✅ Successfully added "${key}" to ${successCount} files`, 'success');
        } else {
            showMessage(`⚠️ Added to ${successCount} files, ${errorCount} failed. Check console for details.`, 'error');
            console.error('Add property errors:', errors);
        }
        
        // Clear and close form
        document.getElementById('add-json-object-path').value = '';
        document.getElementById('add-json-key').value = '';
        document.getElementById('add-json-value').value = '';
        toggleAddJSONItemToForm();
        
        // Clear the search filter inputs to allow for a fresh search
        document.getElementById('fileExtension').value = '';
        document.getElementById('jsonKeyFilter').value = '';
        
    } catch (error) {
        showMessage('❌ Error during adding JSON Item to: ' + error.message, 'error');
    }
}

// Insert after functionality
function toggleInsertAfterForm() {
    const form = document.getElementById('insert-after-form');
    const isVisible = form.style.display === 'block';
    
    if (isVisible) {
        form.style.display = 'none';
    } else {
        form.style.display = 'block';
        
        // Auto-populate target key from JSON key filter
        const jsonKeyFilter = document.getElementById('jsonKeyFilter').value.trim();
        const targetKeyInput = document.getElementById('target-object-key');
        if (jsonKeyFilter && targetKeyInput) {
            targetKeyInput.value = jsonKeyFilter;
        }
        
        // Focus on the first input
        const firstInput = form.querySelector('input');
        if (firstInput) {
            firstInput.focus();
        }
        
        // Hide add JSON item form if open
        const addJSONItemToForm = document.getElementById('add-json-item-to-form');
        if (addJSONItemToForm && addJSONItemToForm.style.display === 'block') {
            addJSONItemToForm.style.display = 'none';
        }
        
        // Hide replace key form if open
        const replaceKeyForm = document.getElementById('replace-key-form');
        if (replaceKeyForm && replaceKeyForm.style.display === 'block') {
            replaceKeyForm.style.display = 'none';
        }
    }
}

function toggleReplaceKeyForm() {
    const form = document.getElementById('replace-key-form');
    const isVisible = form.style.display === 'block';
    
    if (isVisible) {
        form.style.display = 'none';
    } else {
        form.style.display = 'block';
        
        // Auto-populate old key name from JSON key filter
        const jsonKeyFilter = document.getElementById('jsonKeyFilter').value.trim();
        const oldKeyInput = document.getElementById('old-key-name');
        if (jsonKeyFilter && oldKeyInput) {
            oldKeyInput.value = jsonKeyFilter;
        }
        
        // Focus on the first input
        const firstInput = form.querySelector('input');
        if (firstInput) {
            firstInput.focus();
        }
        
        // Hide add JSON item form if open
        const addJSONItemToForm = document.getElementById('add-json-item-to-form');
        if (addJSONItemToForm && addJSONItemToForm.style.display === 'block') {
            addJSONItemToForm.style.display = 'none';
        }
        
        // Hide insert after form if open
        const insertAfterForm = document.getElementById('insert-after-form');
        if (insertAfterForm && insertAfterForm.style.display === 'block') {
            insertAfterForm.style.display = 'none';
        }
    }
}

async function performReplaceKey() {
    const oldKeyName = document.getElementById('old-key-name').value.trim();
    const newKeyName = document.getElementById('new-key-name').value.trim();

    if (!oldKeyName || !newKeyName) {
        showMessage('❌ Please enter both old and new key names', 'error');
        return;
    }

    if (oldKeyName === newKeyName) {
        showMessage('❌ Old key and new key cannot be the same', 'error');
        return;
    }

    try {
        // Get selected file paths
        const selectedFiles = getSelectedFiles();
        if (selectedFiles.length === 0) {
            showMessage('❌ Please select at least one file', 'error');
            return;
        }

        const filePaths = selectedFiles.map(file => file.path);

        // Show progress message
        showMessage(`🔄 Replacing "${oldKeyName}" with "${newKeyName}" in ${filePaths.length} files...`, 'info');
        
        // Call the backend function
        const results = await window.replaceKeys(oldKeyName, newKeyName, filePaths);
        
        // Process results
        let successCount = 0;
        let errorCount = 0;
        let totalReplacements = 0;
        const errors = [];
        const successDetails = [];

        for (const result of results) {
            if (result.success) {
                successCount++;
                totalReplacements += result.replacementCount;
                successDetails.push({
                    filePath: result.filePath,
                    replacements: result.replacementCount
                });
            } else {
                errorCount++;
                errors.push({ 
                    filePath: result.filePath, 
                    error: result.error 
                });
            }
        }

        // Show results with more detailed feedback
        if (errorCount === 0) {
            showMessage(`✅ Successfully replaced "${oldKeyName}" with "${newKeyName}" in ${successCount} files (${totalReplacements} total replacements)`, 'success');
            console.log('Successfully processed files:', successDetails);
        } else {
            showMessage(`⚠️ Replaced in ${successCount} files, ${errorCount} failed. Check console for details.`, 'error');
            console.error('Replace key errors:', errors);
            if (successCount > 0) {
                console.log('Successfully processed files:', successDetails);
            }
        }
        
        // Clear and close form
        document.getElementById('old-key-name').value = '';
        document.getElementById('new-key-name').value = '';
        toggleReplaceKeyForm();
        
        // Clear the search filter inputs to allow for a fresh search
        document.getElementById('fileExtension').value = '';
        document.getElementById('jsonKeyFilter').value = '';
        
    } catch (error) {
        console.error('Error in performReplaceKey:', error);
        showMessage('❌ Error during replace key: ' + error.message, 'error');
    }
}

async function performInsertAfter() {
    const targetKey = document.getElementById('target-object-key').value.trim();
    const newObjectKey = document.getElementById('new-object-key').value.trim();
    const newObjectJSON = document.getElementById('new-object-json').value.trim();

    if (!targetKey || !newObjectKey || !newObjectJSON) {
        showMessage('❌ Please fill in all fields', 'error');
        return;
    }

    try {
        // Validate JSON
        JSON.parse(newObjectJSON);
        
        // Get selected file paths
        const selectedFiles = getSelectedFiles();
        const filePaths = selectedFiles.map(file => file.path);
        
        if (filePaths.length === 0) {
            showMessage('❌ Please select at least one file', 'error');
            return;
        }

        // Show progress message
        showMessage(`➕ Adding "${newObjectKey}" after all occurrences of "${targetKey}" in ${filePaths.length} files...`, 'info');
        
        // Call the backend function
        const results = await window.addJSONItemAfter(filePaths, targetKey, newObjectKey, newObjectJSON);
        
        // Process results
        let successCount = 0;
        let errorCount = 0;
        const errors = [];
        const successDetails = [];

        for (const [filePath, result] of Object.entries(results)) {
            if (result === 'SUCCESS') {
                successCount++;
                successDetails.push(filePath);
            } else {
                errorCount++;
                errors.push({ filePath, error: result });
            }
        }

        // Show results with more detailed feedback
        if (errorCount === 0) {
            showMessage(`✅ Successfully added "${newObjectKey}" after all occurrences of "${targetKey}" in ${successCount} files`, 'success');
            console.log('Successfully processed files:', successDetails);
        } else {
            showMessage(`⚠️ Added to ${successCount} files, ${errorCount} failed. Check console for details.`, 'error');
            console.error('Insert after errors:', errors);
            if (successCount > 0) {
                console.log('Successfully processed files:', successDetails);
            }
        }
        
        // Clear and close form
        document.getElementById('target-object-key').value = '';
        document.getElementById('new-object-key').value = '';
        document.getElementById('new-object-json').value = '';
        toggleInsertAfterForm();
        
        // Clear the search filter inputs to allow for a fresh search
        document.getElementById('fileExtFilter').value = '';
        document.getElementById('jsonKeyFilter').value = '';
        
    } catch (error) {
        showMessage('❌ Error during insert after: ' + error.message, 'error');
    }
}

// Show toast messages using Toastify
function showMessage(message, type = 'info') {
    // Convert message to string if it's not already
    const messageStr = String(message);
    
    // Map our types to Toastify styles
    const toastConfig = {
        text: messageStr,
        duration: 4000,
        close: true,
        gravity: "top",
        position: "right",
        stopOnFocus: true,
    };
    
    // Set colors based on message type
    switch (type) {
        case 'success':
            toastConfig.style = {
                background: "linear-gradient(to right, #00b09b, #96c93d)",
            };
            break;
        case 'error':
            toastConfig.style = {
                background: "linear-gradient(to right, #ff5f6d, #ffc371)",
            };
            toastConfig.duration = 6000; // Keep error messages longer
            break;
        case 'warning':
            toastConfig.style = {
                background: "linear-gradient(to right, #f093fb, #f5576c)",
            };
            break;
        case 'info':
        default:
            toastConfig.style = {
                background: "linear-gradient(to right, #4facfe, #00f2fe)",
            };
            break;
    }
    
    // Show the toast
    Toastify(toastConfig).showToast();
    
    // Also log to console for debugging
    console.log(`[${type.toUpperCase()}] ${messageStr}`);
}

// Initialize when page loads
document.addEventListener('DOMContentLoaded', initializeApp); 