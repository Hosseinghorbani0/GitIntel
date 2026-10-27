<div align="center">

# ⚡ GitIntel

### Motor de Inteligência de Engenharia GitHub e Gerador de Portfólio

**A camada de evidência verificável para engenheiros de software, recrutadores técnicos e agentes de IA.**

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](../../LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](../../CONTRIBUTING.md)

**[🌐 Idioma](#-versões-de-idioma)** — [English](../../README.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [中文](README.zh-CN.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [日本語](README.ja.md) · [Русский](README.ru.md) · [한국어](README.ko.md)

</div>

---

## 💡 O Problema e a Solução

A avaliação de desenvolvedores hoje oscila entre dois extremos problemáticos:

- **Métricas de vaidade** — sequências artificiais de commits, forks vazios e repositórios de brinquedo que inflam estatísticas de linguagens.
- **Alucinações de IA** — scorecards de caixa preta e resumos genéricos sem nenhuma evidência verificável.

O **GitIntel** oferece o terceiro caminho: sinais 100% observáveis e determinísticos derivados diretamente de artefatos reais do GitHub.

---

## 🚀 Funcionalidades Principais

### 1. 🌟 Gerador de Profile README para GitHub
Transforme sua página `username/username` em uma vitrine de engenharia verificada:
- Gera automaticamente um `README.md` baseado em dados técnicos reais.
- Destaca práticas de testes, pipelines de CI/CD e histórico de releases.
- Filtra forks abandonados, priorizando seus projetos originais mais impactantes.

### 2. 🤖 Servidor MCP Nativo (Model Context Protocol)
Permite que assistentes de IA consultem perfis do GitHub sem esgotar sua cota de rate limit:
- Ferramentas JSON-RPC: `gitintel_analyze_profile`, `gitintel_get_signals`, `gitintel_generate_profile_readme`, `gitintel_get_ai_prompt`.
- Compatível com **Claude Desktop**, **Cursor**, **Antigravity** e qualquer cliente MCP.

### 3. 📊 Seis Sinais de Engenharia Verificáveis

| Sinal | O que o GitIntel inspeciona | Limitação honesta |
| :--- | :--- | :--- |
| 🔄 **Evolução de Código** | Recência dos timestamps de push e atualização | Não analisa volume bruto de commits |
| 🛡️ **Manutenção** | Frequência de atualizações; repositórios arquivados marcados como "Concluídos" | Inatividade não é necessariamente má manutenção |
| 🧪 **Evidência de Testes** | Varredura de árvore Git para diretórios de testes e workflows de CI | Não executa testes nem afere cobertura |
| 🤝 **Colaboração** | Proporção verificada de PRs/Issues, colaboradores e forks recebidos | Contribuições externas completas requerem GraphQL |
| 📖 **Documentação** | Detecção tri-estado de README, GitHub Pages, Wiki e LICENSE | Não avalia a qualidade da redação |
| 📦 **Entrega e Releases** | Releases verificados no GitHub e tags de versionamento semântico | Limitado a repositórios originais ativos |

### 4. 💼 Currículo e Portfólio Pronto para Recrutadores
- Exporte resumos em Markdown com base em evidências para candidaturas técnicas.
- Recurso **"Copiar como prompt de IA"** com um clique, otimizado para economia de tokens.

---

## ⚡ Início Rápido

### Pré-requisitos

| Ferramenta | Versão Mínima |
| :--- | :--- |
| Go | 1.26 |
| Node.js | 24 |
| npm | 11 |

### Windows — Inicialização com um clique
```bat
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel
run.bat
```

### Linux e macOS
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

### Token do GitHub (Opcional)
```bash
cp .env.example .env
# Edite o arquivo .env e defina: GITHUB_TOKEN=ghp_your_token_here
```

---

## 🤖 Integração MCP

### Conectar ao Claude Desktop
Adicione ao seu `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "gitintel": {
      "command": "go",
      "args": ["run", "./backend/cmd/mcp"],
      "cwd": "/caminho/para/GitIntel"
    }
  }
}
```

---

## 🏗️ Arquitetura do Sistema

```text
┌─────────────────────────────────────────┐
│      React 19 + Vite (TypeScript)       │
│         http://localhost:5173           │
└──────────────────┬──────────────────────┘
                   │ proxy /api
┌──────────────────▼──────────────────────┐
│          Go 1.26 Backend                │
│        http://localhost:8080            │
│  Cache LRU (1.000 entradas) +           │
│  Cliente GitHub (ETag · Rate Guard)     │
└──────────────────┬──────────────────────┘
                   │ REST + ETag 304
        ┌──────────▼──────────┐
        │   GitHub REST API   │
        └─────────────────────┘
```

**Destaques de Engenharia:**
- **Cache sem alocação extra** — LRU de 1.000 entradas com snapshots imutáveis.
- **Defesa ativa de limite de requisições** — Reserva buffer de segurança de 10 chamadas para evitar bloqueios.
- **Requisições condicionais ETag** — Respostas HTTP 304 consom **zero** da cota de requisições.
- **Interface imune a condições de corrida** — Cancelamento de requisições com `AbortController`.
- **Bilingue e acessível** — Suporte nativo completo a Persa (RTL) e Inglês.
- **Isolamento de segurança localhost** — CORS restrito estritamente a `127.0.0.1:5173`.

---

## 🧪 Testes

```bash
# Testes do Backend
cd backend
go vet ./...
go test ./... -v -count=1

# Verificações do Frontend
cd frontend
npm run lint
npm run build
```

A suíte de testes do backend inclui **109 testes** em 7 pacotes.

---

## 🗺️ Roadmap de Engenharia

Consulte o planejamento detalhado em [ENGINEERING_ROADMAP.md](../../ENGINEERING_ROADMAP.md).

| Área | Status |
| :--- | :--- |
| Motor de cálculo principal e integridade de dados | ✅ Concluído |
| Cache LRU e ciclo de vida de credenciais | ✅ Concluído |
| Confiabilidade do frontend e acessibilidade RTL | ✅ Concluído |
| Benchmarks de desempenho e otimização | ✅ Concluído |
| Servidor MCP nativo | ✅ Núcleo concluído |
| Gerador de Profile README | ✅ Núcleo concluído |
| Testes de regressão baseados em fixtures reais | 🔄 Em andamento (67%) |
| Hardening de segurança e auditoria | 🔄 Em andamento (50%) |
| Testes de frontend com Vitest | ⬜ Planejado |
| GitHub Action para sincronização automática | ⬜ Planejado |

---

## 🤝 Contribuições

Contribuições da comunidade são muito bem-vindas! Consulte o [CONTRIBUTING.md](../../CONTRIBUTING.md) para diretrizes e instruções de configuração.

---

## 🔒 Segurança e Privacidade

- O GitIntel opera **100% em localhost**. Seu token e código nunca saem do seu computador.
- Políticas CORS limitadas estritamente ao ambiente local.
- Proteções integradas contra injeção de prompt no módulo de IA.

---

## 📄 Licença

Distribuído sob a **Licença GNU Affero General Public License v3.0 (GNU AGPL-3.0)** — consulte o arquivo [LICENSE](../../LICENSE) para detalhes.

---

## 🌐 Versões de Idioma

| Idioma | Documento |
| :--- | :--- |
| 🇺🇸 English (canônico) | [README.md](../../README.md) |
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
