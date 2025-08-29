const textarea = document.getElementById('content');
const printable = document.getElementById('printable');

// --- Configuration ---
const AUTOSAVE_DELAY = 30000; // Autosave delay in milliseconds (30 seconds)
let lastSavedContent = textarea.value;
let socket;
let saveTimeout = null;
let toastTimeout = null;

// --- WebSocket Logic ---

function connect() {
    const path = window.location.pathname.substring(1);
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/ws/${path}`;

    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
        console.log('WebSocket connection established');
    };

    socket.onmessage = (event) => {
        const newContent = event.data;
        const shouldBlockUpdate = document.hasFocus() && document.activeElement === textarea;

        if (textarea.value !== newContent && !shouldBlockUpdate) {
            textarea.value = newContent;
            lastSavedContent = newContent;
            updatePrintable(newContent);
        }
    };

    socket.onclose = () => {
        console.log('WebSocket connection closed. Attempting to reconnect in 2 seconds...');
        setTimeout(connect, 2000);
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        socket.close();
    };
}

// --- Saving Logic ---

function saveContent() {
    clearTimeout(saveTimeout); // Clear any pending autosave
    const currentContent = textarea.value;
    if (lastSavedContent.trim() === currentContent.trim()) {
        return; // No meaningful changes
    }

    const path = window.location.pathname.substring(1);
    const data = new URLSearchParams();
    data.append('content', currentContent);

    fetch(`/save/${path}`, {
        method: 'POST',
        body: data,
        keepalive: true
    }).then(response => {
        if (!response.ok) {
            return response.text().then(text => { throw new Error(text) });
        }
        return response.text();
    }).then(() => {
        lastSavedContent = currentContent;
        updatePrintable(currentContent);
    }).catch(error => {
        let errorMessage = error.message;
        // Intercept the generic "Failed to fetch" error to provide a better message.
        if (error instanceof TypeError && error.message === 'Failed to fetch') {
            errorMessage = 'Failed to save. The note might be too large or there was a network issue.';
        }
        showToast(errorMessage || "Failed to save note.");
    });
}

function scheduleSave() {
    clearTimeout(saveTimeout);
    saveTimeout = setTimeout(saveContent, AUTOSAVE_DELAY);
}

// --- Event Listeners ---

// Autosave on input
textarea.addEventListener('input', scheduleSave);

// Save on newline
textarea.addEventListener('keyup', (event) => {
    if (event.key === 'Enter') {
        saveContent(); // This also clears the timeout
    }
});

// Save before leaving the page
window.addEventListener('beforeunload', () => {
    saveContent(); // This also clears the timeout
});


// --- UI Update ---

function updatePrintable(content) {
    printable.textContent = content;
}

function showToast(message) {
    const toast = document.getElementById('toast');
    if (!toast) return;

    toast.textContent = message;
    toast.className = "show";

    clearTimeout(toastTimeout);
    toastTimeout = setTimeout(function(){ toast.className = toast.className.replace("show", ""); }, 3000);
}

// --- Initial setup ---
textarea.focus();
updatePrintable(textarea.value);
connect();