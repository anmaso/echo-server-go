// Load configurations
async function fetchConfigs() {
    try {
        const response = await fetch('/config');
        const configs = await response.json();
        updateConfigList(configs);
        return configs;
    } catch (error) {
        console.error('Error fetching configurations:', error);
    }
}

function renderConfig(config) {
    const item = document.createElement('div');
    item.className = 'config-item collapsed';
    
    const header = document.createElement('div');
    header.className = 'config-header';
    header.innerHTML = `
        <div class="config-title">
            <strong>${config.name}</strong>
            <span>${config.pattern}</span>
        </div>
        <div class="config-actions">
            <button class="action-btn test-btn" title="Test config">Test</button>
            <button class="action-btn edit-btn" title="Edit config">Edit</button>
            <button class="action-btn delete-btn" title="Delete config">Delete</button>
        </div>
    `;

    const content = document.createElement('div');
    content.className = 'config-content';
    content.innerHTML = `<pre>${JSON.stringify(config, null, 2)}</pre>`;

    item.appendChild(header);
    item.appendChild(content);

    // Handle expand/collapse on header text click
    header.querySelector('.config-title').addEventListener('click', (e) => {
        item.classList.toggle('collapsed');
    });

    // Handle button clicks
    header.querySelector('.test-btn').addEventListener('click', (e) => {
        e.stopPropagation();
        // Switch to tester tab
        switchToTab('tester');
        // Populate form with test values
        document.getElementById('method').value = config.methods[0] || 'GET';
        document.getElementById('path').value = config.pattern.replace('^', '').replace('$', '');
        document.getElementById('requestBody').value = '';
    });

    header.querySelector('.edit-btn').addEventListener('click', (e) => {
        e.stopPropagation();
        // Switch to tester tab
        switchToTab('tester');
        // Populate form with config values
        document.getElementById('method').value = 'POST';
        document.getElementById('path').value = '/config';
        document.getElementById('requestBody').value = JSON.stringify(config, null, 2);
    });

    header.querySelector('.delete-btn').addEventListener('click', (e) => {
        e.stopPropagation();
        if (confirm(`Are you sure you want to delete the configuration "${config.name}"?`)) {
            deleteConfig(config.name);
        }
    });

    return item;
}

// Add this helper function to switch tabs
function switchToTab(tabId) {
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
    });
    document.getElementById(tabId).classList.add('active');
}

function updateConfigList(configs) {
    const configList = document.getElementById('configList');
    const filterText = document.getElementById('configFilter').value.toLowerCase();
    
    configList.innerHTML = '';
    configs.forEach(config => {
        if (config.name.toLowerCase().includes(filterText) || 
            config.pattern.toLowerCase().includes(filterText)) {
            configList.appendChild(renderConfig(config));
        }
    });
}

// Load counters
function loadCounters() {
    fetch('/counter')
        .then(response => response.json())
        .then(data => {
            const counterList = document.getElementById('counterList');
            counterList.innerHTML = `
                    <div class="counter-item">
                        <h3>Global Counter: ${data.globalCount}</h3>
                    </div>
                `;
            Object.entries(data.pathCounts).forEach(([path, count]) => {
                const div = document.createElement('div');
                div.className = 'counter-item';
                div.innerHTML = `
                        <h3>${path}</h3>
                        <p>Count: ${count}</p>
                        <button onclick="resetPathCounter('${path}')">Reset</button>
                    `;
                counterList.appendChild(div);
            });
        });
}

document.addEventListener('DOMContentLoaded', () => {
    // Tab switching
    document.querySelectorAll('nav a').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const tabId = e.target.dataset.tab;
            document.querySelectorAll('.tab-content').forEach(tab => {
                tab.classList.remove('active');
            });
            document.getElementById(tabId).classList.add('active');
        });
    });

    // Send test request
    document.getElementById('sendRequest').addEventListener('click', () => {
        const method = document.getElementById('method').value;
        const path = document.getElementById('path').value;
        const body = document.getElementById('requestBody').value;
        document.getElementById('response').textContent = 'Sending request...';

        fetch(path, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: method !== 'GET' ? body : undefined
        })
            .then(response => response.json())
            .then(data => {
                document.getElementById('response').textContent =
                    JSON.stringify(data, null, 2);
            })
            .catch(error => {
                document.getElementById('response').textContent =
                    `Error: ${error.message}`;
            });
    });

    // Counter management
    document.getElementById('refreshCounters').addEventListener('click', loadCounters);
    document.getElementById('resetCounters').addEventListener('click', () => {
        fetch('/counter', { method: 'DELETE' })
            .then(() => loadCounters());
    });

    // Config management
    document.getElementById('configFilter').addEventListener('input', () => {
        fetchConfigs().then(updateConfigList);
    });

    document.getElementById('toggleAll').addEventListener('click', function() {
        const items = document.querySelectorAll('.config-item');
        const isAnyCollapsed = Array.from(items).some(item => item.classList.contains('collapsed'));
        
        items.forEach(item => {
            if (isAnyCollapsed) {
                item.classList.remove('collapsed');
            } else {
                item.classList.add('collapsed');
            }
        });
        
        this.textContent = isAnyCollapsed ? 'Collapse All' : 'Expand All';
    });

    // Initial load
    fetchConfigs();
    loadCounters();
});

// Counter reset function
function resetPathCounter(path) {
    fetch(`/counter/${encodeURIComponent(path)}`, { method: 'DELETE' })
        .then(() => loadCounters());
}

// Config deletion function
function deleteConfig(pattern) {
    fetch(`/config/${encodeURIComponent(pattern)}`, { method: 'DELETE' })
        .then(() => fetchConfigs());
}

// --- History Tab Logic ---

// History Tab Elements (assuming they are declared after DOMContentLoaded or globally accessible)
let startHistoryBtn, stopHistoryBtn, historyMaxSizeInput, updateHistoryConfigBtn,
    refreshHistoryBtn, clearHistoryBtn, historyListDiv, historyStatusSpan,
    currentHistoryMaxSizeSpan;

function setupHistoryTabElements() {
    startHistoryBtn = document.getElementById('startHistory');
    stopHistoryBtn = document.getElementById('stopHistory');
    historyMaxSizeInput = document.getElementById('historyMaxSize');
    updateHistoryConfigBtn = document.getElementById('updateHistoryConfig');
    refreshHistoryBtn = document.getElementById('refreshHistory');
    clearHistoryBtn = document.getElementById('clearHistory');
    historyListDiv = document.getElementById('historyList');
    historyStatusSpan = document.getElementById('historyStatus');
    currentHistoryMaxSizeSpan = document.getElementById('currentHistoryMaxSize');

    // History Event Listeners
    if (startHistoryBtn) startHistoryBtn.addEventListener('click', startHistoryRecording);
    if (stopHistoryBtn) stopHistoryBtn.addEventListener('click', stopHistoryRecording);
    if (updateHistoryConfigBtn) updateHistoryConfigBtn.addEventListener('click', updateHistoryConfiguration);
    if (refreshHistoryBtn) refreshHistoryBtn.addEventListener('click', fetchHistory);
    if (clearHistoryBtn) clearHistoryBtn.addEventListener('click', clearRequestHistory);
}


async function updateHistoryStatus(data) {
    if (data && historyStatusSpan && currentHistoryMaxSizeSpan && historyMaxSizeInput) {
        historyStatusSpan.textContent = data.recordingActive ? 'Recording Active' : 'Not Recording';
        currentHistoryMaxSizeSpan.textContent = data.maxSize;
        historyMaxSizeInput.value = data.maxSize;
    } else {
        // console.warn("History status elements not found or no data provided for update.");
        // Attempt to fetch fresh status if elements exist but no data given
        if (historyStatusSpan && currentHistoryMaxSizeSpan && historyMaxSizeInput) {
            try {
                const initialMaxSize = parseInt(historyMaxSizeInput.value, 10) || 100; // Fallback
                const response = await fetch('/history/config', {
                    method: 'PUT', // This endpoint updates and returns current state
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ maxSize: initialMaxSize })
                });
                if (response.ok) {
                    const currentData = await response.json();
                    historyStatusSpan.textContent = currentData.recordingActive ? 'Recording Active' : 'Not Recording';
                    currentHistoryMaxSizeSpan.textContent = currentData.maxSize;
                    historyMaxSizeInput.value = currentData.maxSize;
                } else {
                     historyStatusSpan.textContent = 'Error fetching status';
                }
            } catch (e) {
                console.error("Error fetching history status:", e);
                if(historyStatusSpan) historyStatusSpan.textContent = 'Error fetching status';
            }
        }
    }
}

async function startHistoryRecording() {
    try {
        const maxSize = parseInt(historyMaxSizeInput.value, 10);
        // Include maxSize in the body, even if it's the current value,
        // as the /start endpoint can optionally reconfigure maxSize.
        const response = await fetch('/history/start', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ maxSize: isNaN(maxSize) ? -1 : maxSize }) // Server handles -1 as "don't change" or default if appropriate
        });
        if (!response.ok) throw new Error(`Error starting history: ${response.statusText} (${response.status})`);
        const data = await response.json();
        updateHistoryStatus(data);
        fetchHistory();
    } catch (error) {
        console.error('Error starting history recording:', error);
        alert(`Error starting history: ${error.message}`);
    }
}

async function stopHistoryRecording() {
    try {
        const response = await fetch('/history/stop', { method: 'POST' });
        if (!response.ok) throw new Error(`Error stopping history: ${response.statusText} (${response.status})`);
        const data = await response.json();
        updateHistoryStatus(data);
    } catch (error) {
        console.error('Error stopping history recording:', error);
        alert(`Error stopping history: ${error.message}`);
    }
}

async function updateHistoryConfiguration() {
    try {
        const maxSize = parseInt(historyMaxSizeInput.value, 10);
        if (isNaN(maxSize) || maxSize < 0) {
            alert("Please enter a valid non-negative integer for max size.");
            return;
        }
        const response = await fetch('/history/config', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ maxSize: maxSize })
        });
        if (!response.ok) throw new Error(`Error updating config: ${response.statusText} (${response.status})`);
        const data = await response.json();
        updateHistoryStatus(data);
    } catch (error) {
        console.error('Error updating history configuration:', error);
        alert(`Error updating config: ${error.message}`);
    }
}

async function fetchHistory() {
    if (!historyListDiv) return;
    try {
        const response = await fetch('/history');
        if (!response.ok) throw new Error(`Error fetching history: ${response.statusText} (${response.status})`);
        const entries = await response.json();
        displayHistory(entries);
    } catch (error) {
        console.error('Error fetching history:', error);
        historyListDiv.innerHTML = '<p>Error loading history.</p>';
    }
}

function renderHistoryEntry(entry) {
    const item = document.createElement('div');
    item.className = 'history-item';

    const header = document.createElement('div');
    header.className = 'history-item-header';
    // Sanitize path to prevent XSS if path could contain HTML
    const pathText = entry.request.path.replace(/</g, "&lt;").replace(/>/g, "&gt;");
    header.innerHTML = `
        <strong>${new Date(entry.timestamp).toLocaleString()}</strong> - 
        ${entry.request.method} ${pathText} -> ${entry.response.statusCode}
    `;
    header.onclick = () => {
      item.classList.toggle('expanded');
      content.style.display = item.classList.contains('expanded') ? 'block' : 'none';
    }

    const content = document.createElement('div');
    content.className = 'history-item-content';
    content.style.display = 'none'; // Initially hidden
    
    // Basic syntax highlighting for JSON
    const prettyRequest = JSON.stringify(entry.request, null, 2)
        .replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, 
            function (match) {
                let cls = 'json-number';
                if (/^"/.test(match)) {
                    if (/:$/.test(match)) {
                        cls = 'json-key';
                    } else {
                        cls = 'json-string';
                    }
                } else if (/true|false/.test(match)) {
                    cls = 'json-boolean';
                } else if (/null/.test(match)) {
                    cls = 'json-null';
                }
                return '<span class="' + cls + '">' + match + '</span>';
            });

    const prettyResponse = JSON.stringify(entry.response, null, 2)
        .replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
            function (match) {
                let cls = 'json-number';
                if (/^"/.test(match)) {
                    if (/:$/.test(match)) {
                        cls = 'json-key';
                    } else {
                        cls = 'json-string';
                    }
                } else if (/true|false/.test(match)) {
                    cls = 'json-boolean';
                } else if (/null/.test(match)) {
                    cls = 'json-null';
                }
                return '<span class="' + cls + '">' + match + '</span>';
            });

    content.innerHTML = `
        <div class="history-detail-section"><strong>Request:</strong><pre>${prettyRequest}</pre></div>
        <div class="history-detail-section"><strong>Response:</strong><pre>${prettyResponse}</pre></div>
        <p>Response Size: ${entry.responseSize} bytes</p>
    `;
    item.appendChild(header);
    item.appendChild(content);
    return item;
}

function displayHistory(entries) {
    if (!historyListDiv) return;
    historyListDiv.innerHTML = ''; 
    if (!entries || entries.length === 0) {
        historyListDiv.innerHTML = '<p>No history recorded.</p>';
        return;
    }
    entries.reverse().forEach(entry => { 
        historyListDiv.appendChild(renderHistoryEntry(entry));
    });
}

async function clearRequestHistory() {
    if (!confirm("Are you sure you want to clear all request history?")) return;
    try {
        const response = await fetch('/history', { method: 'DELETE' });
        if (!response.ok) throw new Error(`Error clearing history: ${response.statusText} (${response.status})`);
        // HTTP 204 No Content, no JSON body to parse
        fetchHistory(); 
        updateHistoryStatus(); // Fetch fresh status as clear might affect it indirectly (e.g. if maxSize was 0)
    } catch (error) {
        console.error('Error clearing history:', error);
        alert(`Error clearing history: ${error.message}`);
    }
}

async function initializeHistoryTab() {
    setupHistoryTabElements(); // Ensure elements are available
    if (!historyMaxSizeInput || !historyStatusSpan) { // Check if history tab elements are on the page
        // console.log("History tab elements not found, skipping initialization for this tab.");
        return;
    }
    try {
        // Get initial status and MaxSize
        // A GET /history/status endpoint would be cleaner.
        // For now, using the "PUT /history/config with current value" approach.
        // This ensures the input field and status are synchronized with the backend.
        const initialMaxSize = parseInt(historyMaxSizeInput.value, 10);
        if (isNaN(initialMaxSize)) { // if input is initially empty or invalid
             historyMaxSizeInput.value = 100; // Set a default for the call
        }
        
        const response = await fetch('/history/config', { 
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            // Send a valid maxSize, defaulting if necessary
            body: JSON.stringify({ maxSize: isNaN(initialMaxSize) ? 100 : initialMaxSize }) 
        });

        if (response.ok) {
            const data = await response.json();
            updateHistoryStatus(data);
        } else {
            // Try to fetch history even if status update fails initially
            console.warn("Failed to get initial history config, UI might not reflect server state.", response.statusText);
            updateHistoryStatus(null); // Will try to fetch status itself
        }
        fetchHistory();
    } catch (e) {
        console.error("Error initializing history tab:", e);
        if(historyStatusSpan) historyStatusSpan.textContent = 'Error initializing.';
    }
}

// Modify the DOMContentLoaded listener to include history tab initialization
document.addEventListener('DOMContentLoaded', () => {
    // ... existing DOMContentLoaded code ...
    
    // Tab switching (ensure this is only set up once)
    // It's safer to check if it's already done or ensure this whole script runs once.
    // For this structure, we assume this is the main DOMContentLoaded setup.
    const navLinks = document.querySelectorAll('nav a');
    navLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const tabId = e.target.dataset.tab;
            document.querySelectorAll('.tab-content').forEach(tab => {
                tab.classList.remove('active');
            });
            const activeTab = document.getElementById(tabId);
            if (activeTab) {
                activeTab.classList.add('active');
            }

            // If switching to history tab, initialize it (or refresh)
            if (tabId === 'history') {
                initializeHistoryTab(); // Or just fetchHistory() if already initialized
            }
        });
    });
    
    // Initial setup for elements that might not be part of a specific tab's init
    // For example, the request tester form elements if they are always present.
    // const sendRequestBtn = document.getElementById('sendRequest');
    // if (sendRequestBtn) { ... }


    // Call setup for history tab elements specifically.
    // This ensures that even if the history tab is not active initially,
    // its elements are found if present on the page.
    setupHistoryTabElements();


    // Initial load for other tabs (if necessary)
    if (document.getElementById('configs')) fetchConfigs();
    if (document.getElementById('counters')) loadCounters();
    
    // Initialize history tab if it's the default active tab or just to get initial values
    // Check if the history tab exists before trying to initialize it
    if (document.getElementById('history')) {
        initializeHistoryTab();
    }
});

// Ensure other specific initializations like for configFilter, toggleAll, etc.,
// are also within DOMContentLoaded or appropriately scoped.
// The original script had them inside DOMContentLoaded.

// The following are existing functions that should remain:
// fetchConfigs, renderConfig, switchToTab, updateConfigList, loadCounters, 
// resetPathCounter, deleteConfig.
// Make sure they are not duplicated. The diff should append the new history logic.
// For this exercise, I'm providing the history logic as a self-contained block
// to be appended and then integrated into the DOMContentLoaded.

// Note: The provided diff will append this. Manual merge into DOMContentLoaded might be needed
// if the structure is complex. The goal is to add the history functionality.
// The `setupHistoryTabElements` and `initializeHistoryTab` calls should be correctly placed.