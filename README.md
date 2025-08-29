# WebNote

[简体中文](README_zh.md)

A simple, fast, and ephemeral note-taking service. Share notes easily with a URL.

## Features

- **Simple Interface**: A clean and intuitive UI for writing and reading notes.
- **Ephemeral Storage**: Notes are stored on the server, but the total storage is limited. Old notes may be cleared.
- **URL-based Sharing**: Every note gets a unique, randomly generated URL for easy sharing.
- **Rate Limiting**: Protects the service from abuse.
- **Raw Content Access**: Access the raw text of any note via `?raw` query parameter or by using clients like `curl` or `Wget`.

## Usage

1.  Navigate to the root URL. You will be redirected to a new, empty note page.
2.  Write your note in the text area.
3.  The note is saved automatically.
4.  Share the URL with others.

## API Usage

You can interact with WebNote programmatically using `curl` or other HTTP clients.

### View a Note

-   **URL**: `GET /{note_path}`
-   **Example**: `curl http://127.0.0.1:8080/mynote`

To get the raw text content:

-   **URL**: `GET /{note_path}?raw=true`
-   **Example**: `curl http://127.0.0.1:8080/mynote?raw=true`

### Save or Update a Note

-   **URL**: `POST /save/{note_path}`
-   **Body**: The content of your note.

**Examples:**

-   Save from raw text:
    ```bash
    curl -X POST -d "This is my note." http://127.0.0.1:8080/save/mynote
    ```
-   Save from a file:
    ```bash
    curl -X POST --data-binary "@path/to/your/file.txt" http://127.0.0.1:8080/save/mynote
    ```

### Delete a Note

To delete a note, send an empty POST request to its save URL.

-   **URL**: `POST /save/{note_path}`
-   **Body**: (empty)
-   **Example**:
    ```bash
    curl -X POST -d "" http://127.0.0.1:8080/save/mynote
    ```

## Development

To run this project locally:

```bash
go run main.go
```

The server will start on port `8080` by default.