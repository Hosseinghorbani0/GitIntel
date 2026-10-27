

<div align="center">


# ⚡ GitIntel

### Deterministic GitHub Engineering Intelligence & Developer Portfolio Engine

**The verifiable evidence layer for software engineers, technical recruiters, and AI coding agents.**

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&logoColor=white)
![Vite](https://img.shields.io/badge/Vite-8-646CFF?logo=vite&logoColor=white)
![MCP](https://img.shields.io/badge/MCP-Compatible-8A2BE2?logo=anthropic&logoColor=white)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**[🌐 Language / زبان](#-language-versions)** — [English](README.md) · [فارسی](docs/translations/README.fa.md) · [Español](docs/translations/README.es.md) · [中文](docs/translations/README.zh-CN.md) · [हिन्दी](docs/translations/README.hi.md) · [العربية](docs/translations/README.ar.md) · [Português](docs/translations/README.pt.md) · [Français](docs/translations/README.fr.md) · [Deutsch](docs/translations/README.de.md) · [日本語](docs/translations/README.ja.md) · [Русский](docs/translations/README.ru.md) · [한국어](docs/translations/README.ko.md)

<p align="center">
  <a href="#-the-problem">The Problem</a> ·
  <a href="#-features">Features</a> ·
  <a href="#-quick-start">Quick Start</a> ·
  <a href="#-mcp--ai-agent-integration">MCP Integration</a> ·
  <a href="#-architecture">Architecture</a> ·
  <a href="#-roadmap">Roadmap</a> ·
  <a href="#-contributing">Contributing</a>
</p>

---

</div>

## 💡 The Problem

Developer assessment lives between two broken extremes:

```text
┌──────────────────────────────────────┐       ┌──────────────────────────────────────┐
│         Vanity Metrics               │       │         Hallucinatory AI             │
│  • Commit streak gaming              │  vs.  │  • "10x rockstar" fluff              │
│  • Star begging & empty forks        │       │  • Opaque black-box scorecards       │
│  • Toy repos inflate language stats  │       │  • Zero verifiable evidence          │
└──────────────────────────────────────┘       └──────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────┐
│                    GitIntel                             │
│  ✅ 100% observable & deterministic signals             │
│  ✅ Byte-weighted language distribution                 │
│  ✅ Verified CI/CD, test trees, release tags            │
│  ✅ Evidence layer for AI agents and profile generators │
└─────────────────────────────────────────────────────────┘
```

| Dimension | Vanity Badges | Generic AI | **GitIntel** |
| :--- | :--- | :--- | :--- |
| **Language Stats** | Repo frequency | Guessed from bio | **Byte-weighted volume** from real codebases |
| **Testing Proof** | Ignored | Hallucinates "TDD expert" | **Git tree–inspected** test suites & CI workflows |
| **Release Cadence** | Untracked | Inferred from commits | **Verified GitHub release tags** |
| **Collaboration** | Fork count as proxy | Buzzwords | **Verified PR/issue split** & contributor counts |
| **AI Usability** | Static image | Garbage-in → garbage-out | **Prompt-ready JSON & native MCP server** |
| **Privacy** | Hosted server | Sends profile to cloud | **100% local-first** on `localhost` |

---

## 🚀 Features

### 1. 🌟 GitHub Profile README Generator
Turn your `username/username` landing page into a verified engineering showcase:
- Auto-generates a Markdown `README.md` grounded in real technical evidence.
- Highlights test practices, CI/CD pipelines, and release history.
- Filters abandoned hobby forks; surfaces your most impactful original work.
- Pair with a scheduled **GitHub Action** to keep your profile perpetually up-to-date.

### 2. 🤖 Native MCP Server (Model Context Protocol)
Let AI coding assistants query your GitHub profile without burning rate-limit quota:
- Exposes `gitintel_analyze_profile`, `gitintel_get_signals`, `gitintel_generate_profile_readme`, and `gitintel_get_ai_prompt` as structured JSON-RPC tools.
- Compresses 100+ repositories into a bounded, high-signal `EvidenceContext` dossier.
- Compatible with **Claude Desktop**, **Cursor**, **Antigravity**, and any MCP-capable agent.

### 3. 📊 Six Verifiable Engineering Signals
Every signal is scored conservatively with supporting evidence and explicit limitation disclosures:

| Signal | What GitIntel Actually Inspects | Honest Limitation |
| :--- | :--- | :--- |
| 🔄 **Code Evolution** | Push & update timestamp recency across repositories | Does not inspect commit volume or diff size |
| 🛡️ **Maintenance** | Active update frequency; archived repos marked "Completed" | Inactivity ≠ poor maintenance |
| 🧪 **Testing Evidence** | Bounded Git tree scan for test dirs & CI workflow files | Does not execute tests or measure line coverage |
| 🤝 **Collaboration** | Verified PR/issue split, contributor counts, incoming forks | Cross-org contributions require authenticated GraphQL |
| 📖 **Documentation** | Tri-state README detection, GitHub Pages, Wiki, LICENSE | Does not evaluate prose quality |
| 📦 **Release & Delivery** | Verified published GitHub releases & semantic version tags | Bounded to high-signal original repositories |

### 4. 💼 Recruiter-Ready Resume & Portfolio Builder
- Export clean, evidence-backed Markdown summaries ready for job applications.
- Curated projects selected by algorithmic signal scoring, with tags like *"Verified CI/CD"* and *"Multi-contributor"*.
- One-click **"Copy as AI Prompt"** to feed token-optimized context to ChatGPT or Claude.

---

## ⚡ Quick Start

### Prerequisites

| Tool | Minimum Version | Check |
| :--- | :--- | :--- |
| Go | 1.26 | `go version` |
| Node.js | 24 | `node --version` |
| npm | 11 | `npm --version` |

### Windows — One-Click Launcher
```bat
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel
run.bat
```
`run.bat` verifies your environment, installs frontend dependencies, checks port availability (`8080` & `5173`), and opens your browser automatically.

### Linux & macOS
```bash
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel

# Terminal 1 — Backend
cd backend
go run ./cmd/server
# Listening on http://localhost:8080

# Terminal 2 — Frontend
cd frontend
npm ci
npm run dev
# Open http://localhost:5173
```

### Optional: GitHub Personal Access Token
GitIntel works without a token using GitHub's public API (60 requests/hr unauthenticated). For richer analysis and higher rate limits, create a classic PAT with `public_repo` read scope and add it to `.env`:

```bash
cp .env.example .env
# Edit .env and set: GITHUB_TOKEN=ghp_your_token_here
```

---

## 🤖 MCP & AI Agent Integration

GitIntel ships a native **Model Context Protocol** server at `backend/cmd/mcp`. Connect it to your AI agent once and query any GitHub profile without writing a single line of API code.

### Connect to Claude Desktop
Add to your `claude_desktop_config.json`:

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

### Connect to Cursor
Add to your Cursor MCP settings using the same `command` and `args` above.

### Available MCP Tools

| Tool | Description |
| :--- | :--- |
| `gitintel_analyze_profile` | Full deterministic analysis: signals, metrics, tech stack, featured projects |
| `gitintel_get_signals` | Six engineering signals with evidence & limitation disclosures |
| `gitintel_generate_profile_readme` | Ready-to-commit GitHub Profile README in Markdown |
| `gitintel_get_ai_prompt` | Pre-engineered, token-optimized prompt context for LLMs |

### Example Agent Query
> *"Analyze @torvalds's GitHub profile with GitIntel and summarize their verified engineering signals in a one-paragraph bio."*

---

## 🏗️ Architecture

```text
┌─────────────────────────────────────────────┐
│          React 19 + Vite (TypeScript)        │
│     Vazirmatn RTL/LTR · Lucide React        │
│           http://localhost:5173              │
└────────────────────┬────────────────────────┘
                     │  /api proxy + AbortController
┌────────────────────▼────────────────────────┐
│           Go 1.26 Backend                   │
│         http://localhost:8080               │
│                                             │
│  ┌───────────────┐  ┌─────────────────────┐ │
│  │  HTTP Handler │  │  Analytics Engine   │ │
│  │  (handler.go) │  │  (engine.go)        │ │
│  └───────┬───────┘  └──────────┬──────────┘ │
│          │                     │            │
│  ┌───────▼───────┐  ┌──────────▼──────────┐ │
│  │  LRU Cache    │  │  GitHub Client       │ │
│  │  (1,000 cap)  │  │  (ETag · Rate Guard) │ │
│  └───────────────┘  └──────────┬──────────┘ │
└───────────────────────────────┼─────────────┘
                                │  HTTPS REST + ETag 304
                   ┌────────────▼────────────┐
                   │    GitHub REST API v3   │
                   │    api.github.com       │
                   └─────────────────────────┘
```

**Engineering hardening highlights:**

- **Zero-allocation cache** — Bounded 1,000-entry LRU serves typed immutable snapshots with no JSON serialization overhead on cache hits.
- **Proactive rate-limit defense** — Reserves a 10-call safety buffer; automatically degrades gracefully when the unauthenticated 60 req/hr budget is low.
- **ETag conditional requests** — `If-None-Match` headers mean HTTP 304 responses consume **zero** rate-limit quota.
- **Race-condition-immune UI** — `AbortController` cancellation ensures rapid successive searches never resolve out of order.
- **Bilingual & accessible** — Full English & Persian (فارسی) support with bidirectional layout isolation (`<bdi>`), Vazirmatn typography, and reduced-motion compliance.
- **Localhost-only CORS** — The backend only accepts requests from `127.0.0.1:5173`; no cross-origin leakage.

---

## 🧪 Testing

```bash
# Backend — unit tests, integration tests, allocation benchmarks
cd backend
go vet ./...
go test ./... -v -count=1

# Frontend — type checking, linting, production build
cd frontend
npm run lint
npm run build
```

The backend test suite currently comprises **109 tests** across 7 packages, including:
- Fixture-based integration tests feeding raw GitHub API JSON through the normalization pipeline.
- Edge-case tests for zero-repo profiles, 100% forked portfolios, Unicode repository names, and 1,000+ repository sets.
- Zero-allocation cache benchmarks, concurrency stress tests, and ETag round-trip tests.

CI runs automatically on every push and pull request to `main` via GitHub Actions.

---

## 🗺️ Roadmap

See the full canonical plan in [ENGINEERING_ROADMAP.md](ENGINEERING_ROADMAP.md).

| Area | Status |
| :--- | :--- |
| Foundational engine & data integrity | ✅ Complete |
| Runtime LRU cache & credential lifecycle | ✅ Complete |
| Frontend reliability & RTL accessibility | ✅ Complete |
| Performance & benchmarks | ✅ Complete |
| Native MCP server (`cmd/mcp`) | ✅ Core complete |
| Profile README generator | ✅ Core complete |
| Fixture-based regression testing | 🔄 67% complete |
| Security hardening & CORS lockdown | 🔄 50% complete |
| Architecture harmonization | 🔄 50% complete |
| Vitest frontend component tests | ⬜ Planned |
| GitHub Action for auto-sync | ⬜ Planned |
| GraphQL external contribution mining | ⬜ Planned |

---

## 🤝 Contributing

Contributions are welcome from developers worldwide. Please read [CONTRIBUTING.md](CONTRIBUTING.md) for setup instructions, coding standards, and the contribution workflow.

**High-impact open areas:**
- **Frontend tests** (`GI-TEST-003`) — Set up Vitest + React Testing Library.
- **GitHub Action** (`GI-AI-003`) — Weekly cron to rebuild and commit profile READMEs.
- **GraphQL mining** (`GI-PROD-001`) — Detect contributions to external organizations.
- **Profile README themes** — Bento Grid, Terminal Hacker, Minimalist formats.

To get started:
1. Fork the repository and create a branch from `main`.
2. Pick a task from [ENGINEERING_ROADMAP.md](ENGINEERING_ROADMAP.md) or open an issue.
3. Run backend and frontend validation commands before submitting a PR.

---

## 🔒 Security & Privacy

- GitIntel runs **entirely on localhost**. Your GitHub Personal Access Token and analysis data never leave your machine.
- CORS is restricted to `127.0.0.1:5173` only — no external origin can reach the backend.
- Prompt injection defenses are built into the LLM integration: untrusted repository data is always treated as data, never as instructions.
- To report a security vulnerability, please open a private GitHub Security Advisory on this repository.

---

## 📄 License & Acknowledgements

Distributed under the **GNU Affero General Public License v3.0 (GNU AGPL-3.0)** — see the [LICENSE](LICENSE) file for details.

Built with:
- [Go](https://go.dev/) · [google/go-github](https://github.com/google/go-github) · [godotenv](https://github.com/joho/godotenv)
- [React 19](https://react.dev/) · [Vite 8](https://vitejs.dev/) · [TypeScript](https://www.typescriptlang.org/) · [Lucide React](https://lucide.dev/)
- [Vazirmatn](https://github.com/rastikerdar/vazirmatn) — Persian typography

---

## 🌐 Language Versions

| Language | File |
| :--- | :--- |
| 🇺🇸 English (canonical) | [README.md](README.md) |
| 🇮🇷 فارسی (Persian) | [docs/translations/README.fa.md](docs/translations/README.fa.md) |
| 🇪🇸 Español (Spanish) | [docs/translations/README.es.md](docs/translations/README.es.md) |
| 🇨🇳 简体中文 (Chinese Simplified) | [docs/translations/README.zh-CN.md](docs/translations/README.zh-CN.md) |
| 🇮🇳 हिन्दी (Hindi) | [docs/translations/README.hi.md](docs/translations/README.hi.md) |
| 🇸🇦 العربية (Arabic) | [docs/translations/README.ar.md](docs/translations/README.ar.md) |
| 🇧🇷 Português (Portuguese) | [docs/translations/README.pt.md](docs/translations/README.pt.md) |
| 🇫🇷 Français (French) | [docs/translations/README.fr.md](docs/translations/README.fr.md) |
| 🇩🇪 Deutsch (German) | [docs/translations/README.de.md](docs/translations/README.de.md) |
| 🇯🇵 日本語 (Japanese) | [docs/translations/README.ja.md](docs/translations/README.ja.md) |
| 🇷🇺 Русский (Russian) | [docs/translations/README.ru.md](docs/translations/README.ru.md) |
| 🇰🇷 한국어 (Korean) | [docs/translations/README.ko.md](docs/translations/README.ko.md) |

<div align="center">

---

If GitIntel helps you build a more honest developer profile, consider giving it a ⭐ on GitHub.

</div>
