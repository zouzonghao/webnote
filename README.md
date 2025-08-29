# 📝 WebNote

[简体中文](README_zh.md)

A blazingly fast, ephemeral, Git-powered note-taking service. Share notes and ideas with just a URL.

---

## ✨ Core Features

-   ⚡️ **Blazing Fast & Simple:** No database, no complex setup. Just pure performance with Go.
-   🔗 **Share via URL:** Every note is a unique, shareable link, generated on the fly.
-   ✍️ **Real-time Sync:** Keep your notes synced across multiple devices or collaborate with others seamlessly using WebSockets.
-   🕰️ **Git-Powered History:** Every save is a `git commit`! Browse your note's history with ease, just like your code.
-   🗑️ **Ephemeral by Design:** Old, inactive note histories are automatically pruned to keep storage lean and focused on recent activity.
-    RAW **Raw Text Access:** `curl` or `wget` your notes directly from the terminal for ultimate scripting power.

---

## 🚀 Getting Started: Deployment

The quickest and recommended way to get WebNote running is with Docker.

### Method 1: Docker Compose (Recommended)

This method uses the pre-built image from Docker Hub, which is perfect for most users.

1.  Ensure you have Docker and Docker Compose installed.
2.  Save the `docker-compose.yml` file from this repository.
3.  Run the following command in the same directory:
    ```bash
    docker-compose up -d
    ```
    Your WebNote instance is now running at `http://localhost:8080`. Note data will be stored in a `./notes-data` directory.

### Method 2: Build from Source

If you prefer to build the image yourself:

1.  **With Docker:**
    ```bash
    # 1. Build the image
    docker build -t webnote .

    # 2. Run the container
    docker run -d -p 8080:8080 -v $(pwd)/notes-data:/app/notes --name webnote_app webnote
    ```

2.  **With Docker Compose:**
    To build from source using Docker Compose, you need to add one line to the `docker-compose.yml` file:
    ```yaml
    services:
      webnote:
        build: . # <-- Add this line
        image: sanqi37/webnote:latest
        # ... rest of the file
    ```
    Then, run the build command:
    ```bash
    docker-compose up -d --build
    ```

---

## 👩‍💻 Development

Want to contribute or run the project locally without Docker?

1.  **Prerequisites:**
    -   Go (version 1.20+)
    -   Node.js & npm (for frontend asset minification)

2.  **Run the Backend:**
    ```bash
    go run main.go
    ```
    The server will start on `http://localhost:8080`.

3.  **Frontend Assets:**
    If you modify `static/script.js` or `static/style.css`, you need to minify them. A helper script is provided for convenience.
    ```bash
    # Make the script executable (only needs to be done once)
    chmod +x minify.sh

    # Run the script to install dependencies and minify files
    ./minify.sh
    ```

---

## ⚙️ Configuration

WebNote is configured via environment variables.

| Variable              | Description                                                  | Default        |
| --------------------- | ------------------------------------------------------------ | -------------- |
| `MAX_STORAGE_SIZE`    | The total size limit for all notes in bytes.                 | `10240000` (10MB) |
| `MAX_CONTENT_SIZE`    | The size limit for a single note in bytes.                   | `102400` (100KB) |
| `HISTORY_RESET_HOURS` | The inactivity threshold in hours to reset a note's history. | `72`           |
| `PORT`                | The port the server will listen on.                          | `8080`         |

---

## 🔌 API Usage

Interact with WebNote programmatically.

-   **View a Note (Raw):**
    ```bash
    curl http://localhost:8080/your-note?raw=true
    ```
-   **Save/Update a Note:**
    ```bash
    curl -X POST -d "Hello, this is my note." http://localhost:8080/save/your-note
    ```
-   **Delete a Note:**
    ```bash
    curl -X POST -d "" http://localhost:8080/save/your-note
