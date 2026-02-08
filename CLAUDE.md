# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

AfdianToMarkdown 是一个 Go CLI 工具，从爱发电 (afdian.com) 下载付费内容并保存为 Markdown 文件。需要用户通过 Cookie Master 浏览器扩展导出 `cookies.json` 才能访问 API。

许可证：AGPL-3.0

## 常用命令

```bash
# 构建
go build -o AfdianToMarkdown.exe .

# 运行（示例）
./AfdianToMarkdown motions -au <作者url_slug>
./AfdianToMarkdown album -u <作品集url>
./AfdianToMarkdown albums -au <作者url_slug>
./AfdianToMarkdown update    # 更新所有已下载作者

# 全局参数
# --host <域名>       主站域名（默认 afdian.com）
# --dir <路径>        数据存储目录（默认 程序目录/data）
# --cookie <路径>     cookies.json 路径（默认 程序目录/cookies.json）
# --disable_comment   不下载评论

# 测试（需要有效的 cookies.json，测试会实际调用 API）
go test ./afdian/ -run TestGetAuthorId -v
go test ./... -v

# 发布（通过 GitHub Actions 触发，打 v* tag 即可）
git tag v0.x.x && git push origin v0.x.x
```

## 架构

### 数据流

```
CLI (main.go) → config.Config 创建
              → afdian.GetCookies() 加载认证
              → motion/album 子包协调下载流程
              → afdian 包执行 API 调用 + JSON 解析 + HTML→MD 转换
              → 文件写入到 {DataDir}/{作者}/{motions|albumName}/
```

### 核心包

- **`config/`** — `Config` 结构体集中管理 Host、DataDir、CookiePath，通过参数传递（无全局状态）
- **`afdian/`** — API 客户端：HTTP 请求构建、Cookie 处理、JSON 解析（gjson）、HTML→Markdown 转换、文件保存
- **`afdian/motion/`** — 按作者下载所有动态，使用 `publish_sn` 游标分页
- **`afdian/album/`** — 按作者下载所有作品集，或下载单个作品集
- **`logger/`** — 自定义 slog Handler，带 ANSI 彩色输出
- **`utils/`** — 文件名安全转换、程序目录解析（`ResolveAppDir`）、默认路径推断、作者目录扫描

### 关键设计约定

- CLI 框架：`urfave/cli/v3`，子命令 + Before/After 钩子
- HTTP 客户端：`carlmjohnson/requests`，请求头模拟 Chrome 浏览器
- JSON 解析：`tidwall/gjson`（路径查询而非反序列化到结构体）
- API 限速：请求间 150ms 延迟（`afdian.DelayMs`）
- 文件命名：`{index}_{标题}.md`，不安全字符替换为下划线
- 图片下载到 `.assets/` 子目录，Markdown 中使用相对路径引用
- `SavePostIfNotExist` 跳过已存在的文件（幂等下载）

### 数据目录结构

```
{AppDataPath}/
├── data/
│   └── {作者url_slug}/
│       ├── motions/           # 动态
│       │   └── 0_标题.md
│       └── {作品集名}/        # 作品集
│           └── 0_标题.md
├── cookies.json
└── AfdianToMarkdown.exe
```

## 开发计划

正在进行 MCP Server 重构，分阶段：
- 第〇阶段：代码重构（消除全局状态 → 数据目录解耦 → 错误处理 → 拆分 afdian.go → 时间字段）
- 第一~四阶段：MCP Server 实现（基础框架 → 核心 Tools → 搜索增强 → 配置文档）

## 注意事项

- 测试直接调用真实 API，需要有效 cookies.json 且测试中硬编码了 Windows 路径
- `ResolveAppDir()` 通过可执行文件路径（排除 go-build 临时目录）或工作目录推断程序目录，数据目录和 cookie 路径可通过 `--dir` 和 `--cookie` 参数独立指定
- 版本号通过 goreleaser ldflags 注入 `main.version`/`main.commit`/`main.date`

## Active Technologies
- Go 1.24 + `urfave/cli/v3`（CLI 框架）、`mark3labs/mcp-go v0.43.2`（MCP SDK） (001-mcp-server-base)
- 文件系统（Markdown 文件，只读） (001-mcp-server-base)

## Recent Changes
- 001-mcp-server-base: Added Go 1.24 + `urfave/cli/v3`（CLI 框架）、`mark3labs/mcp-go v0.43.2`（MCP SDK）
