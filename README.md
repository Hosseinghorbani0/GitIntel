# GitIntel

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![Node Version](https://img.shields.io/badge/Node.js-24-339933?logo=nodedotjs&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

> Deterministic, evidence-grounded GitHub engineering intelligence and developer portfolio engine.

GitIntel transforms public GitHub profile and repository metadata into structured, transparent engineering signals, repository metrics, and markdown-formatted portfolio and resume summaries.

---

## What Problem GitIntel Solves

Developer assessment tools typically suffer from two extremes:
1. **Superficial Vanity Metrics**: Over-relying on stars, follower counts, and green commit squares without context.
2. **Deceptive AI Scoring**: Claiming opaque algorithms can determine human coding intellect or detect AI-generated code.

GitIntel takes a radically honest, deterministic approach:
* **Observable Evidence Only**: Evaluates only observable GitHub metadata and repository artifacts.
* **Conservative Scoring**: Uses transparent weighted models where every score includes its underlying metrics, evidence strings, and explicit methodology limitations.
* **Missing vs. Negative Evidence**: Strictly differentiates between absent features and unqueried/rate-limited metadata.

---

## Engineering Signals

GitIntel calculates 6 core engineering signals categorized by level (`Strong`, `Moderate`, `Limited`, or `Insufficient data`):

| Signal | Description | Observable Evidence | Current Limitations |
| :--- | :--- | :--- | :--- |
| **Code Evolution** | Commit activity cadence and active repository maintenance lifespan. | Commit timestamps, creation dates, and push frequency across repos. | Does not inspect individual commit diffs or line counts. |
| **Maintenance** | Active upkeep versus abandoned or archived repositories. | Archived status, recent push dates, active issue resolution. | Dependent on public GitHub activity windows. |
| **Testing Evidence** | Presence of automated testing artifacts. | Name-marker heuristics across repositories. | Test assertions, coverage percentages, and test runs are not executed. |
| **Collaboration** | Team and open-source interaction indicators. | Forked repositories, open issues, and pull request activity. | Code review depth and discussion quality are not parsed. |
| **Documentation** | Codebase documentation completeness and discoverability. | Bounded tri-state README verification (`verified`, `missing`, `unverified`), license existence, and docs markers. | Detailed README verification is capped at the top 5 repositories to preserve GitHub rate limits. |
| **Release / Delivery** | Production delivery and version tagging practices. | Release tags and published assets. | Currently reports "Insufficient data" pending release collection remediation ([GI-DATA-007](ENGINEERING_ROADMAP.md)). |

### Bounded README Verification & Rate Limits
To prevent exhausting GitHub API rate limits (60 requests/hour unauthenticated; 5,000/hour authenticated with PAT):
* Repositories start in an `unverified` README status.
* The client performs deep verification (`/repos/{owner}/{repo}/readme`) for a bounded candidate set of up to **5 top original repositories**.
* Any HTTP 403 / 429 response gracefully falls back to `unverified` rather than failing the analysis or misreporting missing documentation.

---

## Features

* **Public GitHub Profile Analysis**: Fast, non-destructive metadata ingestion via the GitHub REST API.
* **Deterministic Signal Scoring**: Transparent evidence arrays, confidence ratings, and limitations disclosures.
* **Language Distribution**: Breakdown of top languages across active repositories.
* **Automated Exports**:
  * Markdown Resume summary ready for developer profiles.
  * Markdown Portfolio summary highlighting featured projects and justifications.
* **Optional LLM Interpretation**: Guardrailed, bounded narrative interpretation (disabled by default; requires explicit provider configuration).
* **Local-First & Privacy Preserving**: Runs entirely on `localhost`. Personal Access Tokens (PATs) entered in the UI are never stored in `localStorage`, database, or server logs. Private repositories are filtered out and not retained.

---

## Tech Stack & Architecture

```text
┌────────────────────────────────────────────────────────┐
│                   React 19 + Vite                      │
│         (TypeScript, Lucide React, Vazirmatn)          │
│                http://localhost:5173                   │
└───────────────────────────▲────────────────────────────┘
                            │ /api reverse proxy
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
│  │ In-Memory Cache  │  │     GitHub API Client       │ │
│  │ (5-min TTL map)  │  │  (google/go-github/v72)     │ │
│  └──────────────────┘  └──────────────▲──────────────┘ │
└───────────────────────────────────────┼────────────────┘
                                        │ HTTPS (REST)
                         ┌──────────────▼──────────────┐
                         │      GitHub REST API        │
                         │    api.github.com/v3        │
                         └─────────────────────────────┘
```

* **Backend**: Go 1.26 (`cmd/server`), `net/http` with CORS, `google/go-github/v72`.
* **Frontend**: React 19, TypeScript, Vite 8, Lucide React icons, Vazirmatn typography.
* **Caching**: In-memory thread-safe TTL cache (5-minute default window).
* **Optional AI Layer**: OpenAI-compatible / Hugging Face router interface with prompt bounds.

---

## Quick Start

### Prerequisites
* **Go**: 1.26 or newer
* **Node.js**: 24 or newer
* **npm**: 11 or newer

### Option A: Windows (One-Click Launcher)
Run [run.bat](run.bat) from the root directory:
```bat
run.bat
```
`run.bat` automatically:
1. Validates Go, Node.js, and npm in `PATH`.
2. Checks port availability (default backend `8080`, frontend `5173`).
3. Runs `npm ci` in `frontend/` if dependencies are not yet installed.
4. Starts backend and frontend in separate command windows.
5. Polls `/api/health` until ready and opens `http://localhost:5173/` in your browser.

*(To stop the application, simply close the opened Backend and Frontend command windows.)*

### Option B: Linux / macOS / Manual Setup

```bash
# 1. Clone repository
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel

# 2. Start Backend
cd backend
go run ./cmd/server
# Backend listens on http://localhost:8080

# 3. Start Frontend (in a new terminal)
cd frontend
npm ci
npm run dev -- --host 127.0.0.1 --strictPort
# Frontend opens at http://localhost:5173
```

---

## Environment Variables

Copy `.env.example` to `.env` in the repository root to customize your configuration:

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `PORT` | `8080` | Backend HTTP listening port. |
| `GITINTEL_BACKEND_PORT` | `8080` | Launcher backend port override (`run.bat`). |
| `GITINTEL_FRONTEND_PORT` | `5173` | Launcher frontend port override (`run.bat`). |
| `GITINTEL_API_URL` | `http://localhost:8080` | Target URL for Vite `/api` reverse proxy. |
| `LLM_ANALYSIS_ENABLED` | `false` | Enable/disable optional LLM narrative interpretation. |
| `LLM_PROVIDER_STATUS` | `payment_required` | Safe provider state (`available`, `rate_limited`, `payment_required`, etc.). |
| `LLM_PROVIDER` | `huggingface` | Provider type (`huggingface` or generic OpenAI-compatible). |
| `HF_BASE_URL` | `https://router.huggingface.co/v1` | Hugging Face router endpoint. |
| `HF_MODEL` | `Qwen/Qwen3.8-2.4T-A95B:novita` | Model identifier. |
| `HF_API_KEYS` | *(empty)* | Comma-separated API keys for provider. |
| `LLM_BASE_URL` | *(empty)* | Custom OpenAI-compatible endpoint. |
| `LLM_MODEL` | *(empty)* | Custom model name. |
| `LLM_API_KEY` | *(empty)* | Single API key for custom endpoint. |

> [!WARNING]
> Never commit `.env` or files containing API keys or personal access tokens. GitIntel's `.gitignore` explicitly excludes `.env` and credential files.

---

## Running Tests & Linting

### Backend
```bash
cd backend

# Run Go static analysis
go vet ./...

# Run unit and regression tests
go test ./... -count=1
```

### Frontend
```bash
cd frontend

# Run ESLint
npm run lint

# Run TypeScript typecheck & production build
npm run build
```

---

## Project Structure

```text
GitIntel/
├── .github/
│   └── workflows/
│       ├── backend.yml       # GitHub Actions: Go vet & test
│       └── frontend.yml      # GitHub Actions: Node lint & Vite build
├── backend/
│   ├── cmd/server/main.go    # HTTP server, routing, and CORS middleware
│   ├── internal/
│   │   ├── analytics/        # Deterministic signal calculation & tests
│   │   ├── api/              # HTTP request handlers & API contracts
│   │   ├── cache/            # In-memory TTL cache
│   │   ├── github/           # GitHub REST client with bounded README checks
│   │   └── llm/              # Optional LLM integration & status endpoints
│   ├── go.mod                # Go module definition (go 1.26)
│   └── go.sum                # Go checksums
├── frontend/
│   ├── src/
│   │   ├── App.tsx           # Main application view & signal dashboard
│   │   ├── types.ts          # TypeScript domain models (ReadmeStatus, Signals)
│   │   ├── i18n.ts           # Dual-language support (English / Persian)
│   │   └── services/         # API client layer
│   ├── package.json          # Node dependencies (React 19, Vite 8)
│   └── vite.config.ts        # Vite configuration & backend proxy
├── CONTRIBUTING.md           # Contribution guidelines & open tasks
├── ENGINEERING_ROADMAP.md    # Canonical remediation roadmap & checklist
├── dev.bat / run.bat         # Automated Windows launchers
└── README.md                 # Project documentation
```

---

## Contributing & Help Wanted

GitIntel is in active remediation and development under our [ENGINEERING_ROADMAP.md](ENGINEERING_ROADMAP.md). We actively welcome contributions from developers of all experience levels!

### Open Areas Ready for Collaboration
* **Release Evidence Remediation ([GI-DATA-007](ENGINEERING_ROADMAP.md) to [GI-DATA-010](ENGINEERING_ROADMAP.md))**:
  * Identify repositories with releases and populate `ReleaseStatus` / `ReleaseCount`.
  * Transition the Release/Delivery signal from "Insufficient data" to genuine evidence.
* **Language Byte Volume Weighting ([GI-DATA-011](ENGINEERING_ROADMAP.md))**:
  * Query `/repos/{owner}/{repo}/languages` to weight developer languages by actual codebase bytes instead of raw repository counts.
* **Signal Artifact Detection ([GI-SIGNAL-001](ENGINEERING_ROADMAP.md) to [GI-SIGNAL-003](ENGINEERING_ROADMAP.md))**:
  * Replace name-based testing heuristics with verified test directories (`tests/`, `spec/`) and CI workflow detection (`.github/workflows/`).
* **Frontend Component Decomposition**:
  * Break down the monolithic `App.tsx` into modular React components and add `AbortController` request cancellation.
* **Automated Frontend Testing**:
  * Set up Vitest or React Testing Library for frontend component testing.

See [CONTRIBUTING.md](CONTRIBUTING.md) for full instructions on local setup, picking roadmap tasks, coding guidelines, and submitting small, reviewable pull requests.

---

## License

This project is licensed under the [MIT License](https://opensource.org/licenses/MIT).
