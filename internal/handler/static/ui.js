const sampleRequests = {
    configGet: {
        method: 'GET',
        path: '/config/paths',
        body: ''
    },
    configPostSimple: {
        method: 'POST',
        path: '/config/paths',
        body: JSON.stringify({
            name: "simple-endpoint",
            pattern: "^/api/simple$",
            methods: ["GET"],
            response: {
                statusCode: 200,
                body: '{"message": "simple response"}'
            }
        }, null, 2)
    },
    configPostDelay: {
        method: 'POST',
        path: '/config/paths',
        body: JSON.stringify({
            name: "delayed-endpoint",
            pattern: "^/api/delayed$",
            methods: ["GET"],
            response: {
                statusCode: 200,
                body: '{"message": "delayed response"}',
                delay: "2s"
            }
        }, null, 2)
    },
    configPostTemplate: {
        method: 'POST',
        path: '/config/paths',
        body: JSON.stringify({
            pattern: "^/api/template/.*$",
            methods: ["GET", "POST"],
            response: {
                statusCode: 200,
                body: 'template:{"path":"{{.Path}}","method":"{{.Method}}","headers":{{.Headers}}}',
                includeRequest: true
            }
        }, null, 2)
    },
    configPostError: {
        method: 'POST',
        path: '/config/paths',
        body: JSON.stringify({
            pattern: "^/api/error$",
            methods: ["GET"],
            response: {
                statusCode: 200,
                body: '{"status":"ok"}'
            },
            errorResponse: {
                statusCode: 500,
                body: '{"error":"simulated error"}'
            },
            errorFrequency: 0.5
        }, null, 2)
    },
    regularGet: {
        method: 'GET',
        path: '/api/simple',
        body: ''
    },
    regularPost: {
        method: 'POST',
        path: '/api/template/test',
        body: JSON.stringify({
            message: "Hello, world!",
            timestamp: "2024-01-01T00:00:00Z"
        }, null, 2)
    }
};

function setMethod(value) {
    const radio = document.querySelector(`input[name="method"][value="${value}"]`);
    if (radio) radio.checked = true;
}

function getMethod() {
    const checked = document.querySelector('input[name="method"]:checked');
    return checked ? checked.value : 'GET';
}

function syntaxHighlightJSON(json) {
    const escaped = json
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');

    return escaped.replace(
        /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
        (match) => {
            let cls = 'json-number';
            if (/^"/.test(match)) {
                cls = /:$/.test(match) ? 'json-key' : 'json-string';
            } else if (/true|false/.test(match)) {
                cls = 'json-boolean';
            } else if (/null/.test(match)) {
                cls = 'json-null';
            }
            return `<span class="${cls}">${match}</span>`;
        }
    );
}

function displayResponse(data) {
    const el = document.getElementById('response');
    try {
        const formatted = typeof data === 'string' ? data : JSON.stringify(data, null, 2);
        el.innerHTML = syntaxHighlightJSON(formatted);
    } catch (e) {
        el.textContent = String(data);
    }
}

let _selectedConfig = null;

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

function methodBadgesHTML(methods) {
    const colors = { GET: 'badge-get', POST: 'badge-post', PUT: 'badge-put', DELETE: 'badge-delete' };
    return (methods || []).map(m =>
        `<span class="method-badge ${colors[m] || ''}">${m}</span>`
    ).join('');
}

function renderConfig(config) {
    const row = document.createElement('div');
    row.className = 'config-row';
    row.dataset.name = config.name;

    row.innerHTML = `
        <span class="row-name">${config.name}</span>
        <span class="row-pattern">${config.pattern}</span>
        <span class="row-methods">${methodBadgesHTML(config.methods)}</span>
        <span class="row-actions">
            <button class="action-btn test-btn" title="Test">Test</button>
            <button class="action-btn edit-btn" title="Edit">Edit</button>
            <button class="action-btn delete-btn btn-danger" title="Delete">✕</button>
        </span>
    `;

    row.querySelector('.test-btn').addEventListener('click', (e) => {
        e.stopPropagation();
        testConfig(config);
    });
    row.querySelector('.edit-btn').addEventListener('click', (e) => {
        e.stopPropagation();
        editConfig(config);
    });
    row.querySelector('.delete-btn').addEventListener('click', (e) => {
        e.stopPropagation();
        if (confirm(`Delete "${config.name}"?`)) deleteConfig(config.name);
    });

    row.addEventListener('click', () => selectConfig(config, row));
    return row;
}

function selectConfig(config, row) {
    _selectedConfig = config;
    document.querySelectorAll('.config-row').forEach(r => r.classList.remove('selected'));
    row.classList.add('selected');

    document.getElementById('detailName').textContent = config.name;
    document.getElementById('detailBody').innerHTML = syntaxHighlightJSON(JSON.stringify(config, null, 2));
    document.getElementById('configDetail').classList.remove('hidden');
}

function closeDetail() {
    _selectedConfig = null;
    document.querySelectorAll('.config-row').forEach(r => r.classList.remove('selected'));
    document.getElementById('configDetail').classList.add('hidden');
}

function testConfig(config) {
    switchToTab('tester');
    setMethod(config.methods[0] || 'GET');
    document.getElementById('path').value = config.pattern.replace('^', '').replace('$', '');
    document.getElementById('requestBody').value = '';
    clearSampleSelection();
}

function editConfig(config) {
    switchToTab('tester');
    setMethod('POST');
    document.getElementById('path').value = '/config';
    document.getElementById('requestBody').value = JSON.stringify(config, null, 2);
    clearSampleSelection();
}

function switchToTab(tabId) {
    document.querySelectorAll('.tab-content').forEach(tab => tab.classList.remove('active'));
    document.getElementById(tabId).classList.add('active');
}

function clearSampleSelection() {
    document.querySelectorAll('#sampleRequestList li').forEach(li => li.classList.remove('active'));
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

    // keep detail pane in sync if the selected config is still visible
    if (_selectedConfig) {
        const row = configList.querySelector(`[data-name="${_selectedConfig.name}"]`);
        if (row) row.classList.add('selected');
        else closeDetail();
    }
}

// Load counters
function loadCounters() {
    fetch('/counter')
        .then(response => response.json())
        .then(data => {
            const counterList = document.getElementById('counterList');
            const rows = Object.entries(data.pathCounts)
                .map(([path, count]) => `
                    <tr>
                        <td class="counter-path">${path}</td>
                        <td class="counter-count">${count}</td>
                        <td class="counter-action">
                            <button class="action-btn" onclick="resetPathCounter('${path}')">Reset</button>
                        </td>
                    </tr>`)
                .join('');
            counterList.innerHTML = `
                <table class="counter-table">
                    <thead>
                        <tr>
                            <th>Path</th>
                            <th>Count</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr class="counter-global">
                            <td class="counter-path">— global —</td>
                            <td class="counter-count">${data.globalCount}</td>
                            <td></td>
                        </tr>
                        ${rows}
                    </tbody>
                </table>
            `;
        });
}

document.addEventListener('DOMContentLoaded', () => {
    // Tab switching
    document.querySelectorAll('nav a').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            switchToTab(e.target.dataset.tab);
        });
    });

    // Sample request list
    document.querySelectorAll('#sampleRequestList li').forEach(li => {
        li.addEventListener('click', () => {
            const sample = sampleRequests[li.dataset.key];
            if (!sample) return;

            clearSampleSelection();
            li.classList.add('active');

            setMethod(sample.method);
            document.getElementById('path').value = sample.path;
            document.getElementById('requestBody').value = sample.body;
        });
    });

    // Send request
    document.getElementById('sendRequest').addEventListener('click', () => {
        const method = getMethod();
        const path = document.getElementById('path').value;
        const body = document.getElementById('requestBody').value;

        document.getElementById('response').textContent = 'Sending request...';

        fetch(path, {
            method,
            headers: { 'Content-Type': 'application/json' },
            body: method !== 'GET' ? body : undefined
        })
            .then(response => response.json())
            .then(data => displayResponse(data))
            .catch(error => {
                document.getElementById('response').textContent = `Error: ${error.message}`;
            });
    });

    // Counter management
    document.getElementById('refreshCounters').addEventListener('click', loadCounters);
    document.getElementById('resetCounters').addEventListener('click', () => {
        fetch('/counter', { method: 'DELETE' }).then(() => loadCounters());
    });

    // Config management
    document.getElementById('configFilter').addEventListener('input', fetchConfigs);

    document.getElementById('closeDetail').addEventListener('click', closeDetail);
    document.getElementById('detailTest').addEventListener('click', () => { if (_selectedConfig) testConfig(_selectedConfig); });
    document.getElementById('detailEdit').addEventListener('click', () => { if (_selectedConfig) editConfig(_selectedConfig); });
    document.getElementById('detailDelete').addEventListener('click', () => {
        if (_selectedConfig && confirm(`Delete "${_selectedConfig.name}"?`)) {
            deleteConfig(_selectedConfig.name);
        }
    });

    fetchConfigs();
    loadCounters();
    initLog();
});

function resetPathCounter(path) {
    fetch(`/counter/${encodeURIComponent(path)}`, { method: 'DELETE' }).then(() => loadCounters());
}

function deleteConfig(pattern) {
    fetch(`/config/${encodeURIComponent(pattern)}`, { method: 'DELETE' }).then(() => fetchConfigs());
}

// ── Request Log ──────────────────────────────────────────────────────────────

const _log = {
    entries: [],  // all received entries, oldest first (display is reversed)
    filter: '',
    selected: null,
    paused: false,
};

function initLog() {
    fetch('/ui/requests/config')
        .then(r => r.json())
        .then(d => { document.getElementById('bufferSizeInput').value = d.bufferSize; })
        .catch(() => {});

    const es = new EventSource('/ui/requests/stream');
    es.onmessage = (e) => {
        const entry = JSON.parse(e.data);
        _log.entries.push(entry);
        if (!_log.paused && matchesFilter(entry, _log.filter)) {
            prependLogRow(entry);
        }
    };

    // Filter (regex)
    document.getElementById('logFilter').addEventListener('input', (e) => {
        _log.filter = e.target.value;
        const input = e.target;
        try {
            if (_log.filter) new RegExp(_log.filter, 'i');
            input.classList.remove('filter-invalid');
        } catch {
            input.classList.add('filter-invalid');
        }
        rebuildLogTable();
    });

    // Pause / Resume
    document.getElementById('pauseLog').addEventListener('click', (e) => {
        _log.paused = !_log.paused;
        e.target.textContent = _log.paused ? 'Resume' : 'Pause';
        e.target.classList.toggle('btn-paused', _log.paused);
        if (!_log.paused) rebuildLogTable();
    });

    // Clear
    document.getElementById('clearLog').addEventListener('click', () => {
        fetch('/ui/requests', { method: 'DELETE' }).then(() => {
            _log.entries = [];
            _log.selected = null;
            document.getElementById('logTableBody').innerHTML = '';
            document.getElementById('logDetail').classList.add('hidden');
        });
    });

    // Buffer size
    document.getElementById('setBufferSize').addEventListener('click', () => {
        const size = parseInt(document.getElementById('bufferSizeInput').value, 10);
        if (!size || size < 1) return;
        fetch('/ui/requests/config', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ bufferSize: size }),
        });
    });

    // Close detail
    document.getElementById('closeLogDetail').addEventListener('click', () => {
        _log.selected = null;
        document.querySelectorAll('#logTableBody tr').forEach(r => r.classList.remove('selected'));
        document.getElementById('logDetail').classList.add('hidden');
    });
}

function matchesFilter(entry, filter) {
    if (!filter) return true;
    let re;
    try {
        re = new RegExp(filter, 'i');
    } catch {
        return true; // invalid regex: show everything
    }
    const haystack = [
        entry.method, entry.path, entry.query,
        String(entry.statusCode),
        entry.requestBody, entry.responseBody,
    ].filter(Boolean).join(' ');
    return re.test(haystack);
}

function rebuildLogTable() {
    const tbody = document.getElementById('logTableBody');
    tbody.innerHTML = '';
    // iterate oldest→newest; prepend each so newest ends up at top
    for (const entry of _log.entries) {
        if (matchesFilter(entry, _log.filter)) prependLogRow(entry);
    }
}

function prependLogRow(entry) {
    const tbody = document.getElementById('logTableBody');
    const tr = document.createElement('tr');
    tr.dataset.id = entry.id;
    if (_log.selected && _log.selected.id === entry.id) tr.classList.add('selected');

    const t = new Date(entry.timestamp);
    const time = `${String(t.getHours()).padStart(2,'0')}:${String(t.getMinutes()).padStart(2,'0')}:${String(t.getSeconds()).padStart(2,'0')}.${String(t.getMilliseconds()).padStart(3,'0')}`;
    const duration = entry.durationMs >= 1000
        ? `${(entry.durationMs / 1000).toFixed(2)}s`
        : `${entry.durationMs.toFixed(1)}ms`;

    const statusCls = entry.statusCode >= 500 ? 'status-5xx'
        : entry.statusCode >= 400 ? 'status-4xx'
        : entry.statusCode >= 300 ? 'status-3xx'
        : 'status-2xx';

    const methodColors = { GET: 'badge-get', POST: 'badge-post', PUT: 'badge-put', DELETE: 'badge-delete' };

    tr.innerHTML = `
        <td class="col-time">${time}</td>
        <td class="col-method"><span class="method-badge ${methodColors[entry.method] || ''}">${entry.method}</span></td>
        <td class="col-path">${escapeHtml(entry.path)}${entry.query ? '<span class="query-str">?' + escapeHtml(entry.query) + '</span>' : ''}</td>
        <td class="col-status"><span class="status-badge ${statusCls}">${entry.statusCode}</span></td>
        <td class="col-duration">${duration}</td>
    `;
    tr.addEventListener('click', () => selectLogEntry(entry, tr));
    tbody.prepend(tr);
}

function selectLogEntry(entry, tr) {
    _log.selected = entry;
    document.querySelectorAll('#logTableBody tr').forEach(r => r.classList.remove('selected'));
    tr.classList.add('selected');

    const t = new Date(entry.timestamp);
    document.getElementById('logDetailTitle').textContent =
        `${entry.method} ${entry.path} — ${t.toISOString()}`;

    document.getElementById('logDetailBody').innerHTML = `
        <div class="detail-section">
            <div class="detail-section-title">Request</div>
            <div class="detail-row"><span class="detail-key">Method</span><span>${entry.method}</span></div>
            <div class="detail-row"><span class="detail-key">Path</span><span>${escapeHtml(entry.path)}</span></div>
            ${entry.query ? `<div class="detail-row"><span class="detail-key">Query</span><span>${escapeHtml(entry.query)}</span></div>` : ''}
            <div class="detail-row"><span class="detail-key">Headers</span><pre class="detail-pre">${escapeHtml(formatHeaders(entry.requestHeaders))}</pre></div>
            <div class="detail-row"><span class="detail-key">Body</span><pre class="detail-pre">${formatBody(entry.requestBody)}</pre></div>
        </div>
        <div class="detail-section">
            <div class="detail-section-title">Response</div>
            <div class="detail-row"><span class="detail-key">Status</span><span>${entry.statusCode}</span></div>
            <div class="detail-row"><span class="detail-key">Duration</span><span>${entry.durationMs.toFixed(2)}ms</span></div>
            <div class="detail-row"><span class="detail-key">Headers</span><pre class="detail-pre">${escapeHtml(formatHeaders(entry.responseHeaders))}</pre></div>
            <div class="detail-row"><span class="detail-key">Body</span><pre class="detail-pre">${formatBody(entry.responseBody)}</pre></div>
        </div>
    `;
    document.getElementById('logDetail').classList.remove('hidden');
}

function formatHeaders(headers) {
    if (!headers) return '(none)';
    return Object.entries(headers).map(([k, v]) => `${k}: ${v.join(', ')}`).join('\n');
}

function formatBody(body) {
    if (!body) return '<span class="empty-body">(empty)</span>';
    try {
        return syntaxHighlightJSON(JSON.stringify(JSON.parse(body), null, 2));
    } catch {
        return escapeHtml(body);
    }
}

function escapeHtml(str) {
    if (!str) return '';
    return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
