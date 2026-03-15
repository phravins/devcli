document.addEventListener('DOMContentLoaded', () => {
    const codeEditor = document.getElementById('code');
    const filenameInput = document.getElementById('filename');
    const outputLog = document.getElementById('output-log');
    const termInput = document.getElementById('term-input');
    const termLog = document.getElementById('terminal-log');
    const saveStatus = document.getElementById('save-status');

    // --- State Management ---
    let currentUser = null;
    let isAutoSaveEnabled = true;

    // --- Tab Switching ---
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');

    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const tabId = btn.getAttribute('data-tab');
            
            tabBtns.forEach(b => b.classList.remove('active'));
            tabContents.forEach(c => c.classList.remove('active'));

            btn.classList.add('active');
            document.getElementById(`${tabId}-view`).classList.add('active');

            if (tabId === 'terminal') {
                termInput.focus();
            }
        });
    });

    // --- Resizer Logic ---
    const resizer = document.getElementById('resizer');
    const editorPane = document.querySelector('.editor-pane');
    const bottomPane = document.querySelector('.bottom-pane');
    let isResizing = false;

    resizer.addEventListener('mousedown', (e) => {
        isResizing = true;
        document.body.style.cursor = 'row-resize';
    });

    document.addEventListener('mousemove', (e) => {
        if (!isResizing) return;
        const containerHeight = document.querySelector('.split-container').offsetHeight;
        const topHeight = e.clientY - document.querySelector('.split-container').offsetTop;
        const topFlex = topHeight / containerHeight * 5; // Simplified flex calc
        
        if (topFlex > 0.5 && topFlex < 4.5) {
            editorPane.style.flex = topFlex;
            bottomPane.style.flex = 5 - topFlex;
        }
    });

    document.addEventListener('mouseup', () => {
        isResizing = false;
        document.body.style.cursor = 'default';
    });

    // --- Auth Modal ---
    const authModal = document.getElementById('auth-modal');
    const loginTrigger = document.getElementById('login-trigger');
    const closeAuth = document.getElementById('close-auth');
    const modalTabs = document.querySelectorAll('.modal-tab');
    const loginForm = document.getElementById('login-form');
    const signupForm = document.getElementById('signup-form');

    loginTrigger.addEventListener('click', () => authModal.classList.remove('hidden'));
    closeAuth.addEventListener('click', () => authModal.classList.add('hidden'));

    modalTabs.forEach(tab => {
        tab.addEventListener('click', () => {
            modalTabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
            const mode = tab.getAttribute('data-mode');
            if (mode === 'login') {
                loginForm.classList.remove('hidden');
                signupForm.classList.add('hidden');
            } else {
                loginForm.classList.add('hidden');
                signupForm.classList.remove('hidden');
            }
        });
    });

    // --- Local Storage Sync ---
    const loadFromLocal = () => {
        const savedCode = localStorage.getItem('devcli_autosave_code');
        const savedFile = localStorage.getItem('devcli_autosave_filename');
        if (savedCode) codeEditor.value = savedCode;
        if (savedFile) filenameInput.value = savedFile;
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
        
        // Switch to output tab
        document.querySelector('[data-tab="output"]').click();
        log.textContent = "Executing...";
        log.style.color = 'var(--text-secondary)';

        try {
            const response = await fetch('/run', {
                method: 'POST',
                body: code
            });
            const result = await response.json();
            
            if (result.error) {
                log.style.color = 'var(--error)';
                log.textContent = (result.output || "") + "\nError: " + result.error;
            } else {
                log.style.color = 'var(--success)';
                log.textContent = result.output;
            }
        } catch (e) {
            log.style.color = 'var(--error)';
            log.textContent = "Execution Failed: " + e.message;
        }
    };

    window.saveCode = async () => {
        const code = codeEditor.value;
        const filename = filenameInput.value;
        
        if (!filename) {
            showToast("Filename required!", "error");
            return;
        }

        try {
            const response = await fetch('/save', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ filename, content: code })
            });
            
            if (response.ok) {
                showToast("Saved successfully to " + filename, "success");
            } else {
                const txt = await response.text();
                showToast("Error: " + txt, "error");
            }
        } catch (e) {
            showToast("Network Error: " + e.message, "error");
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
                
                // Scroll to bottom
                const container = document.getElementById('terminal-view');
                container.scrollTop = container.scrollHeight;
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

    // --- Auth Handlers (Placeholders) ---
    loginForm.addEventListener('submit', (e) => {
        e.preventDefault();
        showToast("Login logic pending backend...", "info");
    });

    signupForm.addEventListener('submit', (e) => {
        e.preventDefault();
        showToast("Registration logic pending backend...", "info");
    });

    document.getElementById('google-auth-btn').addEventListener('click', () => {
        showToast("Connecting to Google...", "info");
        window.location.href = '/auth/google';
    });

    document.getElementById('drive-btn').addEventListener('click', () => {
        showToast("Google Drive integration coming soon.", "info");
    });
});
