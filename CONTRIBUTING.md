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

## 2. Help Wanted: High-Priority Areas

We actively invite developers, researchers, and technical writers to contribute! Our work is organized into an authoritative roadmap in [ENGINEERING_ROADMAP.md](ENGINEERING_ROADMAP.md).

Here are the highest-impact open areas where we need help right now:

### A. Release Evidence Remediation (`GI-DATA-007` to `GI-DATA-010`)
* **Problem**: Currently, `repo.HasReleases` is omitted during repository normalization in `backend/internal/github/client.go:150-179`, leaving the Release/Delivery signal permanently reporting `"Insufficient data"`.
* **Tasks**:
  * Implement bounded release queries (check tags/releases only for top candidate repos, max 5, preserving rate limits).
  * Update the `Repository` struct with `ReleaseStatus` (`verified`, `missing`, `unverified`) or `ReleaseCount`.
  * Connect `computeReleaseActivity()` in `backend/internal/analytics/engine.go` to reflect verified releases.

### B. Language Byte Volume Weighting (`GI-DATA-011`)
* **Problem**: Currently, language distribution counts repositories (`counts[repo.Language]++`), which distorts specialization (e.g. 5 small Python helper scripts can outweigh 1 large Go project).
* **Task**: Query GitHub's `/repos/{owner}/{repo}/languages` endpoint for candidate repositories to aggregate code volume by bytes, with a graceful fallback if rate limits are constrained.

### C. Engineering Signal Refinement (`GI-SIGNAL-001` to `GI-SIGNAL-003`)
* **Problem**: Testing signals currently use repository-name heuristics (`test`, `spec`), which can yield false positives. Documentation signals check for `"docs"` substrings in repo descriptions.
* **Task**: Replace name heuristics with observable indicators: CI workflow file presence (`.github/workflows/*.yml`), test directories (`tests/`, `*_test.go`), and verified GitHub Pages flags (`repo.HasPages`).

### D. Frontend Modernization & Polish (Phase 7)
* Decompose monolithic `frontend/src/App.tsx` into reusable components (`ProfileCard`, `SignalsList`, `RepositoryGrid`, `MarkdownExport`).
* Add `AbortController` support to cancel in-flight API requests when the user rapidly changes usernames.

### E. Test Coverage & Fixtures
* Add frontend component tests using Vitest or React Testing Library.
* Provide real, sanitized GitHub API JSON fixtures for offline integration testing.

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
