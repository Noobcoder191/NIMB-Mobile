// NIMB Mobile Frontend - Fetch API Version

let settingsSaveInProgress = false;

// Toast notifications
function showToast(message, type = 'info') {
    const container = document.getElementById('toastContainer');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `<span>${message}</span>`;
    container.appendChild(toast);

    setTimeout(() => {
        toast.classList.add('hide');
        setTimeout(() => toast.remove(), 200);
    }, 3000);
}

// Confirmation modal
function showConfirm(message) {
    return new Promise((resolve) => {
        const overlay = document.createElement('div');
        overlay.className = 'setup-overlay';
        overlay.innerHTML = `
            <div class="setup-modal" style="max-width: 400px; padding: 32px;">
                <h2 style="font-family: var(--font-display); margin-bottom: 16px;">Confirm Action</h2>
                <p style="color: var(--text-secondary); margin-bottom: 24px;">${message}</p>
                <div style="display: flex; gap: 12px; justify-content: flex-end;">
                    <button class="btn btn-secondary" id="confirmCancel">Cancel</button>
                    <button class="btn btn-danger" id="confirmOk">Confirm</button>
                </div>
            </div>
        `;
        document.body.appendChild(overlay);

        overlay.querySelector('#confirmCancel').onclick = () => {
            overlay.remove();
            resolve(false);
        };
        overlay.querySelector('#confirmOk').onclick = () => {
            overlay.remove();
            resolve(true);
        };
    });
}

// Copy to clipboard
async function copyToClipboard(text) {
    try {
        await navigator.clipboard.writeText(text);
        showToast('Copied to clipboard', 'success');
    } catch (err) {
        showToast('Failed to copy', 'error');
    }
}

// Navigation
function switchPage(pageId) {
    document.querySelectorAll('section').forEach(p => p.classList.remove('active'));
    document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));

    document.getElementById(pageId).classList.add('active');
    document.querySelector(`[data-page="${pageId}"]`).classList.add('active');

    const title = document.getElementById('pageTitle');
    title.style.animation = 'none';
    title.offsetHeight;
    title.style.animation = 'fadeSlide 0.4s var(--ease-out)';
    title.textContent = pageId.charAt(0).toUpperCase() + pageId.slice(1);

    // Close mobile sidebar and overlay if open
    document.querySelector('.sidebar')?.classList.remove('open');
    const overlay = document.querySelector('.sidebar-overlay');
    if (overlay) overlay.classList.remove('show');
}

// Mobile hamburger menu
function toggleSidebar() {
    const sidebar = document.querySelector('.sidebar');
    const isOpen = sidebar?.classList.toggle('open');

    // Create or remove overlay
    let overlay = document.querySelector('.sidebar-overlay');
    if (isOpen) {
        if (!overlay) {
            overlay = document.createElement('div');
            overlay.className = 'sidebar-overlay';
            overlay.onclick = () => toggleSidebar();
            document.body.appendChild(overlay);
        }
        overlay.classList.add('show');
    } else if (overlay) {
        overlay.classList.remove('show');
    }
}

// Formatting
function formatNum(n) {
    if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M';
    if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K';
    return n.toString();
}

function formatUptime(s) {
    const h = Math.floor(s / 3600);
    const m = Math.floor((s % 3600) / 60);
    return `${h}h ${m}m`;
}

// API calls using fetch
async function fetchData() {
    try {
        const res = await fetch('/api/health');
        const data = await res.json();
        updateUI(data);
        setOnlineStatus(true);
    } catch (e) {
        console.error('fetchData error:', e);
        setOnlineStatus(false);
    }
}

async function getConfig() {
    const res = await fetch('/api/config');
    return res.json();
}

async function saveConfig(config) {
    const res = await fetch('/api/config/save', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
    });
    return res.json();
}

async function setModel(model) {
    const res = await fetch('/api/model', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ model })
    });
    return res.json();
}

async function setAPIKey(key) {
    const res = await fetch('/api/apikey', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key })
    });
    return res.json();
}

async function resetStats() {
    const res = await fetch('/api/stats/reset', { method: 'POST' });
    return res.json();
}

async function startTunnelAPI() {
    const res = await fetch('/api/tunnel/start', { method: 'POST' });
    return res.json();
}

async function stopTunnelAPI() {
    const res = await fetch('/api/tunnel/stop', { method: 'POST' });
    return res.json();
}

function setOnlineStatus(isOnline) {
    const dot = document.getElementById('statusDot');
    const text = document.getElementById('statusText');

    if (isOnline) {
        dot.style.background = 'var(--success)';
        dot.style.animation = 'pulse 2s infinite';
        text.textContent = 'Online';
    } else {
        dot.style.background = 'var(--error)';
        dot.style.animation = 'none';
        text.textContent = 'Offline';
    }
}

function updateUI(data) {
    // Stats
    document.getElementById('totalReq').innerText = data.stats.messageCount;
    document.getElementById('errCount').innerText = data.stats.errorCount;

    const rate = data.stats.messageCount > 0
        ? ((data.stats.errorCount / data.stats.messageCount) * 100).toFixed(2)
        : '0.00';
    const errorRateEl = document.getElementById('errorRate');
    errorRateEl.innerText = rate + '%';
    errorRateEl.className = 'stat-value ' + (parseFloat(rate) > 5 ? 'error' : '');

    document.getElementById('tokenUsage').innerText = formatNum(data.stats.totalTokens);
    document.getElementById('promptTok').innerText = data.stats.promptTokens.toLocaleString();
    document.getElementById('compTok').innerText = data.stats.completionTokens.toLocaleString();
    document.getElementById('totalTok').innerText = data.stats.totalTokens.toLocaleString();

    document.getElementById('uptimeDisplay').innerText = formatUptime(data.uptime);
    document.getElementById('uptimeStat').innerText = Math.floor(data.uptime / 3600) + 'h';

    document.getElementById('lastReq').innerText = data.stats.lastRequestTime
        ? new Date(data.stats.lastRequestTime).toLocaleString()
        : '-';
    document.getElementById('sessionStart').innerText = data.stats.startTime
        ? new Date(data.stats.startTime).toLocaleString()
        : '-';

    // Model display only - skip if save in progress
    if (!settingsSaveInProgress && document.activeElement.id !== 'modelName') {
        document.getElementById('modelName').value = data.config.currentModel || '';
    }
    document.getElementById('currentModelDisplay').innerText = data.model || '-';

    // Tunnel
    updateTunnelUI(data.tunnel.status, data.tunnel.url);

    // Error Log
    updateErrorLog(data.stats.errorLog || []);
}

function updateTunnelUI(status, url) {
    const dot = document.getElementById('tunnelDot');
    const statusText = document.getElementById('tunnelStatus');
    const urlEl = document.getElementById('tunnelUrl');
    const startBtn = document.getElementById('startTunnelBtn');
    const stopBtn = document.getElementById('stopTunnelBtn');

    if (status === 'running' && url) {
        dot.style.background = 'var(--success)';
        dot.style.animation = 'pulse 2s infinite';
        statusText.textContent = 'Running';
        urlEl.textContent = url + '/v1/chat/completions';
        urlEl.classList.remove('hidden');
        startBtn.classList.add('hidden');
        stopBtn.classList.remove('hidden');
    } else if (status === 'starting') {
        dot.style.background = 'var(--warning)';
        dot.style.animation = 'pulse 1s infinite';
        statusText.textContent = 'Starting...';
        urlEl.classList.add('hidden');
    } else {
        dot.style.background = 'var(--text-muted)';
        dot.style.animation = 'none';
        statusText.textContent = 'Stopped';
        urlEl.classList.add('hidden');
        startBtn.classList.remove('hidden');
        stopBtn.classList.add('hidden');
    }
}

function updateErrorLog(logs) {
    const container = document.getElementById('errorLog');
    document.getElementById('errLogCount').innerText = logs.length;

    if (logs.length === 0) {
        container.innerHTML = '<div style="text-align: center; padding: 20px; color: var(--text-muted); font-size: 13px;">No errors recorded</div>';
        return;
    }

    const html = logs.slice(0, 15).map(e => `
        <div class="log-item">
            <span class="log-time">${new Date(e.timestamp).toLocaleTimeString()}</span>
            <span class="log-code">${e.code}</span>
            <span class="log-msg">${e.message}</span>
        </div>
    `).join('');

    container.innerHTML = html;
}

// AUTO-SAVE SETTINGS

document.getElementById('temperature').addEventListener('input', async (e) => {
    const value = parseFloat(e.target.value);
    document.getElementById('tempVal').innerText = value;

    try {
        const config = await getConfig();
        config.temperature = value;
        await saveConfig(config);
    } catch (error) {
        console.error('Failed to save temperature:', error);
    }
});

document.getElementById('contextSizeSlider').addEventListener('input', async (e) => {
    const value = parseInt(e.target.value);
    document.getElementById('contextSize').value = value;

    try {
        const config = await getConfig();
        config.contextSize = value;
        await saveConfig(config);
    } catch (error) {
        console.error('Failed to save context size:', error);
    }
});

document.getElementById('contextSize').addEventListener('input', async (e) => {
    const value = Math.min(128000, Math.max(0, parseInt(e.target.value) || 128000));
    document.getElementById('contextSizeSlider').value = value;

    try {
        const config = await getConfig();
        config.contextSize = value;
        await saveConfig(config);
    } catch (error) {
        console.error('Failed to save context size:', error);
    }
});

document.getElementById('maxTokensSlider').addEventListener('input', async (e) => {
    const value = parseInt(e.target.value);
    document.getElementById('maxTokens').value = value;

    try {
        const config = await getConfig();
        config.maxTokens = value;
        await saveConfig(config);
    } catch (error) {
        console.error('Failed to save max tokens:', error);
    }
});

document.getElementById('maxTokens').addEventListener('input', async (e) => {
    const value = Math.min(1000, Math.max(0, parseInt(e.target.value) || 0));
    document.getElementById('maxTokensSlider').value = value;

    try {
        const config = await getConfig();
        config.maxTokens = value;
        await saveConfig(config);
    } catch (error) {
        console.error('Failed to save max tokens:', error);
    }
});

async function saveToggleSetting(settingName, value) {
    try {
        const config = await getConfig();
        config[settingName] = value;
        await saveConfig(config);
    } catch (error) {
        console.error(`Failed to save ${settingName}:`, error);
        showToast(`Failed to save ${settingName}`, 'error');
    }
}

document.getElementById('showReasoning').addEventListener('change', (e) => {
    saveToggleSetting('showReasoning', e.target.checked);
});

document.getElementById('enableThinking').addEventListener('change', (e) => {
    saveToggleSetting('enableThinking', e.target.checked);
});

document.getElementById('logRequests').addEventListener('change', (e) => {
    saveToggleSetting('logRequests', e.target.checked);
});

document.getElementById('streamingEnabled').addEventListener('change', (e) => {
    saveToggleSetting('streamingEnabled', e.target.checked);
});

async function saveSettings() {
    const config = {
        showReasoning: document.getElementById('showReasoning').checked,
        enableThinking: document.getElementById('enableThinking').checked,
        logRequests: document.getElementById('logRequests').checked,
        streamingEnabled: document.getElementById('streamingEnabled').checked,
        contextSize: parseInt(document.getElementById('contextSize').value) || 128000,
        maxTokens: parseInt(document.getElementById('maxTokens').value) || 0,
        temperature: parseFloat(document.getElementById('temperature').value)
    };

    try {
        const result = await saveConfig(config);
        if (result.success) {
            showToast('Settings saved', 'success');
        } else {
            showToast('Failed to save settings', 'error');
        }
    } catch (e) {
        showToast('Failed to save settings', 'error');
    }
}

async function saveModelNow(model) {
    try {
        await setModel(model);
    } catch (e) {
        showToast('Failed to save model', 'error');
    }
    settingsSaveInProgress = false;
}

async function saveModel() {
    const model = document.getElementById('modelName').value.trim();
    if (!model) {
        showToast('Please enter a model name', 'error');
        return;
    }

    try {
        const result = await setModel(model);
        if (result.success) {
            showToast('Model updated', 'success');
            fetchData();
        } else {
            showToast('Failed to update model', 'error');
        }
    } catch (e) {
        showToast('Failed to update model', 'error');
    }
}

async function saveApiKey() {
    const key = document.getElementById('apiKey').value.trim();
    if (!key) {
        showToast('Please enter an API key', 'error');
        return;
    }

    try {
        const result = await setAPIKey(key);
        if (result.success) {
            showToast('API Key updated', 'success');
            document.getElementById('apiKey').value = '';
        } else {
            showToast('Failed to update API key', 'error');
        }
    } catch (e) {
        showToast('Failed to update API key', 'error');
    }
}

async function startTunnel() {
    updateTunnelUI('starting', null);
    try {
        const result = await startTunnelAPI();
        if (result.success) {
            showToast('Starting tunnel...', 'info');
        } else {
            showToast(result.error || 'Failed to start tunnel', 'error');
            updateTunnelUI('stopped', null);
        }
    } catch (e) {
        showToast('Failed to start tunnel', 'error');
        updateTunnelUI('stopped', null);
    }
}

async function stopTunnel() {
    try {
        await stopTunnelAPI();
        showToast('Tunnel stopped', 'info');
        fetchData();
    } catch (e) {
        showToast('Failed to stop tunnel', 'error');
    }
}

async function resetStatsBtn() {
    const confirmed = await showConfirm('Are you sure you want to reset all statistics?');
    if (!confirmed) return;

    try {
        await resetStats();
        showToast('Statistics reset', 'success');
        fetchData();
    } catch (e) {
        showToast('Failed to reset stats', 'error');
    }
}

function refreshData() {
    fetchData();
    showToast('Data refreshed', 'info');
}

// Window controls (no-op on mobile/browser)
function minimize() { }
function maximize() { }
function closeWindow() { }

// NIM Models list
const NIM_MODELS = [
    "deepseek-ai/deepseek-v3.2",
    "deepseek-ai/deepseek-v3.1-terminus",
    "deepseek-ai/deepseek-v3.1",
    "deepseek-ai/deepseek-r1",
    "deepseek-ai/deepseek-r1-0528",
    "deepseek-ai/deepseek-r1-distill-llama-8b",
    "deepseek-ai/deepseek-r1-distill-qwen-32b",
    "deepseek-ai/deepseek-r1-distill-qwen-14b",
    "deepseek-ai/deepseek-r1-distill-qwen-7b",
    "mistralai/mistral-large-3-675b-instruct-2512",
    "mistralai/mistral-medium-3-instruct",
    "mistralai/mistral-small-3.1-24b-instruct-2503",
    "mistralai/mistral-small-24b-instruct",
    "mistralai/mistral-7b-instruct-v0.3",
    "mistralai/mixtral-8x22b-instruct-v0.1",
    "mistralai/mixtral-8x7b-instruct-v0.1",
    "nvidia/llama-3.3-nemotron-super-49b-v1.5",
    "nvidia/llama-3.1-nemotron-ultra-253b-v1",
    "nvidia/llama-3.1-nemotron-nano-8b-v1",
    "meta/llama-3.3-70b-instruct",
    "meta/llama-3.2-90b-vision-instruct",
    "meta/llama-3.1-405b-instruct",
    "meta/llama-3.1-70b-instruct",
    "meta/llama-3.1-8b-instruct",
    "google/gemma-3-27b-it",
    "google/gemma-2-27b-it",
    "google/gemma-2-9b-it",
    "microsoft/phi-4-mini-instruct",
    "microsoft/phi-3.5-mini-instruct",
    "qwen/qwen3-235b-a22b",
    "qwen/qwq-32b",
    "qwen/qwen2.5-coder-32b-instruct",
    "moonshotai/kimi-k2-instruct"
];

// Custom dropdown functionality
function initModelDropdown() {
    const input = document.getElementById('modelName');
    const list = document.getElementById('modelList');
    let highlightedIndex = -1;

    function renderList(filter = '') {
        const filterLower = filter.toLowerCase();
        const filtered = NIM_MODELS.filter(m => m.toLowerCase().includes(filterLower));

        list.innerHTML = filtered.map((m, i) =>
            `<div class="dropdown-item" data-value="${m}" data-index="${i}">${m}</div>`
        ).join('');

        highlightedIndex = -1;
    }

    function showDropdown() {
        renderList(input.value);
        list.classList.add('show');
    }

    function hideDropdown() {
        list.classList.remove('show');
        highlightedIndex = -1;
    }

    function selectItem(value) {
        settingsSaveInProgress = true;
        input.value = value;
        hideDropdown();
        saveModelNow(value);
    }

    input.addEventListener('blur', () => {
        const model = input.value.trim();
        if (model) {
            settingsSaveInProgress = true;
            saveModelNow(model);
        }
    });

    input.addEventListener('focus', showDropdown);
    input.addEventListener('input', () => renderList(input.value));

    input.addEventListener('blur', (e) => {
        setTimeout(() => hideDropdown(), 150);
    });

    input.addEventListener('keydown', (e) => {
        const items = list.querySelectorAll('.dropdown-item');

        if (e.key === 'ArrowDown') {
            e.preventDefault();
            highlightedIndex = Math.min(highlightedIndex + 1, items.length - 1);
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            highlightedIndex = Math.max(highlightedIndex - 1, 0);
        } else if (e.key === 'Enter' && highlightedIndex >= 0) {
            e.preventDefault();
            const item = items[highlightedIndex];
            if (item) selectItem(item.dataset.value);
        } else if (e.key === 'Escape') {
            hideDropdown();
        }

        items.forEach((item, i) => {
            item.classList.toggle('highlighted', i === highlightedIndex);
            if (i === highlightedIndex) item.scrollIntoView({ block: 'nearest' });
        });
    });

    list.addEventListener('click', (e) => {
        const item = e.target.closest('.dropdown-item');
        if (item) selectItem(item.dataset.value);
    });

    renderList();
}

// Load initial settings from backend
async function loadInitialSettings() {
    try {
        const config = await getConfig();
        document.getElementById('showReasoning').checked = config.showReasoning;
        document.getElementById('enableThinking').checked = config.enableThinking;
        document.getElementById('logRequests').checked = config.logRequests;
        document.getElementById('streamingEnabled').checked = config.streamingEnabled;
        document.getElementById('contextSize').value = config.contextSize || 128000;
        document.getElementById('contextSizeSlider').value = config.contextSize || 128000;
        document.getElementById('maxTokens').value = config.maxTokens || 0;
        document.getElementById('maxTokensSlider').value = config.maxTokens || 0;
        document.getElementById('temperature').value = config.temperature;
        document.getElementById('tempVal').innerText = config.temperature;
    } catch (error) {
        console.error('Failed to load initial settings:', error);
    }
}

// Initialize
loadInitialSettings();
setInterval(fetchData, 2000);
fetchData();
initModelDropdown();
