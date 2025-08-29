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

## Development

To run this project locally:

```bash
go run main.go
```

The server will start on port `8080` by default.