# Feature Specification: MCP Server 基础框架

**Feature Branch**: `001-mcp-server-base`
**Created**: 2026-02-08
**Status**: Draft
**Input**: User description: "在现有 CLI 程序中添加 mcp 子命令，以 stdio 方式启动 MCP Server，实现基础框架搭建与核心文档检索 Tools"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 连接 MCP Server (Priority: P1)

用户在 Claude Code 的 MCP 配置中添加 AfdianToMarkdown 作为 MCP Server。启动 Claude Code 后，Server 通过 stdio 传输正常连接，Claude Code 能发现并列出该 Server 提供的所有 Tools。

**Why this priority**: 这是所有后续功能的基础。如果 Server 无法启动和连接，其他一切都无从谈起。

**Independent Test**: 在 Claude Code MCP 配置中添加 Server 配置后，启动 Claude Code，确认 Server 连接成功且 Tool 列表可见。

**Acceptance Scenarios**:

1. **Given** 用户已配置 MCP Server（指定 `mcp` 子命令和 `--dir` 参数），**When** Claude Code 启动并初始化 MCP 连接，**Then** Server 成功启动，Claude Code 显示该 Server 为已连接状态
2. **Given** MCP Server 正在运行，**When** Claude Code 请求 Tool 列表，**Then** 返回所有已注册的 Tool 及其参数描述
3. **Given** 用户未指定 `--dir` 参数，**When** MCP Server 启动，**Then** 使用程序所在目录下的 `data` 文件夹作为默认数据目录

---

### User Story 2 - 浏览已下载的作者和文章列表 (Priority: P1)

用户通过 Claude Code 对话，要求查看已下载的作者列表，或查看某位作者下所有文章的标题列表。Claude 调用相应的 MCP Tool 返回结构化信息。

**Why this priority**: 浏览是检索的入口，用户需要先知道有什么内容才能进一步查询。

**Independent Test**: 在有已下载文档的数据目录下启动 Server，调用 `list_authors` 确认返回作者列表，再调用 `list_posts` 确认返回该作者下的文章列表。

**Acceptance Scenarios**:

1. **Given** 数据目录下存在多个作者文件夹，**When** 调用 `list_authors`，**Then** 返回所有作者名称的列表
2. **Given** 某作者文件夹下有 motions 和 albums 子目录，**When** 调用 `list_posts` 并指定该作者，**Then** 返回按类别（动态/作品集）分组的文章标题列表，包含文件路径
3. **Given** 指定的作者不存在，**When** 调用 `list_posts`，**Then** 返回明确的错误信息说明作者未找到
4. **Given** 某作者目录下没有任何 markdown 文件，**When** 调用 `list_posts`，**Then** 返回空列表

---

### User Story 3 - 阅读指定文章 (Priority: P1)

用户在 Claude Code 中要求阅读某篇已下载的文章。Claude 调用 MCP Tool 获取文章的完整 Markdown 内容，然后基于该内容回答用户的问题。

**Why this priority**: 文章阅读是核心价值。用户下载这些文档就是为了在 Claude 中检索和阅读。

**Independent Test**: 调用 `read_post` 指定作者和标题（或路径），确认返回完整的 Markdown 内容。

**Acceptance Scenarios**:

1. **Given** 文章文件存在，**When** 调用 `read_post` 并指定文件路径，**Then** 返回该文件的完整 Markdown 内容
2. **Given** 文章文件存在，**When** 调用 `read_post` 并指定作者名和标题关键词，**Then** 匹配并返回对应文章内容
3. **Given** 指定的文件不存在，**When** 调用 `read_post`，**Then** 返回明确的错误信息
4. **Given** 标题关键词匹配到多篇文章，**When** 调用 `read_post`，**Then** 返回匹配列表供用户选择，而非随机返回一篇

---

### User Story 4 - 全文关键词搜索 (Priority: P2)

用户在 Claude Code 中搜索某个关键词或主题。Claude 调用 MCP Tool 在所有已下载文档中进行全文搜索，返回匹配的文章片段及上下文，帮助用户快速定位相关内容。

**Why this priority**: 搜索是 800+ 篇文档场景下的关键能力，但优先级略低于浏览和阅读，因为用户可以先通过浏览找到目标。

**Independent Test**: 调用 `search` 指定一个关键词，确认返回包含该关键词的文章片段、上下文行和文件路径。

**Acceptance Scenarios**:

1. **Given** 多篇文章包含关键词"哲学"，**When** 调用 `search` 查询"哲学"，**Then** 返回所有匹配文章的片段，每个片段包含匹配行及其前后各 3 行上下文
2. **Given** 搜索结果超过 20 条，**When** 调用 `search`，**Then** 只返回前 20 条最相关的匹配，并提示还有更多结果
3. **Given** 指定了可选的 `author` 参数，**When** 调用 `search`，**Then** 仅在该作者的文档中搜索
4. **Given** 关键词在所有文档中都不存在，**When** 调用 `search`，**Then** 返回空结果并说明未找到匹配

---

### Edge Cases

- 数据目录为空或不存在时，`list_authors` 应返回空列表而非报错
- 文件名包含特殊字符（中文、空格、标点）时，路径处理应正常工作
- 非常大的 Markdown 文件（超过 100KB）被 `read_post` 读取时，应完整返回内容
- MCP Server 启动时 `--dir` 指向的目录不存在，应返回明确的启动错误
- `search` 关键词包含正则特殊字符时，应作为纯文本匹配而非正则表达式

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 程序 MUST 提供 `mcp` 子命令，与现有的 `motions`、`albums` 等子命令同级
- **FR-002**: `mcp` 子命令 MUST 以 stdio 传输模式启动 MCP Server
- **FR-003**: `mcp` 子命令 MUST 不依赖 cookies.json（不加载 Cookie），仅需 `--dir` 参数指定数据目录
- **FR-004**: MCP Server MUST 注册 `list_authors` Tool，返回数据目录下所有已下载的作者列表
- **FR-005**: MCP Server MUST 注册 `list_posts` Tool，接受 `author` 参数，返回该作者下按类别分组的文章列表
- **FR-006**: MCP Server MUST 注册 `read_post` Tool，支持通过文件路径或作者+标题关键词读取文章内容
- **FR-007**: MCP Server MUST 注册 `search` Tool，接受 `query` 必选参数和 `author` 可选参数，执行全文关键词搜索
- **FR-008**: `search` Tool MUST 返回匹配行的前后各 3 行作为上下文，且结果上限为 20 条
- **FR-009**: 所有 Tool 返回结果 MUST 包含文件的相对路径信息
- **FR-010**: `list_posts` 返回结果 MUST 区分动态（motions）和作品集（albums）两种类别

### Key Entities

- **Author（作者）**: 由数据目录下的子文件夹名标识，包含 motions 和/或 albums 子目录
- **Post（文章）**: 一个 Markdown 文件，属于某位作者的 motions 或某个 album，有标题和发布时间（编码在文件名中）
- **Album（作品集）**: 作者目录下除 motions 外的子文件夹，包含一组相关文章

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: MCP Server 能被 Claude Code 成功连接并正常交互，连接建立后 Tool 列表在 2 秒内返回
- **SC-002**: `list_authors` 和 `list_posts` 在 800+ 篇文档的数据目录下即时响应（用户无感知延迟）
- **SC-003**: `read_post` 能正确返回数据目录下任意 Markdown 文件的完整内容
- **SC-004**: `search` 在 800+ 篇文档中完成关键词搜索并返回结果，用户等待时间不超过 3 秒
- **SC-005**: 用户能通过 Claude Code 对话自然地浏览、阅读和搜索已下载的爱发电文档，无需离开对话界面

## Assumptions

- 数据目录结构遵循 AfdianToMarkdown 的标准格式：`{作者}/{motions|albums}/{文件}.md`
- 文件名中已编码了发布时间（如 `2024-03-15_标题.md`），可从文件名中提取
- 搜索功能在第一阶段使用简单的逐文件关键词匹配，无需倒排索引或向量搜索
- MCP Server 运行期间数据目录内容不会频繁变化，无需实时监听文件变更
- 使用 `github.com/mark3labs/mcp-go` 作为 MCP SDK

## Scope Boundaries

**包含在本阶段**:
- `mcp` 子命令和 stdio 传输
- 4 个核心 Tool：`list_authors`、`list_posts`、`read_post`、`search`
- 基本的关键词搜索（单关键词，纯文本匹配）

**不包含在本阶段**:
- 多关键词搜索（AND/OR 逻辑）
- 搜索结果按相关度排序
- 倒排索引或向量搜索
- 文档元信息提取（发布时间、评论等结构化数据）
- MCP Resource 或 Prompt 类型的注册
- 下载功能的 MCP 化（保持仅 CLI 使用）
