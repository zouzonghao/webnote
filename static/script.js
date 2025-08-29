const textarea = document.getElementById('content');
const printable = document.getElementById('printable');
let timeout = null;
let lastSavedContent = textarea.value;

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
    timeout = setTimeout(saveContent, 1500); // 1.5 second delay
});

// Save before leaving the page
window.addEventListener('beforeunload', (event) => {
    clearTimeout(timeout); // Cancel any pending debounced save
    saveContent(); // Save immediately
});

// Initial setup
textarea.focus();
updatePrintable(textarea.value);