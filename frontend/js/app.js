let currentFiles = [];

async function searchFiles() {
    const folderPath = document.getElementById('folderPath').value.trim();
    const fileExtensionInput = document.getElementById('fileExtension').value.trim();
    const jsonKeyFilter = document.getElementById('jsonKeyFilter').value.trim();
    
    if (!folderPath) {
        showMessage('Base path not configured. Please check config.env file.', 'error');
        return;
    }

    document.getElementById('filesContainer').innerHTML = '<div class="loading"><p>🔍 Searching for files...</p></div>';

    try {
        // Pass both extension and JSON key filters to the backend
        const rootNode = await window.browseFolder(folderPath, fileExtensionInput, jsonKeyFilter);
        
        currentFiles = rootNode.allFiles;
        displayFileTree(rootNode);
        
        const fileType = fileExtensionInput || 'all';
        const keyFilter = jsonKeyFilter ? ` with key "${jsonKeyFilter}"` : '';
        showMessage(`Found ${rootNode.count} files matching: ${fileType}${keyFilter}`, 'success');
    } catch (error) {
        showMessage('Error searching files: ' + error.message, 'error');
        document.getElementById('filesContainer').innerHTML = '<div class="loading"><p>❌ Error searching files</p></div>';
    }
}

function showMessage(message, type = 'info') {
    const messageDiv = document.getElementById('message');
    messageDiv.innerHTML = '<div class="' + (type === 'error' ? 'error' : 'success') + '">' + message + '</div>';
    setTimeout(() => {
        messageDiv.innerHTML = '';
    }, 5000);
}

function displayFileTree(rootNode) {
    const container = document.getElementById('filesContainer');
    
    if (rootNode.count === 0) {
        container.innerHTML = '<div class="loading"><p>📄 No files found matching the specified criteria</p></div>';
        return;
    }

    const treeHTML = '<div class="file-tree">' +
        '<div class="tree-header">' +
            '<h3>📁 ' + rootNode.name + ' (' + rootNode.count + ' files)</h3>' +
            '<button class="btn btn-secondary mass-update-btn" onclick="toggleMassUpdateForm()">📝 Mass Update All Files</button>' +
        '</div>' +
        '<div id="mass-update-form" class="mass-update-form" style="display: none;">' +
            '<h4>🚀 Add JSON Item to All ' + rootNode.count + ' Files</h4>' +
            '<div class="form-row">' +
                '<input type="text" id="mass-object-path" placeholder="Object path (e.g., user.address or leave empty for root)" />' +
                '<input type="text" id="mass-key" placeholder="Key name" />' +
            '</div>' +
            '<textarea id="mass-value" placeholder="Value (JSON format, e.g., &quot;string&quot;, 123, {&quot;nested&quot;: true})"></textarea>' +
            '<div class="mass-update-buttons">' +
                '<button class="btn btn-primary" onclick="performMassUpdate()">🔄 Update All Files</button>' +
                '<button class="btn" onclick="toggleMassUpdateForm()">Cancel</button>' +
            '</div>' +
        '</div>' +
        renderTreeNode(rootNode, 0) +
        '</div>';

    container.innerHTML = treeHTML;
}

function renderTreeNode(node, depth) {
    let html = '';
    const indent = '  '.repeat(depth);
    
    if (node.isDir) {
        // Render directory
        const hasContent = (node.children && node.children.length > 0) || (node.files && node.files.length > 0);
        if (hasContent) {
            html += '<div class="tree-folder" style="margin-left: ' + (depth * 20) + 'px;">' +
                '<span class="folder-toggle" onclick="toggleFolder(this)">📁</span> ' +
                '<span class="folder-name">' + node.name + '</span>' +
                '<div class="folder-content" style="display: block;">';
            
            // Render files in this directory
            if (node.files && node.files.length > 0) {
                node.files.forEach((file, index) => {
                    const globalIndex = currentFiles.findIndex(f => f.path === file.path);
                    html += '<div class="tree-file" style="margin-left: ' + ((depth + 1) * 20) + 'px;">' +
                        '<div class="file-item">' +
                            '<span class="file-icon">📄</span> ' +
                            '<span class="file-name">' + file.name + '</span>' +
                            '<button class="btn btn-small file-action-btn" onclick="loadFileContent(' + globalIndex + ')">View</button>' +
                        '</div>' +
                        '<div class="file-path">' + file.path + '</div>' +
                        '<div id="content-' + globalIndex + '" class="file-content" style="display: none;"></div>' +
                        '</div>';
                });
            }
            
            // Render subdirectories
            if (node.children && node.children.length > 0) {
                node.children.forEach(child => {
                    html += renderTreeNode(child, depth + 1);
                });
            }
            
            html += '</div></div>';
        }
    }
    
    return html;
}

function toggleFolder(element) {
    const folderContent = element.parentElement.querySelector('.folder-content');
    const isOpen = folderContent.style.display !== 'none';
    
    if (isOpen) {
        folderContent.style.display = 'none';
        element.textContent = '📂';
    } else {
        folderContent.style.display = 'block';
        element.textContent = '📁';
    }
}

// Keep the old displayFiles function for backward compatibility if needed
function displayFiles(files) {
    const container = document.getElementById('filesContainer');
    
    if (files.length === 0) {
        container.innerHTML = '<div class="loading"><p>📄 No files found matching the specified criteria</p></div>';
        return;
    }

    const filesHTML = files.map((file, index) => {
        return '<div class="file-card">' +
            '<div class="file-header">' +
                '<div class="file-name">📄 ' + file.name + '</div>' +
            '</div>' +
            '<div class="file-path">' + file.path + '</div>' +
            '<div class="file-actions">' +
                '<button class="btn btn-small" onclick="loadFileContent(' + index + ')">View Content</button>' +
            '</div>' +
            '<div id="content-' + index + '" class="file-content" style="display: none;"></div>' +
        '</div>';
    }).join('');

    container.innerHTML = '<div class="files-grid">' + filesHTML + '</div>';
}

async function loadFileContent(fileIndex) {
    const file = currentFiles[fileIndex];
    const contentDiv = document.getElementById('content-' + fileIndex);
    
    if (contentDiv.style.display === 'none') {
        try {
            contentDiv.innerHTML = '<div class="loading-content">📄 Loading file content...</div>';
            contentDiv.style.display = 'block';
            
            const content = await window.getJSONFileContent(file.path);
            const beautifulJson = formatJsonContent(content);
            contentDiv.innerHTML = beautifulJson;
        } catch (error) {
            contentDiv.innerHTML = '<div class="error">❌ Error loading file: ' + error.message + '</div>';
            contentDiv.style.display = 'block';
        }
    } else {
        contentDiv.style.display = 'none';
    }
}

function formatJsonContent(jsonObject) {
    const jsonString = JSON.stringify(jsonObject, null, 2);
    
    // Create a beautiful formatted JSON display
    const lines = jsonString.split('\n');
    let html = '<div class="json-display">';
    html += '<div class="json-header">';
    html += '<span class="json-title">📄 JSON Content</span>';
    html += '<button class="copy-btn" onclick="copyJsonToClipboard(this)" title="Copy to clipboard">📋 Copy</button>';
    html += '</div>';
    html += '<div class="json-content">';
    html += '<pre class="json-code">';
    
    lines.forEach((line, index) => {
        const lineNumber = (index + 1).toString().padStart(3, ' ');
        const indentLevel = (line.match(/^\s*/)[0].length / 2);
        const trimmedLine = line.trim();
        
        let coloredLine = trimmedLine;
        
        // Color different JSON elements
        coloredLine = coloredLine.replace(/"([^"]+)":/g, '<span class="json-key">"$1"</span>:');
        coloredLine = coloredLine.replace(/:\s*"([^"]*)"([,}]?)/g, ': <span class="json-string">"$1"</span>$2');
        coloredLine = coloredLine.replace(/:\s*(\d+\.?\d*)([,}]?)/g, ': <span class="json-number">$1</span>$2');
        coloredLine = coloredLine.replace(/:\s*(true|false)([,}]?)/g, ': <span class="json-boolean">$1</span>$2');
        coloredLine = coloredLine.replace(/:\s*(null)([,}]?)/g, ': <span class="json-null">$1</span>$2');
        coloredLine = coloredLine.replace(/([{}\[\]])/g, '<span class="json-bracket">$1</span>');
        
        html += '<span class="json-line">';
        html += '<span class="line-number">' + lineNumber + '</span>';
        html += '<span class="line-content" style="padding-left: ' + (indentLevel * 20) + 'px;">' + coloredLine + '</span>';
        html += '</span>';
    });
    
    html += '</pre>';
    html += '</div>';
    html += '<div class="json-stats">';
    html += '<span>📊 ' + Object.keys(jsonObject).length + ' root properties</span>';
    html += '<span>📏 ' + lines.length + ' lines</span>';
    html += '<span>💾 ' + new Blob([jsonString]).size + ' bytes</span>';
    html += '</div>';
    html += '</div>';
    
    return html;
}

function copyJsonToClipboard(button) {
    const jsonContent = button.closest('.json-display').querySelector('.json-code').textContent;
    navigator.clipboard.writeText(jsonContent).then(() => {
        const originalText = button.textContent;
        button.textContent = '✅ Copied!';
        button.style.backgroundColor = '#10b981';
        setTimeout(() => {
            button.textContent = originalText;
            button.style.backgroundColor = '';
        }, 2000);
    }).catch(err => {
        console.error('Failed to copy: ', err);
        button.textContent = '❌ Failed';
        setTimeout(() => {
            button.textContent = '📋 Copy';
        }, 2000);
    });
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function toggleMassUpdateForm() {
    const form = document.getElementById('mass-update-form');
    if (form.style.display === 'none') {
        form.style.display = 'block';
        
        // Auto-populate object path with current JSON key filter
        const jsonKeyFilter = document.getElementById('jsonKeyFilter').value.trim();
        if (jsonKeyFilter) {
            document.getElementById('mass-object-path').value = jsonKeyFilter;
        }
    } else {
        form.style.display = 'none';
        // Clear form when closing
        document.getElementById('mass-object-path').value = '';
        document.getElementById('mass-key').value = '';
        document.getElementById('mass-value').value = '';
    }
}

async function performMassUpdate() {
    const objectPath = document.getElementById('mass-object-path').value.trim();
    const key = document.getElementById('mass-key').value.trim();
    const valueStr = document.getElementById('mass-value').value.trim();

    if (!key || !valueStr) {
        showMessage('❌ Please enter both key and value', 'error');
        return;
    }

    try {
        // Parse the value as JSON
        const value = JSON.parse(valueStr);
        
        // Get all file paths from currentFiles
        const filePaths = currentFiles.map(file => file.path);
        
        if (filePaths.length === 0) {
            showMessage('❌ No files to update', 'error');
            return;
        }

        // Show progress message
        showMessage(`🔄 Updating ${filePaths.length} files...`, 'info');
        
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
            showMessage(`✅ Successfully updated ${successCount} files with key "${key}"`, 'success');
        } else {
            showMessage(`⚠️ Updated ${successCount} files, ${errorCount} failed. Check console for details.`, 'error');
            console.error('Mass update errors:', errors);
        }
        
        // Clear and close form
        document.getElementById('mass-object-path').value = '';
        document.getElementById('mass-key').value = '';
        document.getElementById('mass-value').value = '';
        toggleMassUpdateForm();
        
    } catch (error) {
        showMessage('❌ Error during mass update: ' + error.message, 'error');
    }
}

// Set default folder path from environment variable or default
document.addEventListener('DOMContentLoaded', async function () {
    try {
        const basePath = await getBasePath();
        document.getElementById('folderPath').value = basePath;
        showMessage('Base path loaded from config.env: ' + basePath, 'success');
    } catch (error) {
        showMessage('Could not load base path from config.env. Please check the file.', 'error');
        document.getElementById('folderPath').value = 'Please configure config.env';
    }
}); 