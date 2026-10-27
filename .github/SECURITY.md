# Security Policy

## Supported Versions

GitIntel is actively developed as a local-first engineering intelligence engine. Security updates are applied to the active branch and releases:

| Version | Supported          |
| ------- | ------------------ |
| 1.5.x   | :white_check_mark: |
| 1.0.x   | :white_check_mark: |
| < 1.0.0 | :x:                |

## Reporting a Vulnerability

We take the security of GitIntel and its users seriously. As a local-first tool designed to handle personal GitHub access tokens and repository metadata, privacy and isolation are primary engineering priorities.

If you discover a security vulnerability in GitIntel:

1. **Do not disclose it publicly** via GitHub Issues, PRs, or public forums.
2. Please report it privately through **[GitHub Private Security Advisories](https://github.com/Hosseinghorbani0/GitIntel/security/advisories/new)**.
3. Alternatively, contact the maintainer directly at: `hosseingh1068@gmail.com` with the subject line `[SECURITY] GitIntel Vulnerability Report`.

### What to Include in Your Report
- A description of the issue and potential impact.
- Clear reproduction steps or proof-of-concept (POC).
- Affected components (backend HTTP server, MCP server, cache, or frontend).
- Any recommended remediation.

### Response Timeline
- **Initial Acknowledgement**: Within 48 hours.
- **Triage & Status Update**: Within 5 business days.
- **Resolution & Release**: Coordinated disclosure after remediation is published.

## Security Architecture Principles
- **Localhost Isolation**: The backend server is strictly bound to localhost and enforces CORS restrictions (`127.0.0.1:5173`).
- **Zero Remote Storage**: Personal Access Tokens (PATs) and repository data are never sent to third-party tracking servers.
- **Prompt Injection Defense**: Untrusted GitHub metadata (commit messages, issue descriptions, profile bios) is strictly bounded and isolated from instruction tokens when passed to LLMs.
