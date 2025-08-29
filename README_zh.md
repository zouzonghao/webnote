# 📝 WebNote

[English Version](README.md)

一个快如闪电、即用即走、由 Git 驱动的笔记服务。仅需一个 URL，即可分享笔记与想法。

---

## ✨ 核心特性

-   ⚡️ **极致性能 & 简单:** 无需数据库，没有复杂配置。纯 Go 语言带来极致性能。
-   🔗 **URL 分享:** 每个笔记都是一个独特的、可分享的链接，即时生成。
-   ✍️ **实时同步:** 使用 WebSocket 技术，让你的笔记在多个设备间无缝同步，或与他人轻松协作。
-   🕰️ **Git 驱动的版本历史:** 每一次保存都是一次 `git commit`！像浏览代码一样轻松查看笔记的修改历史。
-   🗑️ **临时化设计:** 旧的、不活跃的笔记历史会被自动修剪，以保持存储的精简，并专注于近期活动。
-   🦾 **命令行访问:** 通过 `curl` 或 `wget` 直接从终端访问笔记的纯文本内容，实现强大的脚本化操作。

---

## 🚀 快速开始：部署

推荐使用 Docker 来快速启动 WebNote。

### 方法一：使用 Docker Compose (推荐)

此方法使用 Docker Hub 上的预构建镜像，适合绝大多数用户。

1.  确保您已安装 Docker 和 Docker Compose。
2.  保存本项目中的 `docker-compose.yml` 文件。
3.  在文件所在目录运行以下命令：
    ```bash
    docker-compose up -d
    ```
    现在，您的 WebNote 实例已在 `http://localhost:8080` 运行。笔记数据将被存储在 `./notes-data` 目录中。

### 方法二：从源代码构建

如果您希望自己构建镜像：

1.  **使用 Docker:**
    ```bash
    # 1. 构建镜像
    docker build -t webnote .

    # 2. 运行容器
    docker run -d -p 8080:8080 -v $(pwd)/notes-data:/app/notes --name webnote_app webnote
    ```

2.  **使用 Docker Compose:**
    要使用 Docker Compose 从源代码构建，您需要在 `docker-compose.yml` 文件中添加一行：
    ```yaml
    services:
      webnote:
        build: . # <-- 添加此行
        image: sanqi37/webnote:latest
        # ... 文件其余部分
    ```
    然后，运行构建命令：
    ```bash
    docker-compose up -d --build
    ```

---

## 👩‍💻 本地开发

想要贡献代码或在本地（不使用 Docker）运行项目？

1.  **环境要求:**
    -   Go (版本 1.20+)
    -   Node.js & npm (用于压缩前端静态资源)

2.  **运行后端:**
    ```bash
    go run main.go
    ```
    服务器将在 `http://localhost:8080` 启动。

3.  **前端资源:**
    如果您修改了 `static/script.js` 或 `static/style.css`，您需要将它们压缩。项目提供了一个方便的辅助脚本。
    ```bash
    # 赋予脚本可执行权限 (只需执行一次)
    chmod +x minify.sh

    # 运行脚本来安装依赖并压缩文件
    ./minify.sh
    ```

---

## ⚙️ 配置选项

WebNote 通过环境变量进行配置。

| 环境变量            | 描述                                               | 默认值         |
| --------------------- | -------------------------------------------------- | -------------- |
| `MAX_STORAGE_SIZE`    | 所有笔记的总大小限制（字节）。                     | `10240000` (10MB) |
| `MAX_CONTENT_SIZE`    | 单个笔记的大小限制（字节）。                       | `102400` (100KB) |
| `HISTORY_RESET_HOURS` | 笔记历史重置前的不活跃时间阈值（小时）。           | `72`           |
| `PORT`                | 服务器监听的端口。                                 | `8080`         |

---

## 🔌 API 用法

通过编程方式与 WebNote 交互。

-   **查看笔记 (纯文本):**
    ```bash
    curl http://localhost:8080/your-note?raw=true
    ```
-   **保存/更新笔记:**
    ```bash
    curl -X POST -d "你好，这是我的笔记。" http://localhost:8080/save/your-note
    ```
-   **删除笔记:**
    ```bash
    curl -X POST -d "" http://localhost:8080/save/your-note
