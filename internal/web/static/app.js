document.addEventListener('DOMContentLoaded', () => {
    const codeEditor = document.getElementById('code');
    const filenameInput = document.getElementById('filename');
    const outputLog = document.getElementById('output-log');
    const termInput = document.getElementById('term-input');
    const termLog = document.getElementById('terminal-log');
    const saveStatus = document.getElementById('save-status');

    // --- Settings Modal ---
    const settingsModal = document.getElementById('settings-modal');
    const settingsTrigger = document.getElementById('settings-trigger');
    const closeSettings = document.getElementById('close-settings');
    const settingsForm = document.getElementById('settings-form');

    settingsTrigger.addEventListener('click', () => settingsModal.classList.remove('hidden'));
    closeSettings.addEventListener('click', () => settingsModal.classList.add('hidden'));

    settingsForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const theme = document.getElementById('theme-select').value;
        const defaultPath = document.getElementById('default-save-path').value;
        
        localStorage.setItem('devcli_theme', theme);
        localStorage.setItem('devcli_default_path', defaultPath);
        
        showToast("Settings saved!", "success");
        settingsModal.classList.add('hidden');
        
        sendLogToTUI(`User updated settings: Theme=${theme}, Path=${defaultPath}`);
    });

    // --- Save Modal & Logic ---
    const saveModal = document.getElementById('save-modal');
    const closeSave = document.getElementById('close-save');
    const confirmSaveBtn = document.getElementById('confirm-save-btn');
    const customSavePathInput = document.getElementById('custom-save-path');
    const locationBtnTabs = document.querySelectorAll('.location-tabs .tab-btn');
    let saveLocation = 'local';

    window.saveCode = () => {
        const currentFilename = filenameInput.value;
        const defaultPath = localStorage.getItem('devcli_default_path') || "";
        customSavePathInput.value = currentFilename || defaultPath;
        saveModal.classList.remove('hidden');
    };

    closeSave.addEventListener('click', () => saveModal.classList.add('hidden'));

    locationBtnTabs.forEach(btn => {
        btn.addEventListener('click', () => {
            locationBtnTabs.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            saveLocation = btn.getAttribute('data-location');
            
            if (saveLocation === 'local') {
                document.getElementById('local-save-options').classList.remove('hidden');
                document.getElementById('drive-save-options').classList.add('hidden');
            } else {
                document.getElementById('local-save-options').classList.add('hidden');
                document.getElementById('drive-save-options').classList.remove('hidden');
            }
        });
    });

    confirmSaveBtn.addEventListener('click', async () => {
        const code = codeEditor.value;
        const path = customSavePathInput.value;

        if (saveLocation === 'local') {
            if (!path) {
                showToast("Path required!", "error");
                return;
            }
            try {
                const response = await fetch('/save', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ filename: path, content: code })
                });
                
                if (response.ok) {
                    showToast("Saved locally to " + path, "success");
                    filenameInput.value = path;
                    saveToLocal();
                    saveModal.classList.add('hidden');
                    sendLogToTUI(`File saved locally: ${path}`);
                } else {
                    const txt = await response.text();
                    showToast("Error: " + txt, "error");
                }
            } catch (e) {
                showToast("Network Error: " + e.message, "error");
            }
        } else {
            // Drive Save
            try {
                const response = await fetch('/drive/save', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ filename: path || "untitled.py", content: code })
                });
                if (response.ok) {
                    showToast("Saved to Google Drive!", "success");
                    saveModal.classList.add('hidden');
                    sendLogToTUI(`File saved to Google Drive: ${path || "untitled.py"}`);
                } else {
                    showToast("Drive Save Failed. Please connect account.", "error");
                }
            } catch (e) {
                showToast("Drive error: " + e.message, "error");
            }
        }
    });

    // --- Logging Sync to TUI ---
    const sendLogToTUI = async (msg) => {
        try {
            await fetch('/logs', {
                method: 'POST',
                body: msg
            });
        } catch (e) {
            console.error("Failed to sync log to TUI:", e);
        }
    };

    // --- Local Storage Sync Init ---
    const loadFromLocal = () => {
        const savedCode = localStorage.getItem('devcli_autosave_code');
        const savedFile = localStorage.getItem('devcli_autosave_filename');
        if (savedCode) codeEditor.value = savedCode;
        if (savedFile) filenameInput.value = savedFile;

        const savedPath = localStorage.getItem('devcli_default_path');
        if (savedPath) document.getElementById('default-save-path').value = savedPath;
    };

    const saveToLocal = () => {
        localStorage.setItem('devcli_autosave_code', codeEditor.value);
        localStorage.setItem('devcli_autosave_filename', filenameInput.value);
        saveStatus.textContent = 'Auto-saved locally';
    };

    codeEditor.addEventListener('input', saveToLocal);
    filenameInput.addEventListener('input', saveToLocal);
    loadFromLocal();

    // --- Execution Logic ---
    window.runCode = async () => {
        const code = codeEditor.value;
        const log = outputLog;
        
        document.querySelector('[data-tab="output"]').click();
        log.textContent = "Executing...";
        log.style.color = 'var(--text-secondary)';

        sendLogToTUI(`Started execution of ${filenameInput.value || 'unsaved script'}`);

        try {
            const response = await fetch('/run', {
                method: 'POST',
                body: code
            });
            const result = await response.json();
            
            if (result.error) {
                log.style.color = 'var(--error)';
                log.textContent = (result.output || "") + "\nError: " + result.error;
                sendLogToTUI(`Execution failed: ${result.error}`);
            } else {
                log.style.color = 'var(--success)';
                log.textContent = result.output;
                sendLogToTUI(`Execution completed successfully`);
            }
        } catch (e) {
            log.style.color = 'var(--error)';
            log.textContent = "Execution Failed: " + e.message;
            sendLogToTUI(`Network error during execution: ${e.message}`);
        }
    };

    window.clearOutput = () => {
        outputLog.textContent = "Ready...";
        outputLog.style.color = 'var(--text-secondary)';
    };

    // --- Terminal Logic ---
    termInput.addEventListener('keydown', async (e) => {
        if (e.key === 'Enter') {
            const cmd = termInput.value;
            if (!cmd.trim()) return;

            termLog.textContent += `\n$ ${cmd}\n`;
            termInput.value = '';
            termInput.disabled = true;

            try {
                const response = await fetch('/terminal', {
                    method: 'POST',
                    body: cmd
                });
                const result = await response.json();
                termLog.textContent += result.output + "\n";
                
                const container = document.getElementById('terminal-view');
                container.scrollTop = container.scrollHeight;
                
                sendLogToTUI(`Shell command run via web: ${cmd}`);
            } catch (e) {
                termLog.textContent += "Error executing command.\n";
            }
            
            termInput.disabled = false;
            termInput.focus();
        }
    });

    // --- UI Helpers ---
    const showToast = (msg, type = 'info') => {
        const toast = document.getElementById('toast');
        toast.textContent = msg;
        toast.className = `toast ${type}`;
        toast.classList.remove('hidden');
        setTimeout(() => toast.classList.add('hidden'), 3000);
    };

    document.getElementById('google-auth-btn').addEventListener('click', () => {
        showToast("Connecting to Google...", "info");
        window.location.href = '/auth/google';
    });

    document.getElementById('drive-sidebar-btn').addEventListener('click', () => {
        saveLocation = 'drive';
        saveModal.classList.remove('hidden');
        document.querySelectorAll('.location-tabs .tab-btn').forEach(b => {
            b.classList.toggle('active', b.getAttribute('data-location') === 'drive');
        });
        document.getElementById('local-save-options').classList.add('hidden');
        document.getElementById('drive-save-options').classList.remove('hidden');
    });
});
