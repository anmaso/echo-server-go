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