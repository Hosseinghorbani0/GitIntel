<div align="center">

# ⚡ GitIntel

### 确定性 GitHub 工程智能与开发者作品集引擎

**面向软件工程师、技术招聘人员和 AI 编程智能体的可验证证据层。**

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](../../LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](../../CONTRIBUTING.md)

**[🌐 语言版本](#-语言版本)** — [English](../../README.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [中文](README.zh-CN.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [日本語](README.ja.md) · [Русский](README.ru.md) · [한국어](README.ko.md)

</div>

---

## 💡 痛点与背景

当前对开发者的评估往往陷入两个极端：

- **虚荣指标** — 被操纵的 commit 连续打卡记录、空 fork、人为堆砌语言统计数据的玩具仓库。
- **大模型幻觉** — 黑盒评分卡、毫无事实依据的空泛溢美之词。

**GitIntel** 提供了第三条道路：基于真实 GitHub 工件生成的 100% 可观测、确定性工程信号。

---

## 🚀 核心功能

### 1. 🌟 GitHub Profile README 生成器
将你的 `username/username` 个人主页转变为真实技术实力的展示窗口：
- 依据实际工程数据自动化生成 `README.md`。
- 高亮测试实践、CI/CD 流水线与版本发布历史。
- 智能过滤废弃 fork，优先呈现你最具技术价值的原创项目。

### 2. 🤖 原生 MCP 服务器 (Model Context Protocol)
让 AI 助手能够深入查询 GitHub 档案且不浪费 API 配额：
- JSON-RPC 工具：`gitintel_analyze_profile`、`gitintel_get_signals`、`gitintel_generate_profile_readme`、`gitintel_get_ai_prompt`。
- 全面兼容 **Claude Desktop**、**Cursor**、**Antigravity** 及所有主流 MCP 客户端。

### 3. 📊 六大可验证工程信号

| 信号 | GitIntel 实际检测内容 | 诚实限制说明 |
| :--- | :--- | :--- |
| 🔄 **代码演进 (Code Evolution)** | Push 与 Update 时间戳的新鲜度 | 不统计具体的 commit 数量 |
| 🛡️ **工程维护 (Maintenance)** | 更新频率；已归档仓库标记为“已结项” | 不活跃不等于维护不良 |
| 🧪 **测试证据 (Testing Evidence)** | Git 树扫描测试目录与 CI 工作流 | 不运行测试也不统计覆盖率 |
| 🤝 **协同实践 (Collaboration)** | 验证 PR/Issue 比例、贡献者数量、入向 fork | 外部贡献洞察需 GraphQL 支持 |
| 📖 **工程文档 (Documentation)** | README 三态检测、GitHub Pages、Wiki、LICENSE | 不评估文档行文优劣 |
| 📦 **发布交付 (Release & Delivery)** | 验证 GitHub Releases 与语义化版本标签 | 仅限于活跃的原创项目 |

### 4. 💼 招聘就绪的简历与作品集
- 导出附带可信证据的 Markdown 摘要，直接用于求职与技术面试。
- 一键 **“复制为 AI Prompt”**，为大模型上下文提供高度压缩、省 Token 的结构化背景信息。

---

## ⚡ 快速开始

### 环境依赖

| 工具 | 最低版本要求 |
| :--- | :--- |
| Go | 1.26 |
| Node.js | 24 |
| npm | 11 |

### Windows — 一键启动
```bat
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel
run.bat
```

### Linux 与 macOS
```bash
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel

# 终端 1 — 后端服务
cd backend
go run ./cmd/server

# 终端 2 — 前端服务
cd frontend
npm ci
npm run dev
```

### GitHub 访问令牌 (可选)
```bash
cp .env.example .env
# 编辑 .env 文件并填入：GITHUB_TOKEN=ghp_your_token_here
```

---

## 🤖 MCP 集成配置

### 连接到 Claude Desktop
将以下内容加入你的 `claude_desktop_config.json`：

```json
{
  "mcpServers": {
    "gitintel": {
      "command": "go",
      "args": ["run", "./backend/cmd/mcp"],
      "cwd": "/path/to/GitIntel"
    }
  }
}
```

---

## 🏗️ 系统架构

```text
┌─────────────────────────────────────────┐
│      React 19 + Vite (TypeScript)       │
│         http://localhost:5173           │
└──────────────────┬──────────────────────┘
                   │ proxy /api
┌──────────────────▼──────────────────────┐
│          Go 1.26 后端服务                │
│        http://localhost:8080            │
│  LRU 缓存 (1,000 条) +                   │
│  GitHub 客户端 (ETag · Rate Guard)       │
└──────────────────┬──────────────────────┘
                   │ REST + ETag 304
        ┌──────────▼──────────┐
        │   GitHub REST API   │
        └─────────────────────┘
```

**工程亮点：**
- **零分配内存缓存** — 1,000 槽位 LRU 缓存，带不可变数据快照。
- **主动速率限制防御** — 预留 10 次请求安全缓冲，防止配额耗尽。
- **ETag 条件请求** — HTTP 304 响应消耗 **0** 速率限制配额。
- **竞态免疫的前端状态** — 使用 `AbortController` 自动取消过时请求。
- **双语与无障碍设计** — 完整支持波斯语 (RTL) 与英语无缝切换。
- **本地专属安全隔离** — 后端 CORS 仅允许 `127.0.0.1:5173`。

---

## 🧪 自动化测试

```bash
# 后端测试
cd backend
go vet ./...
go test ./... -v -count=1

# 前端校验
cd frontend
npm run lint
npm run build
```

后端测试套件覆盖 7 个模块，包含 **109 项单元与集成测试**。

---

## 🗺️ 工程路线图

完整计划请参阅 [ENGINEERING_ROADMAP.md](../../ENGINEERING_ROADMAP.md)。

| 领域 | 状态 |
| :--- | :--- |
| 核心计算引擎与数据完整性 | ✅ 已完成 |
| LRU 缓存与凭证管理声明周期 | ✅ 已完成 |
| 前端健壮性与 RTL 无障碍支持 | ✅ 已完成 |
| 性能基准与高负载优化 | ✅ 已完成 |
| 原生 MCP 服务器 | ✅ 核心已完成 |
| Profile README 自动化生成 | ✅ 核心已完成 |
| 基于真实数据的回归测试套件 | 🔄 进行中 (67%) |
| 安全加固与审计 | 🔄 进行中 (50%) |
| Vitest 前端单元测试 | ⬜ 规划中 |
| GitHub Action 自动同步工作流 | ⬜ 规划中 |

---

## 🤝 参与贡献

欢迎所有社区贡献！请先阅读 [CONTRIBUTING.md](../../CONTRIBUTING.md) 了解环境搭建与贡献规范。

---

## 🔒 安全与隐私

- GitIntel 完全运行于**本地 (localhost)**。你的 Token 与代码数据绝不上传第三方服务器。
- CORS 访问严格限制为本地回环 `127.0.0.1:5173`。
- LLM 集成模块内置 Prompt 注入防御机制与源码剥离隔离保护。

---

## 📄 开源许可证

本项目基于 **GNU Affero General Public License v3.0 (GNU AGPL-3.0)** 开源 — 详情请参阅 [LICENSE](../../LICENSE) 文件。

---

## 🌐 语言版本

| 语言 | 对应文档 |
| :--- | :--- |
| 🇺🇸 English (权威源版本) | [README.md](../../README.md) |
| 🇮🇷 فارسی | [README.fa.md](README.fa.md) |
| 🇪🇸 Español | [README.es.md](README.es.md) |
| 🇨🇳 简体中文 | [README.zh-CN.md](README.zh-CN.md) |
| 🇮🇳 हिन्दी | [README.hi.md](README.hi.md) |
| 🇸🇦 العربية | [README.ar.md](README.ar.md) |
| 🇧🇷 Português | [README.pt.md](README.pt.md) |
| 🇫🇷 Français | [README.fr.md](README.fr.md) |
| 🇩🇪 Deutsch | [README.de.md](README.de.md) |
| 🇯🇵 日本語 | [README.ja.md](README.ja.md) |
| 🇷🇺 Русский | [README.ru.md](README.ru.md) |
| 🇰🇷 한국어 | [README.ko.md](README.ko.md) |
