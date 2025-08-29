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

## 部署

### 使用 Docker

1.  **从源代码构建镜像:**
    ```bash
    docker build -t webnote .
    ```

2.  **运行容器:**
    此命令会将笔记数据存储在当前工作目录下的 `notes-data` 目录中。
    ```bash
    docker run -d -p 8080:8080 -v $(pwd)/notes-data:/app/notes --name webnote_app webnote
    ```

### 使用 Docker Compose

项目提供的 `docker-compose.yml` 文件使用的是 Docker Hub 上的预构建镜像。

1.  **使用预构建镜像运行:**
    ```bash
    docker-compose up -d
    ```

2.  **从源代码构建并运行:**
    如果您想从本地的 `Dockerfile` 构建镜像，可以修改 `docker-compose.yml` 文件，添加 `build` 指令：
    ```yaml
    version: '3.8'

    services:
      webnote:
        build: . # 添加此行
        image: webnote # 可选：为镜像命名
        restart: unless-stopped
        container_name: webnote_app
        ports:
          - "8080:8080"
        volumes:
          - ./notes-data:/app/notes
        user: root
        environment:
          - MAX_STORAGE_SIZE=10240000
          - MAX_CONTENT_SIZE=102400
    ```
    然后运行：
    ```bash
    docker-compose up -d --build
    ```