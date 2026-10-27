<div align="center">

# ⚡ GitIntel

### Deterministische GitHub-Engineering-Intelligence & Portfolio-Engine

**Die verifizierbare Nachweisschicht für Softwareingenieure, technische Recruiter und KI-Coding-Agenten.**

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](../../LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](../../CONTRIBUTING.md)

**[🌐 Sprache](#-sprachversionen)** — [English](../../README.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [中文](README.zh-CN.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [日本語](README.ja.md) · [Русский](README.ru.md) · [한국어](README.ko.md)

</div>

---

## 💡 Das Problem und die Lösung

Die Bewertung von Entwicklern bewegt sich heute zwischen zwei fehlerhaften Extremen:

- **Eitelkeitsmetriken** — manipulierte Commit-Serien, leere Forks und Spielzeug-Repositories, um Sprachstatistiken künstlich aufzublähen.
- **LLM-Halluzinationen** — undurchsichtige Blackbox-Bewertungen und generische Floskeln ohne jeden überprüfbaren Beleg.

**GitIntel** bietet den dritten Weg: 100% beobachtbare, deterministische Engineering-Signale, die direkt aus echten GitHub-Artefakten abgeleitet werden.

---

## 🚀 Kernfunktionen

### 1. 🌟 GitHub Profile-README-Generator
Verwandeln Sie Ihre `username/username`-Profilseite in ein verifiziertes technisches Aushängeschild:
- Erstellt automatisch eine `README.md` basierend auf realen Entwicklungsdaten.
- Hebt Testpraktiken, CI/CD-Pipelines und Release-Historien hervor.
- Filtert verwaiste Forks heraus und fokussiert auf Ihre wirkungsvollsten Originalprojekte.

### 2. 🤖 Nativer MCP-Server (Model Context Protocol)
Ermöglicht KI-Assistenten die strukturierte Abfrage von GitHub-Profilen ohne unnötigen API-Verbrauch:
- JSON-RPC-Werkzeuge: `gitintel_analyze_profile`, `gitintel_get_signals`, `gitintel_generate_profile_readme`, `gitintel_get_ai_prompt`.
- Vollständig kompatibel mit **Claude Desktop**, **Cursor**, **Antigravity** und jedem MCP-fähigen Agenten.

### 3. 📊 Sechs verifizierbare Engineering-Signale

| Signal | Was GitIntel tatsächlich prüft | Ehrliche Systemgrenze |
| :--- | :--- | :--- |
| 🔄 **Code-Evolution** | Aktualität von Push- und Update-Zeitstempeln | Analysiert nicht das reine Commit-Volumen |
| 🛡️ **Wartung & Pflege** | Aktualisierungsfrequenz; archivierte Repos als „Abgeschlossen“ markiert | Vorübergehende Inaktivität ≠ schlechte Pflege |
| 🧪 **Test-Nachweise** | Git-Tree-Scan nach Testordnern und CI-Workflows | Führt keine Tests aus und misst keine Testabdeckung |
| 🤝 **Kollaboration** | Verifiziertes PR/Issue-Verhältnis, Mitwirkende, eingehende Forks | Externe Beitragsanalysen erfordern GraphQL |
| 📖 **Dokumentation** | Drei-Zustands-Prüfung von README, GitHub Pages, Wiki und LICENSE | Bewertet nicht die stilistische Qualität der Texte |
| 📦 **Release & Delivery** | Verifizierte GitHub-Releases und semantische Versions-Tags | Beschränkt auf aktive Original-Repositories |

### 4. 💼 Recruiter-fähiger Lebenslauf & Portfolio
- Exportieren Sie faktenbasierte Markdown-Zusammenfassungen für Bewerbungen und Vorstellungsgespräche.
- Ein-Klick-Funktion **„Als KI-Prompt kopieren“** für token-effizienten Kontext in Sprachmodellen.

---

## ⚡ Schnellstart

### Voraussetzungen

| Werkzeug | Mindestversion |
| :--- | :--- |
| Go | 1.26 |
| Node.js | 24 |
| npm | 11 |

### Windows — 1-Klick-Starter
```bat
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel
run.bat
```

### Linux & macOS
```bash
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel

# Terminal 1 — Backend
cd backend
go run ./cmd/server

# Terminal 2 — Frontend
cd frontend
npm ci
npm run dev
```

### GitHub Access Token (Optional)
```bash
cp .env.example .env
# .env anpassen: GITHUB_TOKEN=ghp_your_token_here
```

---

## 🤖 MCP-Integration

### Verbindung mit Claude Desktop
Fügen Sie Folgendes zu Ihrer `claude_desktop_config.json` hinzu:

```json
{
  "mcpServers": {
    "gitintel": {
      "command": "go",
      "args": ["run", "./backend/cmd/mcp"],
      "cwd": "/pfad/zu/GitIntel"
    }
  }
}
```

---

## 🏗️ Systemarchitektur

```text
┌─────────────────────────────────────────┐
│      React 19 + Vite (TypeScript)       │
│         http://localhost:5173           │
└──────────────────┬──────────────────────┘
                   │ proxy /api
┌──────────────────▼──────────────────────┐
│          Go 1.26 Backend                │
│        http://localhost:8080            │
│  LRU-Cache (1.000 Einträge) +           │
│  GitHub-Client (ETag · Rate Guard)      │
└──────────────────┬──────────────────────┘
                   │ REST + ETag 304
        ┌──────────▼──────────┐
        │   GitHub REST API   │
        └─────────────────────┘
```

**Technische Besonderheiten:**
- **Allokationsfreier In-Memory-Cache** — 1.000 LRU-Slots mit unveränderlichen Daten-Snapshots.
- **Proaktiver Rate-Limit-Schutz** — Reserviert 10 Aufrufe Sicherheitsabstand vor Ratenbegrenzung.
- **Bedingte ETag-Abfragen** — HTTP-304-Antworten verbrauchen **null** API-Kontingent.
- **Race-Condition-resistentes Frontend** — Sichere Anfrageannullierung via `AbortController`.
- **Zweisprachig und barrierefrei** — Vollständige RTL-Unterstützung (Persisch/Arabisch) sowie Englisch.
- **Lokale Sicherheitsisolation** — CORS strikt beschränkt auf `127.0.0.1:5173`.

---

## 🧪 Tests

```bash
# Backend-Tests
cd backend
go vet ./...
go test ./... -v -count=1

# Frontend-Validierung
cd frontend
npm run lint
npm run build
```

Die Testsuite des Backends umfasst **109 Tests** in 7 Paketen.

---

## 🗺️ Roadmap

Detaillierte Meilensteine finden Sie in [ENGINEERING_ROADMAP.md](../../ENGINEERING_ROADMAP.md).

| Bereich | Status |
| :--- | :--- |
| Core-Berechnungs-Engine & Datenintegrität | ✅ Abgeschlossen |
| LRU-Cache & Credential-Lebenszyklus | ✅ Abgeschlossen |
| Frontend-Stabilität & RTL-Barrierefreiheit | ✅ Abgeschlossen |
| Performance-Benchmarks & Optimierungen | ✅ Abgeschlossen |
| Nativer MCP-Server | ✅ Kern fertig |
| Profile-README-Generator | ✅ Kern fertig |
| Regressionssuite mit echten Fixture-Daten | 🔄 In Arbeit (67%) |
| Sicherheits-Hardening & Auditing | 🔄 In Arbeit (50%) |
| Frontend-Tests mit Vitest | ⬜ Geplant |
| GitHub Action für automatischen Sync | ⬜ Geplant |

---

## 🤝 Mitwirken

Beiträge aus der Community sind herzlich willkommen! Bitte lesen Sie [CONTRIBUTING.md](../../CONTRIBUTING.md) für Richtlinien und Anweisungen.

---

## 🔒 Sicherheit und Datenschutz

- GitIntel läuft **ausschließlich lokal (localhost)**. Weder Token noch Code verlassen Ihren Rechner.
- CORS-Richtlinien sind strikt an die lokale Umgebung gebunden.
- Schutzmechanismen gegen Prompt-Injection in der LLM-Schicht integriert.

---

## 📄 Lizenz

Veröffentlicht unter der **GNU Affero General Public License v3.0 (GNU AGPL-3.0)** — weitere Einzelheiten in der Datei [LICENSE](../../LICENSE).

---

## 🌐 Sprachversionen

| Sprache | Datei |
| :--- | :--- |
| 🇺🇸 English (kanonisch) | [README.md](../../README.md) |
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
