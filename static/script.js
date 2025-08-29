const textarea = document.getElementById('content');
const printable = document.getElementById('printable');

// --- Configuration ---
const DEBOUNCE_DELAY = 500; // Debounce delay in milliseconds

let timeout = null;
let lastSavedContent = textarea.value;
let socket;

// --- WebSocket Logic ---

function connect() {
    const path = window.location.pathname.substring(1);
    // Construct WebSocket URL, handling http/https protocols
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/ws/${path}`;

    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
        console.log('WebSocket connection established');
    };

    socket.onmessage = (event) => {
        const newContent = event.data;

        // Determine if we should block the update to prevent overwriting user input.
        // This should only happen if the current browser tab is active AND the textarea has focus.
        const shouldBlockUpdate = document.hasFocus() && document.activeElement === textarea;

        if (textarea.value !== newContent && !shouldBlockUpdate) {
            textarea.value = newContent;
            lastSavedContent = newContent;
            updatePrintable(newContent);
        }
    };

    socket.onclose = () => {
        console.log('WebSocket connection closed. Attempting to reconnect in 2 seconds...');
        setTimeout(connect, 2000); // Attempt to reconnect after a delay
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        socket.close(); // This will trigger the onclose event and reconnection logic
    };
}


// --- Saving Logic ---

// Function to save content to the server
function saveContent() {
    const currentContent = textarea.value;
    if (lastSavedContent === currentContent) {
        return; // No changes, no need to save
    }

    const path = window.location.pathname.substring(1);
    const data = new URLSearchParams();
    data.append('content', currentContent);

    fetch(`/save/${path}`, {
        method: 'POST',
        body: data,
        keepalive: true // Important for unload events
    }).then(() => {
        lastSavedContent = currentContent;
        updatePrintable(currentContent);
    });
}

// Update printable content for printing
function updatePrintable(content) {
    while (printable.firstChild) {
        printable.removeChild(printable.firstChild);
    }
    printable.appendChild(document.createTextNode(content));
}

// Debounced save on input
textarea.addEventListener('input', () => {
    clearTimeout(timeout);
    timeout = setTimeout(saveContent, DEBOUNCE_DELAY);
});

// Save before leaving the page
window.addEventListener('beforeunload', (event) => {
    clearTimeout(timeout); // Cancel any pending debounced save
    saveContent(); // Save immediately
});

// --- Initial setup ---
textarea.focus();
updatePrintable(textarea.value);
connect(); // Establish WebSocket connection on page load