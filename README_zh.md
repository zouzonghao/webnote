# WebNote

一个简单、快速、临时的笔记服务。通过 URL 轻松分享笔记。

## 特性

- **简洁界面**: 一个干净直观的界面，用于编写和阅读笔记。
- **临时存储**: 笔记存储在服务器上，但总存储空间有限。旧的笔记可能会被清除。
- **基于 URL 的分享**: 每个笔记都会获得一个独特的、随机生成的 URL，方便分享。
- **速率限制**: 保护服务免受滥用。
- **原始文本访问**: 通过 `?raw` 查询参数或使用 `curl`、`Wget` 等客户端访问任何笔记的纯文本内容。

## 使用方法

1.  访问根 URL。您将被重定向到一个新的空白笔记页面。
2.  在文本区域写下您的笔记。
3.  笔记会自动保存。
4.  与他人分享该 URL。

## API 用法

您可以使用 `curl` 或其他 HTTP 客户端通过编程方式与 WebNote 交互。

### 查看笔记

-   **URL**: `GET /{note_path}`
-   **示例**: `curl http://127.0.0.1:8080/mynote?raw=true`

查看raw文本内容：

-   **URL**: `GET /{note_path}?raw=true`
-   **Example**: `curl http://127.0.0.1:8080/mynote?raw=true`

### 保存或更新笔记

-   **URL**: `POST /save/{note_path}`
-   **请求体**: 您的笔记内容。

**示例:**

-   从原始文本保存：
    ```bash
    curl -X POST -d "这是我的笔记。" http://127.0.0.1:8080/save/mynote
    ```
-   从文件保存：
    ```bash
    curl -X POST --data-binary "@path/to/your/file.txt" http://127.0.0.1:8080/save/mynote
    ```

### 删除笔记

要删除一个笔记，向其保存 URL 发送一个空的 POST 请求。

-   **URL**: `POST /save/{note_path}`
-   **请求体**: (空)
-   **示例**:
    ```bash
    curl -X POST -d "" http://127.0.0.1:8080/save/mynote
    ```

## 开发

在本地运行此项目：

```bash
go run main.go
```

默认情况下，服务器将在 `8080` 端口上启动。