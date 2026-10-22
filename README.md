<div align="center">

# ⚡ GitIntel

### Deterministic GitHub Engineering Intelligence & Developer Portfolio Engine

**The verifiable evidence layer for modern software engineers, technical recruiters, and AI coding agents.**

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![React Version](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&logoColor=white)
![Vite](https://img.shields.io/badge/Vite-8-646CFF?logo=vite&logoColor=white)
![MCP Ready](https://img.shields.io/badge/MCP-Compatible-8A2BE2?logo=anthropic&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green.svg)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

<p align="center">
  <a href="#-why-gitintel">Why GitIntel?</a> •
  <a href="#-flagship-capabilities">Flagship Features</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-ai-agent--mcp-integration">MCP & AI Integration</a> •
  <a href="#-engineering-signals-matrix">Signals Matrix</a> •
  <a href="#-architecture">Architecture</a> •
  <a href="#-roadmap--community">Roadmap</a>
</p>

---

</div>

## 💡 The Problem GitIntel Solves

Developer assessment has been broken across two extremes:

```text
┌──────────────────────────────────────┐       ┌──────────────────────────────────────┐
│       1. Vanity Metric Era           │       │          2. Hallucinatory AI         │
│  - Green commit streak gaming        │  vs.  │  - Flowery "10x rockstar" fluff      │
│  - Star begging & empty forks        │       │  - Opaque black-box score cards      │
│  - Toy repos outweighing real code   │       │  - Zero verifiable evidence backing  │
└──────────────────────────────────────┘       └──────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          3. The GitIntel Era                                │
│       ✅ 100% Observable & Deterministic Artifacts                          │
│       ✅ Byte-Weighted Language Distribution (Real Volume)                  │
│       ✅ Verified CI/CD, Test Trees, Release Tags & Issue/PR Splits         │
│       ✅ The Missing Evidence Layer for AI Models & Profile Generators      │
└─────────────────────────────────────────────────────────────────────────────┘
```

| Dimension | Vanity Badges | Generic AI Prompts | **GitIntel Engine** |
| :--- | :--- | :--- | :--- |
| **Language Stats** | Repo frequency (5 toy scripts = 5 repos) | Guessed from bio / self-claim | **Byte-weighted volume** sampled from real codebases |
| **Testing Proof** | Ignored or based on repo names (`test-app`) | Hallucinates "expert in TDD" | **Tree-inspected test suites** (`tests/`, `*_test.go`, CI workflows) |
| **Release Cadence** | Untracked | Inferred from commit dates | **Verified GitHub release tags** & semantic delivery intervals |
| **Collaboration** | Equates fork clicks with collaboration | Buzzwords ("Strong team player") | **Verified open issue / pull-request splits** & contributor counts |
| **AI Usability** | None (static PNG/SVG) | Garbage-in, garbage-out | **Prompt-Ready Evidence JSON & Native MCP Server** |
| **Privacy & Security** | Hosted server tracks IP | Sends full profile to cloud LLM | **100% Local-First** (`localhost`), zero secret retention |

---

## 🚀 Flagship Capabilities

### 1. 🌟 Dynamic GitHub Profile README Generator (`username/username`)
Turn your GitHub landing page from a generic template into an **engineering masterpiece**:
* Auto-generate a beautiful, evidence-grounded `README.md` showcasing your real technical volume.
* Highlights verified testing practices, CI/CD pipelines, and release milestones.
* Filter out abandoned hobby forks and elevate your high-impact, maintained original software.
* Pair it with our scheduled **GitHub Action** to keep your profile permanently auto-synced.

### 2. 🤖 Native AI Agent Skill & MCP Server (Model Context Protocol)
AI coding assistants (Claude Desktop, Cursor, Antigravity, OpenAI Agents) lack a clean, bounded, rate-limit-conscious way to inspect developer experience:
* GitIntel acts as an **AI Tool / MCP Server**: an agent queries `analyze_github_developer(username)` and receives a pristine, structured `EvidenceContext`.
* Eliminates prompt token bloat by summarizing 100+ repositories into a bounded, high-signal evidence dossier.
* "Copy as AI Prompt" button: one-click copy of pre-engineered, token-optimized context to feed ChatGPT or Claude.

### 3. 📊 6 Verifiable Deterministic Engineering Signals
Every signal is scored conservatively, disclosing supporting repositories, concrete metrics, and explicit limitations:

| Signal | What GitIntel Actually Inspects | Honest Limitation Disclosure |
| :--- | :--- | :--- |
| 🔄 **Code Evolution** | Timestamp recency across pushes & repository updates. | Does not inspect diff size or commit volume in REST mode. |
| 🛡️ **Maintenance** | Active updates, commit freshness, and archived status. | Inactivity does not imply poor maintenance; library stability respected. |
| 🧪 **Testing Evidence** | Bounded Git tree inspection for test directories (`tests/`, `spec/`) & CI workflows. | Does not execute test suites or inspect line coverage. |
| 🤝 **Collaboration** | Incoming forks, separate verified PR counts vs. issue counts, contributor bounds. | Out-of-org contributions require authenticated GraphQL token. |
| 📖 **Documentation** | Bounded tri-state README detection, GitHub Pages, Wiki, and LICENSE metadata. | Does not evaluate prose quality or documentation completeness. |
| 📦 **Release & Delivery** | Verified published GitHub releases, release counts, and semantic version tagging. | Bounded to high-signal active original repositories. |

### 4. 💼 Recruiter-Ready Tech Resume & Portfolio Builder
* Export clean, evidence-backed Markdown and HTML summaries ready for job applications.
* Features curated projects selected by algorithmic signal scoring, complete with technical justification tags (*"Verified CI/CD"*, *"Active Releases"*, *"Multi-contributor"*).

---

## ⚡ Quick Start (Up in 60 Seconds)

### Prerequisites
* **Go**: 1.26+ • **Node.js**: 24+ • **npm**: 11+

### Option A: Windows (One-Click Launcher)
Simply double-click or run:
```bat
run.bat
```
> `run.bat` automatically verifies Go/Node environments, installs frontend dependencies, checks port availability (`8080` & `5173`), launches both servers, and opens your browser.

### Option B: Linux & macOS (Terminal)
```bash
# 1. Clone repository
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel

# 2. Start Backend (Terminal 1)
cd backend
go run ./cmd/server
# Backend listening on http://localhost:8080

# 3. Start Frontend (Terminal 2)
cd frontend
npm ci
npm run dev -- --host 127.0.0.1 --strictPort
# Open http://localhost:5173 in your browser
```

---

## 🤖 AI Agent & MCP Integration

GitIntel exposes a clean, typed evidence contract ([`EvidenceContext`](backend/internal/llm/evidence.go)) specifically engineered for consumption by LLMs and Agentic IDEs.

### Connect to Claude Desktop or Cursor (MCP Configuration)
Add GitIntel to your `claude_desktop_config.json` or Cursor MCP settings:

```json
{
  "mcpServers": {
    "gitintel": {
      "command": "go",
      "args": ["run", "./backend/cmd/server", "--mcp-mode"],
      "env": {
        "PORT": "8080"
      }
    }
  }
}
```

Now, your AI agent can natively execute:
> *"Analyze @Hosseinghorbani0's GitHub profile with GitIntel and write a tailored cover letter highlighting their verified Go testing and CI/CD experience."*

---

## 🏗️ Architecture & Engineering Hardening

GitIntel is built with extreme attention to data integrity, memory safety, and rate-limit defense:

```text
┌────────────────────────────────────────────────────────┐
│                   React 19 + Vite                      │
│     (TypeScript, Lucide React, Vazirmatn RTL/LTR)      │
│                http://localhost:5173                   │
└───────────────────────────▲────────────────────────────┘
                            │ /api reverse proxy (with AbortController cancellation)
┌───────────────────────────▼────────────────────────────┐
│                    Go 1.26 Backend                     │
│                 http://localhost:8080                  │
│                                                        │
│  ┌──────────────────┐  ┌─────────────────────────────┐ │
│  │   HTTP Handler   │  │   Analytics Signal Engine   │ │
│  │  (api/handler.go)│  │   (analytics/engine.go)     │ │
│  └────────▲─────────┘  └──────────────▲──────────────┘ │
│           │                           │                │
│  ┌────────▼─────────┐  ┌──────────────▼──────────────┐ │
│  │ Bounded LRU Cache│  │  Rate-Guarded GitHub Client │ │
│  │ (1,000 entries)  │  │  (ETag 304, Budget Defense) │ │
│  └──────────────────┘  └──────────────▲──────────────┘ │
└───────────────────────────────────────┼────────────────┘
                                        │ HTTPS (REST + ETag)
                         ┌──────────────▼──────────────┐
                         │      GitHub REST API        │
                         │    api.github.com/v3        │
                         └─────────────────────────────┘
```

* **Zero-Allocation Cache**: Bounded 1,000-entry in-memory LRU cache serves typed immutable snapshots without JSON serialization penalties.
* **Proactive Rate-Limit Defense**: Reserves a 10-call safety buffer. Automatically falls back to aggregate metadata if the unauthenticated 60 req/hr budget is low.
* **Conditional ETag Requests**: Cached GitHub GET headers re-issue `If-None-Match`; HTTP 304 responses consume **zero** rate-limit quota.
* **Race-Condition Immune UI**: `AbortController` request cancellation ensures rapid user searches never resolve out of order.
* **Bilingual & Accessible**: Full English & Persian (فارسی) support with bidirectional layout isolation (`<bdi>`), Vazirmatn typography, and reduced-motion compliance.

---

## 🧪 Testing & Verification

GitIntel enforces strict continuous integration standards:

```bash
# Backend unit, integration, and allocation benchmarks
cd backend
go vet ./...
go test ./... -v -count=1

# Frontend TypeScript compilation & ESLint validation
cd frontend
npm run lint
npm run build
```

---

## 🗺️ Roadmap & Community

We are actively executing our canonical [ENGINEERING_ROADMAP.md](ENGINEERING_ROADMAP.md).

### 🚀 What We're Building Next
* [ ] **Dynamic GitHub Profile README Themes**: Bento Grid, Terminal Hacker, Minimalist, and Enterprise formats.
* [ ] **Native MCP Server Package (`cmd/mcp`)**: Full JSON-RPC 2.0 tool endpoints for AI agents.
* [ ] **Automated GitHub Action**: Scheduled workflow to rebuild and commit profile READMEs.
* [ ] **GraphQL External Contribution Mining**: Ingesting non-owned PR contributions to organizations like Kubernetes and React.
* [ ] **PDF / HTML Portfolio Export**: One-click printable resume downloads with custom CSS styling.

### Contributing
We welcome contributions from developers worldwide! Check out [CONTRIBUTING.md](CONTRIBUTING.md) to pick up open tasks, propose new signals, or improve AI prompt templates.

---

## ⚖️ License & Acknowledgements

Distributed under the **MIT License**. See [LICENSE](LICENSE) for details.

<div align="center">

**If you believe in transparent, honest, evidence-based developer metrics, give GitIntel a ⭐️ on GitHub!**

</div>
