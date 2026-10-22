# Contributing to GitIntel

Thank you for your interest in contributing to **GitIntel**!

GitIntel is an open-source, local-first platform designed to transform public GitHub profile metadata into **transparent, deterministic engineering signals and portfolio summaries**.

---

## 1. Core Philosophy & Guiding Rules

1. **Observable Evidence Only**: We evaluate only observable metadata and artifacts directly extracted from GitHub. We never claim to measure human coding intellect, evaluate psychological traits, or detect AI-generated code.
2. **Honesty About Missing Data**: When metadata cannot be collected (e.g. rate limits or unqueried endpoints), we record it as `unverified` or `"Insufficient data"`. Never assume negative evidence when data is merely missing.
3. **No Mocks Masking Production**: Unit and integration tests must validate real behavior. Production clients must never hardcode assumptions.
4. **Zero Secret Leaks**: Never commit `.env` files, API keys, personal access tokens, or credentials in tests, examples, or screenshots.

---

## 2. Help Wanted: High-Priority Open Areas

We actively invite developers, researchers, and open-source contributors to collaborate! Our canonical task backlog is tracked in [ENGINEERING_ROADMAP.md](ENGINEERING_ROADMAP.md).

Here are the highest-impact open areas where we welcome Pull Requests:

### A. Dynamic GitHub Profile README Generator (`GI-AI-001`)
* **Goal**: Build preset markdown & SVG templates for users' `username/username` profile READMEs.
* **Tasks**:
  * Implement theme generators in `backend/internal/analytics/` (Minimalist, Bento Grid, Hacker Terminal).
  * Expose an endpoint `/api/profile-readme` and integrate a live preview tab in `PortfolioView.tsx`.

### B. Native Model Context Protocol (MCP) Server (`GI-AI-002`)
* **Goal**: Enable AI coding assistants (Claude Desktop, Cursor, Antigravity) to query GitIntel natively via JSON-RPC.
* **Tasks**:
  * Build an MCP stdio/SSE server in `backend/cmd/mcp/` exposing `gitintel_analyze_profile` and `gitintel_get_signals`.
  * Validate with Claude Desktop and Cursor MCP configs.

### C. Automated GitHub Action Workflow (`GI-AI-003`)
* **Goal**: A composite GitHub Action (`gitintel-profile-action`) that runs on a weekly cron in a developer's profile repository to automatically regenerate and commit their `README.md`.

### D. "Copy as AI Prompt" Exporter & PDF Resume (`GI-AI-004` & `GI-AI-005`)
* **Goal**: Give developers one-click prompt templates to feed ChatGPT/Claude, plus print-ready `@media print` CSS for 1-page PDF resumes.

### E. Multi-Repository External Contribution Mining (`GI-PROD-001`)
* **Goal**: Query GitHub's GraphQL API (`is:pr author:<username>`) to detect contributions made to external open-source organizations (Kubernetes, React, Linux) beyond owned repositories.

### F. Automated Frontend Tests with Vitest (`GI-TEST-003`)
* **Goal**: Set up Vitest and React Testing Library in `frontend/` to test component rendering, language switching, and filter predicates.

---

## 3. Local Development Setup

### Prerequisites
* **Go**: 1.26+ (`go version`)
* **Node.js**: 24+ (`node --version`)
* **npm**: 11+ (`npm --version`)
* **Git**: 2.30+

### Starting Backend & Frontend
* **Windows**:
  Run `run.bat` or `dev.bat` from the repository root:
  ```bat
  run.bat
  ```
* **Linux / macOS**:
  ```bash
  # Terminal 1: Backend
  cd backend
  go run ./cmd/server

  # Terminal 2: Frontend
  cd frontend
  npm ci
  npm run dev -- --host 127.0.0.1 --strictPort
  ```

* Default ports:
  * Backend: `http://localhost:8080` (Health check: `http://localhost:8080/api/health`)
  * Frontend: `http://localhost:5173`

---

## 4. Validation Before Submitting

Every Pull Request must pass local validation before submission. Run:

### Backend Checks
```bash
cd backend
go vet ./...
go test ./... -count=1
```

### Frontend Checks
```bash
cd frontend
npm run lint
npm run build
```

---

## 5. Contribution Workflow

1. **Pick or Propose a Task**:
   * Review [ENGINEERING_ROADMAP.md](ENGINEERING_ROADMAP.md) to pick an open task (e.g. `GI-DATA-007`).
   * Or open a GitHub Issue to discuss a new bug fix or feature idea.
2. **Branch from `main`**:
   * Create a descriptive branch:
     ```bash
     git checkout -b feat/release-evidence-collection
     ```
3. **Keep Commits Atomic & Focused**:
   * Follow [Conventional Commits](https://www.conventionalcommits.org/):
     * `feat(...)`: New capability or signal
     * `fix(...)`: Bug fix or rate-limit remediation
     * `test(...)`: Adding or updating test suites
     * `docs(...)`: Documentation or roadmap updates
     * `ci(...)`: GitHub Actions or automation changes
   * **Rule of Thumb**: Keep PRs small (1–2 files per PR, or clearly isolated changes) so they are fast and easy to review.
4. **Update Documentation**:
   * When completing a task from `ENGINEERING_ROADMAP.md`, update its checkbox and paste the validation evidence.
5. **Open a Pull Request**:
   * Push your branch to your fork and submit a PR to `main`.
   * Include a clear description of what changed, why, and the command output confirming tests pass.
