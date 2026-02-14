# Feature Specification: MCP HTTP Transport

**Feature Branch**: `002-mcp-http-transport`
**Created**: 2026-02-10
**Status**: Draft
**Input**: User description: "为MCP服务器实现HTTP调用功能，使其可以在NAS上通过cron任务自动更新并以HTTP方式提供MCP服务。使用HTTP Streamable传输（非SSE），运行在Tailscale内网中无需身份认证。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 以HTTP模式启动MCP服务 (Priority: P1)

用户在NAS上运行命令启动MCP服务器，服务器监听指定端口，通过HTTP Streamable协议为远程MCP客户端（如Claude Desktop、其他AI工具）提供文档查询服务。

**Why this priority**: 这是核心功能，没有HTTP传输能力就无法实现远程访问，其他所有场景都依赖此功能。

**Independent Test**: 启动HTTP服务后，使用任意MCP客户端通过HTTP连接到服务器，调用`list_authors`工具并获得正确响应。

**Acceptance Scenarios**:

1. **Given** 用户已有数据目录和已下载的文档, **When** 用户运行`./AfdianToMarkdown mcp --http`命令, **Then** 服务器启动并监听默认端口，日志输出监听地址
2. **Given** HTTP MCP服务器已启动, **When** MCP客户端通过HTTP连接并调用`list_authors`工具, **Then** 服务器返回正确的作者列表
3. **Given** HTTP MCP服务器已启动, **When** 用户指定自定义端口`--port 9090`, **Then** 服务器在9090端口上监听

---

### User Story 2 - 优雅关闭HTTP服务 (Priority: P2)

用户需要停止HTTP MCP服务（手动或通过进程管理工具），服务器在完成当前请求后优雅退出，不会丢失正在处理的响应。

**Why this priority**: 优雅关闭对于服务可靠性至关重要，特别是在NAS上由进程管理器管理生命周期的场景。

**Independent Test**: 启动HTTP服务，发送请求的同时发送SIGTERM信号，验证正在处理的请求完成后服务退出。

**Acceptance Scenarios**:

1. **Given** MCP HTTP服务正在运行且无活跃请求, **When** 服务器收到SIGTERM信号, **Then** 服务器立即关闭
2. **Given** MCP HTTP服务正在处理请求, **When** 服务器收到SIGTERM/SIGINT信号, **Then** 服务器等待当前请求完成后关闭

---

### User Story 3 - 使用现有stdio模式（向后兼容） (Priority: P3)

已有的stdio传输模式继续正常工作，用户可以选择stdio或HTTP模式运行MCP服务器。

**Why this priority**: 向后兼容确保现有用户不受影响，但功能已存在，只需确保不被破坏。

**Independent Test**: 在不添加`--http`参数的情况下运行`mcp`子命令，验证stdio模式仍然正常工作。

**Acceptance Scenarios**:

1. **Given** 用户运行`./AfdianToMarkdown mcp`（无额外参数）, **When** 启动完成, **Then** 服务器以stdio模式运行，行为与之前一致

---

### Edge Cases

- 当指定端口已被占用时，服务器应输出明确的错误信息并退出
- 当数据目录为空或不存在时，服务器仍应正常启动（工具返回空结果）
- 当服务器收到多个并发请求时，应能正确处理不会崩溃
- 当服务器在处理请求过程中收到终止信号时，应等待当前请求完成后再关闭

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 支持通过`--http`标志以HTTP Streamable传输模式启动MCP服务器
- **FR-002**: 系统 MUST 支持通过`--port`参数指定HTTP监听端口，默认值为8080
- **FR-003**: 系统 MUST 支持通过`--host`参数指定HTTP监听地址，默认值为`0.0.0.0`（监听所有接口）
- **FR-004**: HTTP模式下所有已有MCP工具（list_authors、list_posts、read_post、search）MUST 正常工作，返回与stdio模式一致的结果
- **FR-005**: 系统 MUST 不包含任何身份认证机制（服务运行在Tailscale内网中）
- **FR-006**: 系统 MUST 在收到SIGTERM或SIGINT信号时优雅关闭HTTP服务器
- **FR-007**: 系统 MUST 在启动时通过日志输出实际监听地址和端口
- **FR-008**: 系统 MUST 在`--http`标志未设置时保持现有的stdio传输模式不变

### Key Entities

- **MCP Server**: 核心服务实例，注册工具并处理请求。可通过stdio或HTTP传输层对外提供服务。
- **HTTP Transport**: 网络传输层，将MCP协议映射到HTTP Streamable端点，处理连接管理和请求路由。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: MCP客户端能通过HTTP连接到服务器并成功调用所有4个工具（list_authors、list_posts、read_post、search），每次调用在5秒内返回结果
- **SC-002**: 服务器在收到终止信号后10秒内完成优雅关闭
- **SC-003**: 服务器能同时处理多个并发MCP请求而不丢失或错误响应

## Assumptions

- 服务运行在Tailscale VPN网络内，不需要TLS加密或身份认证
- NAS环境支持Go编译的二进制文件和cron调度
- mcp-go库已支持HTTP Streamable传输（非SSE）
- 用户会通过外部进程管理（如systemd、supervisord或简单的shell脚本）来管理服务的生命周期
- cron任务和部署脚本由用户自行编写，项目提供使用示例即可
