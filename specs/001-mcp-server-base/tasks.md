# Tasks: MCP Server 基础框架

**Input**: Design documents from `/specs/001-mcp-server-base/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 添加 mcp-go 依赖，创建包目录结构

- [x] T001 Add `github.com/mark3labs/mcp-go` dependency via `go get github.com/mark3labs/mcp-go@v0.43.2`
- [x] T002 Create `mcp/` package directory with empty files: `mcp/server.go`, `mcp/tools.go`, `mcp/search.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 实现 storage/reader.go 数据读取层和数据模型，所有 MCP Tool 都依赖此层

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 Define data model structs (`PostInfo`, `AuthorPosts`, `SearchResult`, `SearchResponse`) in `storage/reader.go`
- [x] T004 Implement `ListAuthors(dataDir string) ([]string, error)` in `storage/reader.go` — scan data directory for author subdirectories, filter out `.assets/` and non-directory entries
- [x] T005 Implement `ParsePostInfo(fileName, category, authorDir string) PostInfo` helper in `storage/reader.go` — extract title and publish time from filename format `{YYYY-MM-DD_HH_MM_SS}_{SafeTitle}.md`
- [x] T006 Implement `ListPosts(dataDir, author string) (*AuthorPosts, error)` in `storage/reader.go` — scan motions/ and album subdirectories, return grouped post lists with relative paths
- [x] T007 Implement `ReadPost(dataDir, relativePath string) (string, error)` in `storage/reader.go` — read a single markdown file by relative path, return full content
- [x] T008 Implement `FindPostByTitle(dataDir, author, titleKeyword string) ([]PostInfo, error)` in `storage/reader.go` — case-insensitive title substring match across all posts of an author

**Checkpoint**: storage/reader.go complete — all file I/O operations available for Tool handlers

---

## Phase 3: User Story 1 - 连接 MCP Server (Priority: P1) 🎯 MVP

**Goal**: 程序能以 MCP Server 模式启动，通过 stdio 被 Claude Code 连接，返回 Tool 列表

**Independent Test**: 构建程序后在 Claude Code MCP 配置中添加 Server，确认连接成功且 Tool 列表可见

### Implementation for User Story 1

- [x] T009 [US1] Modify `Before` hook in `main.go` to conditionally skip cookie loading when the subcommand is `mcp` — only initialize logger and Config (dataDir), skip `afdian.GetCookies()` call
- [x] T010 [US1] Implement `NewServer(cfg *config.Config, version string) *server.MCPServer` in `mcp/server.go` — create MCP server with `server.NewMCPServer()`, register all 4 tools with their input schemas per contracts/mcp-tools.md, return server instance
- [x] T011 [US1] Implement `Serve(s *server.MCPServer) error` in `mcp/server.go` — call `server.ServeStdio(s)` to start stdio transport
- [x] T012 [US1] Add `mcp` subcommand in `main.go` — register alongside existing `motions`/`albums`/`update` commands, action calls `mcp.NewServer(cfg, version)` then `mcp.Serve(s)`, accepts `--dir` flag (via global Before hook)
- [ ] T013 [US1] Build and verify connection — run `go build -o AfdianToMarkdown.exe .`, add MCP config to Claude Code, confirm server connects and tool list is visible

**Checkpoint**: MCP Server 可启动、可连接，Claude Code 能看到 4 个 Tool（handler 返回 placeholder 响应即可）

---

## Phase 4: User Story 2 - 浏览已下载的作者和文章列表 (Priority: P1)

**Goal**: `list_authors` 和 `list_posts` Tool 返回真实数据

**Independent Test**: 在 Claude Code 中调用 `list_authors` 看到作者列表，调用 `list_posts` 看到分组文章列表

### Implementation for User Story 2

- [x] T014 [P] [US2] Implement `handleListAuthors` handler in `mcp/tools.go` — call `storage.ListAuthors(cfg.DataDir)`, format output as per contract (count + bulleted list), handle empty/missing directory edge cases
- [x] T015 [P] [US2] Implement `handleListPosts` handler in `mcp/tools.go` — extract `author` param via `request.RequireString("author")`, call `storage.ListPosts()`, format output with motions section and per-album sections showing `[date] title → relative/path`, handle author-not-found error
- [ ] T016 [US2] Verify in Claude Code — ask Claude to list authors and list posts for a specific author, confirm output matches contract format

**Checkpoint**: 浏览功能完整，可在 Claude Code 中查看作者和文章列表

---

## Phase 5: User Story 3 - 阅读指定文章 (Priority: P1)

**Goal**: `read_post` Tool 能通过路径或标题关键词读取文章内容

**Independent Test**: 在 Claude Code 中通过路径读取文章，通过标题关键词读取文章，确认内容完整

### Implementation for User Story 3

- [x] T017 [US3] Implement `handleReadPost` handler in `mcp/tools.go` — support two modes: (1) if `path` param provided, call `storage.ReadPost(cfg.DataDir, path)` directly; (2) if `author`+`title` provided, call `storage.FindPostByTitle()` — if single match read and return content, if multiple matches return list for user to choose, if no match return error. Prepend file path header to output per contract
- [ ] T018 [US3] Verify in Claude Code — read a post by path, read by author+title keyword, test multi-match scenario, test not-found scenario

**Checkpoint**: 文章阅读功能完整，可通过路径或标题检索并阅读全文

---

## Phase 6: User Story 4 - 全文关键词搜索 (Priority: P2)

**Goal**: `search` Tool 能在所有文档中进行关键词搜索并返回匹配片段

**Independent Test**: 在 Claude Code 中搜索一个关键词，确认返回匹配片段、上下文行和文件路径

### Implementation for User Story 4

- [x] T019 [US4] Implement `Search(dataDir, query, author string, maxResults int) (*SearchResponse, error)` in `mcp/search.go` — walk markdown files (optionally filtered by author), read each file line by line, case-insensitive plain text match via `strings.Contains(strings.ToLower(...))`, collect matching lines with 3 lines context before/after, cap at maxResults, track total count for truncation indicator
- [x] T020 [US4] Implement `formatSearchResponse(resp *SearchResponse) string` helper in `mcp/search.go` — format output per contract: header with count, each result block with file path + line number + context lines (using `>` prefix for match line), truncation notice if applicable
- [x] T021 [US4] Implement `handleSearch` handler in `mcp/tools.go` — extract `query` (required) and `author` (optional) params, validate query not empty, call `Search(cfg.DataDir, query, author, 20)`, format and return result
- [ ] T022 [US4] Verify in Claude Code — search for a known keyword, test with author filter, test empty result, test result truncation with common keyword

**Checkpoint**: 搜索功能完整，可在 Claude Code 中搜索关键词并查看匹配片段

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: 最终验证和配置文档

- [ ] T023 [P] Add Claude Code MCP configuration example to quickstart.md at `specs/001-mcp-server-base/quickstart.md` with actual built binary path
- [x] T024 Ensure `go build` succeeds with no warnings, verify all edge cases from spec (empty dir, missing author, special chars in filenames, large files, regex special chars in search query)
- [ ] T025 Run full end-to-end validation per quickstart.md — build, configure Claude Code, test all 4 tools in conversation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (`go get` must complete first)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (needs data model structs) — but Tool handlers can initially return placeholders
- **User Story 2 (Phase 4)**: Depends on Phase 2 (storage functions) + Phase 3 (server running)
- **User Story 3 (Phase 5)**: Depends on Phase 2 (storage functions) + Phase 3 (server running)
- **User Story 4 (Phase 6)**: Depends on Phase 3 (server running), search.go is independent of storage/reader.go
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (连接)**: Foundation only — no dependency on other stories
- **US2 (浏览)**: Depends on US1 (server must be running) — independent of US3/US4
- **US3 (阅读)**: Depends on US1 (server must be running) — independent of US2/US4
- **US4 (搜索)**: Depends on US1 (server must be running) — independent of US2/US3

### Within Each User Story

- Storage layer functions before Tool handlers
- Tool handlers before verification
- Verify each story independently before moving to next

### Parallel Opportunities

- T014 and T015 (US2 handlers) can run in parallel — different functions in same file but no dependencies
- US2, US3, US4 implementation can start in parallel after US1 is complete
- T019 and T020 (US4 search logic) are sequential (T020 formats T019's output)

---

## Parallel Example: User Story 2

```bash
# These two handlers can be implemented in parallel (different functions, no dependencies):
Task T014: "Implement handleListAuthors handler in mcp/tools.go"
Task T015: "Implement handleListPosts handler in mcp/tools.go"
```

## Parallel Example: After Phase 3

```bash
# Once US1 (server connection) is verified, these can start in parallel:
Task T014: "US2 - handleListAuthors in mcp/tools.go"
Task T017: "US3 - handleReadPost in mcp/tools.go"
Task T019: "US4 - Search function in mcp/search.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T008)
3. Complete Phase 3: User Story 1 (T009-T013)
4. **STOP and VALIDATE**: Confirm MCP Server connects to Claude Code
5. Proceed to remaining stories

### Incremental Delivery

1. Setup + Foundational → Storage layer ready
2. US1 → Server connects → **MVP confirmed**
3. US2 → Browse authors/posts → Usable for navigation
4. US3 → Read articles → Core RAG value delivered
5. US4 → Search → Full feature set complete
6. Polish → Edge cases, docs → Production ready

---

## Notes

- [P] tasks = different files or independent functions, no dependencies
- [Story] label maps task to specific user story for traceability
- Total: 25 tasks (2 setup + 6 foundational + 5 US1 + 3 US2 + 2 US3 + 4 US4 + 3 polish)
- No test tasks generated (not requested in spec)
- `cfg` (config.Config) is captured in closures when registering tool handlers — passed from main.go to mcp.NewServer()
