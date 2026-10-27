<div align="center">

# ⚡ GitIntel

### Motor de Inteligencia de Ingeniería GitHub y Generador de Portafolio

**La capa de evidencia verificable para ingenieros de software, reclutadores técnicos y agentes de IA.**

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](../../LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](../../CONTRIBUTING.md)

**[🌐 Idioma](#-versiones-de-idioma)** — [English](../../README.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [中文](README.zh-CN.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [日本語](README.ja.md) · [Русский](README.ru.md) · [한국어](README.ko.md)

</div>

---

## 💡 El Problema

La evaluación de desarrolladores vive entre dos extremos rotos:

- **Métricas de vanidad** — racha de commits manipulada, forks vacíos, repositorios de juguete que inflan las estadísticas de lenguaje.
- **IA alucinatoria** — tarjetas de puntaje de caja negra, frases genéricas sin evidencia verificable.

**GitIntel** es la tercera vía: señales 100% observables y deterministas derivadas directamente de los artefactos reales de GitHub.

---

## 🚀 Características

### 1. 🌟 Generador de Profile README para GitHub
Convierte tu página `username/username` en una vitrina de ingeniería verificada:
- Genera automáticamente un `README.md` basado en datos técnicos reales.
- Destaca prácticas de pruebas, pipelines CI/CD e historial de releases.
- Filtra forks abandonados; muestra tu trabajo original más impactante.

### 2. 🤖 Servidor MCP Nativo (Model Context Protocol)
Permite que los asistentes de IA consulten perfiles de GitHub sin consumir rate limit:
- Herramientas JSON-RPC: `gitintel_analyze_profile`, `gitintel_get_signals`, `gitintel_generate_profile_readme`, `gitintel_get_ai_prompt`.
- Compatible con **Claude Desktop**, **Cursor**, **Antigravity** y cualquier agente MCP.

### 3. 📊 Seis Señales de Ingeniería Verificables

| Señal | Qué inspecciona GitIntel | Limitación honesta |
| :--- | :--- | :--- |
| 🔄 **Evolución del código** | Recencia de timestamps de push y actualización | No inspecciona volumen de commits |
| 🛡️ **Mantenimiento** | Frecuencia de actualizaciones; repos archivados marcados "Completado" | Inactividad ≠ mal mantenimiento |
| 🧪 **Evidencia de pruebas** | Escaneo de árbol Git para dirs de pruebas y workflows CI | No ejecuta pruebas ni mide cobertura |
| 🤝 **Colaboración** | Split PR/issue verificado, conteo de contribuidores, forks entrantes | Contribuciones externas requieren GraphQL |
| 📖 **Documentación** | Detección tri-estado de README, GitHub Pages, Wiki, LICENSE | No evalúa calidad de prosa |
| 📦 **Release y entrega** | Releases GitHub verificados y tags de versión semántica | Limitado a repos originales activos |

### 4. 💼 Currículum y Portafolio Listo para Reclutadores
- Exporta resúmenes Markdown respaldados por evidencia para solicitudes de empleo.
- Un clic **"Copiar como prompt de IA"** para contexto optimizado en tokens.

---

## ⚡ Inicio Rápido

### Prerrequisitos

| Herramienta | Versión mínima |
| :--- | :--- |
| Go | 1.26 |
| Node.js | 24 |
| npm | 11 |

### Windows — Lanzador con un clic
```bat
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel
run.bat
```

### Linux y macOS
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

### Token de GitHub (Opcional)
```bash
cp .env.example .env
# Edita .env y establece: GITHUB_TOKEN=ghp_your_token_here
```

---

## 🤖 Integración MCP

### Conectar a Claude Desktop
Agrega a tu `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "gitintel": {
      "command": "go",
      "args": ["run", "./backend/cmd/mcp"],
      "cwd": "/ruta/a/GitIntel"
    }
  }
}
```

---

## 🏗️ Arquitectura

```text
┌─────────────────────────────────────────┐
│      React 19 + Vite (TypeScript)       │
│         http://localhost:5173           │
└──────────────────┬──────────────────────┘
                   │ proxy /api
┌──────────────────▼──────────────────────┐
│          Go 1.26 Backend                │
│        http://localhost:8080            │
│  Caché LRU (1.000 entradas) +           │
│  Cliente GitHub (ETag · Rate Guard)     │
└──────────────────┬──────────────────────┘
                   │ REST + ETag 304
        ┌──────────▼──────────┐
        │   GitHub REST API   │
        └─────────────────────┘
```

**Puntos destacados de ingeniería:**
- **Caché sin asignación** — LRU de 1.000 entradas con snapshots inmutables.
- **Defensa proactiva de rate limit** — Reserva un búfer de seguridad de 10 llamadas.
- **Solicitudes condicionales ETag** — HTTP 304 consume **cero** cuota de rate limit.
- **UI inmune a condiciones de carrera** — Cancelación con `AbortController`.
- **Bilingüe y accesible** — Soporte completo en inglés y persa con layout RTL.
- **CORS solo localhost** — El backend solo acepta `127.0.0.1:5173`.

---

## 🧪 Testing

```bash
# Backend
cd backend
go vet ./...
go test ./... -v -count=1

# Frontend
cd frontend
npm run lint
npm run build
```

La suite de pruebas del backend incluye **109 pruebas** en 7 paquetes.

---

## 🗺️ Hoja de Ruta

Consulta el plan completo en [ENGINEERING_ROADMAP.md](../../ENGINEERING_ROADMAP.md).

| Área | Estado |
| :--- | :--- |
| Motor base e integridad de datos | ✅ Completo |
| Caché LRU y ciclo de credenciales | ✅ Completo |
| Confiabilidad frontend y accesibilidad RTL | ✅ Completo |
| Rendimiento y benchmarks | ✅ Completo |
| Servidor MCP nativo | ✅ Core completo |
| Generador de Profile README | ✅ Core completo |
| Pruebas de regresión basadas en fixtures | 🔄 67% |
| Hardening de seguridad | 🔄 50% |
| Pruebas frontend con Vitest | ⬜ Planificado |
| GitHub Action para auto-sync | ⬜ Planificado |

---

## 🤝 Contribuir

Las contribuciones son bienvenidas. Lee [CONTRIBUTING.md](../../CONTRIBUTING.md) para instrucciones de configuración y el flujo de trabajo.

---

## 🔒 Seguridad y Privacidad

- GitIntel funciona **completamente en localhost**. Tu token y datos nunca salen de tu máquina.
- CORS restringido a `127.0.0.1:5173`.
- Defensas contra inyección de prompts integradas en la integración LLM.

---

## 📄 Licencia

Distribuido bajo la **Licencia GNU Affero General Public License v3.0 (GNU AGPL-3.0)** — consulta el archivo [LICENSE](../../LICENSE) para detalles.

---

## 🌐 Versiones de Idioma

| Idioma | Archivo |
| :--- | :--- |
| 🇺🇸 English (canónico) | [README.md](../../README.md) |
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
