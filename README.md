# GitIntel

GitHub Engineering Intelligence

GitIntel is a local MVP that turns a GitHub profile into structured engineering signals, repository intelligence, and portfolio or resume-ready outputs. It emphasizes observable evidence from public GitHub metadata while avoiding claims that GitHub data can prove human authorship or AI usage.

## Features

- Real GitHub profile analysis using the public GitHub API
- Repository metrics, language distribution, and usage stats
- Deterministic engineering signal engine
- Resume and portfolio markdown generation
- Repository link export
- Local Windows-friendly setup with no Docker or cloud dependencies

## Architecture

- Backend: Go
- Frontend: React + TypeScript + Vite
- Local cache: in-memory TTL cache
- Data source: GitHub REST API via google/go-github

## Requirements

- Windows 10 or 11
- Go 1.26+
- Node.js 24+
- npm
- Chrome or Edge

## Windows Setup

1. Clone the repository.
2. Open a terminal in the project root.
3. Run:

   ```bat
   run.bat
   ```

`run.bat` checks the local toolchain and default ports, installs frontend dependencies with `npm ci` if needed, and starts the Go API and Vite in separate visible command windows. It waits for `/api/health` and the Vite page before opening the frontend. Vite proxies `/api` to the backend port selected by the launcher. If a port is occupied or a service fails to become ready, the launcher reports the failure and does not terminate any process. Close the GitIntel Backend and GitIntel Frontend command windows to stop the services.

## Environment Variables

- `PORT` optional: backend port, defaults to `8080`
- `GITINTEL_BACKEND_PORT` optional: `run.bat` backend port, defaults to `8080`
- `GITINTEL_FRONTEND_PORT` optional: `run.bat` frontend port, defaults to `5173`
- `GITINTEL_API_URL` optional: Vite API proxy target; set automatically by `run.bat`
- `HF_API_KEYS` optional: comma-separated Hugging Face credentials for the local model test endpoint
- `HF_BASE_URL`, `HF_MODEL`, and `LLM_PROVIDER` optional: Hugging Face-compatible provider configuration
- `LLM_ANALYSIS_ENABLED` defaults to `false`; enable only after configuring and verifying an available provider/model
- `LLM_PROVIDER_STATUS` accepts safe states such as `payment_required`, `rate_limited`, `unavailable`, `unverified`, `disabled`, or `available`; the known Hugging Face/Qwen model defaults to its last verified payment-required state
- `LLM_BASE_URL`, `LLM_MODEL`, `LLM_API_KEY`, and `LLM_API_KEYS` support a generic OpenAI-compatible endpoint

Copy `.env.example` to the workspace-root `.env` and set local values there. The backend loads `.env` from its working directory or the workspace root; process environment variables take precedence. `.env` and local secret-file patterns are ignored by Git and must not be committed. Never put credentials in frontend configuration, source files, tests, or screenshots. Revoke and replace any key pasted into chat, logs, or other shared contexts.

If the launcher reports a port conflict, it leaves the existing listener untouched. Choose alternate ports in the same terminal before running it, for example:

```bat
set GITINTEL_BACKEND_PORT=8081
set GITINTEL_FRONTEND_PORT=5174
run.bat
```

The temporary model test is available at `POST /api/llm/test` with `{ "language": "en" }` or `{ "language": "fa" }`; it uses a fixed fixture. `POST /api/analyze` always runs deterministic analysis and caches bounded evidence. `POST /api/interpret` with `{ "username": "octocat", "language": "en" }` or `{ "username": "octocat", "language": "fa" }` interprets that cached evidence only when `LLM_ANALYSIS_ENABLED=true`. Provider failure never removes the deterministic analysis or report. Credential status is available from `GET /api/llm/status`; responses contain status only, never the configured key list.

## Running Locally

From the project root:

```bat
run.bat
```

Or start each side manually:

```bat
cd backend
go run ./cmd/server
```

```bat
cd frontend
npm install
set GITINTEL_API_URL=http://localhost:8080
set GITINTEL_FRONTEND_PORT=5173
npm run dev -- --host 127.0.0.1 --strictPort
```

## GitHub Token

A personal access token may be entered in the analysis form to authenticate that analysis request. The frontend does not persist it in localStorage, logs, reports, or API responses. Repository results marked private are discarded by the GitHub client before they reach the deterministic analyzer or LLM evidence builder; private-repository access is not a supported analysis mode.

## API

### Health

- `GET /api/health`

### Analysis

- `POST /api/analyze`
  - Body: `{ "username": "octocat", "token": "optional" }`

### Report endpoints

- `GET /api/report/:username`
- `POST /api/resume`
- `POST /api/portfolio`
- `GET /api/github/profile/:username`
- `GET /api/github/repos/:username`

## Engineering Signals

The MVP uses a transparent weighted model and exposes the evidence used to generate each signal. The model is intentionally conservative and does not claim to measure engineering quality or prove human-only coding.

Signal categories:

- Code Evolution
- Maintenance
- Testing Evidence
- Collaboration
- Documentation
- Release / Delivery Signals

Each signal includes:

- value
- confidence
- evidence
- explanation

The user-facing interface describes results as "Strong", "Moderate", "Limited", or "Insufficient data" instead of pretending to offer objective skill scores.

The current analyzer uses repository metadata and timestamps; it does not inspect commits, source files, test directories, CI workflows, README contents, contributor lists, or release endpoints. Test and documentation signals are limited metadata/name hints, while release evidence is unavailable until release metadata is fetched. These signals must not be interpreted as code-quality or competence measurements.

The Hugging Face model-testing endpoint is separate from profile analysis and uses a fixed fixture. The optional interpretation endpoint receives bounded evidence from the most recent deterministic analysis; repository descriptions and source code are excluded.

## Privacy

GitIntel analyzes GitHub data available to the connected account. It does not claim to determine whether code was written by a human or AI. Engineering signals are observable evidence only.

## Limitations

- GitHub API rate limits can affect large portfolios
- Some activity metrics are not reliably available from the public API
- Repository metadata is limited to what the GitHub API exposes
- AI report generation remains disabled until a provider/model is explicitly configured and enabled
- GitHub commit, contributor, CI, README-content, and release metadata are not currently collected by the analyzer
- This is an MVP and does not yet include OAuth, persistent accounts, or team analytics

## Roadmap

- GitHub OAuth integration
- Persistent historical snapshots
- richer activity rollups and repository timeline analytics
- recruiter-ready exports and enterprise-ready features

## Contributing

Open an issue or pull request with a clear explanation and a focused change.

## License

MIT
