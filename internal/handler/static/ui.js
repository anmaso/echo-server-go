// Load configurations
const loadConfigs = function () {
    fetch('/config')
        .then(response => response.json())
        .then(data => {
            const configList = document.getElementById('configList');
            configList.innerHTML = '';
            data.forEach(config => {
                const div = document.createElement('div');
                div.className = 'config-item';
                div.innerHTML = `
                        <h3>${config.pattern}</h3>
                        <p>Methods: ${[].concat(config.methods).join(', ')}</p>
                        <pre>${JSON.stringify(config, null, 2)}</pre>
                        <button onclick="deleteConfig('${config.name}')">Delete</button>
                    `;
                configList.appendChild(div);
            });
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

    // Initial load
    loadConfigs();
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
        .then(() => loadConfigs());
}