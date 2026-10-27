<div align="center">

# ⚡ GitIntel

### Moteur d'Intelligence d'Ingénierie GitHub et Générateur de Portfolio

**La couche de preuves vérifiables pour les développeurs, recruteurs techniques et agents IA.**

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](../../LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](../../CONTRIBUTING.md)

**[🌐 Langue](#-versions-linguistiques)** — [English](../../README.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [中文](README.zh-CN.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [日本語](README.ja.md) · [Русский](README.ru.md) · [한국어](README.ko.md)

</div>

---

## 💡 Le Problème et la Solution

L'évaluation des développeurs est aujourd'hui prise en étau entre deux extrêmes défaillants :

- **Métriques de vanité** — séquences de commits gonflées artificiellement, forks vides et dépôts jouets pour fausser les statistiques de langages.
- **Hallucinations des LLM** — cartes de score opaques et résumés flatteurs dénués de toute preuve vérifiable.

**GitIntel** offre la troisième voie : des signaux d'ingénierie déterministes et 100% observables, extraits directement des artefacts GitHub réels.

---

## 🚀 Fonctionnalités Clés

### 1. 🌟 Générateur de Profile README GitHub
Transformez votre page `username/username` en une vitrine d'ingénierie authentique :
- Génère automatiquement un `README.md` fondé sur des faits techniques mesurables.
- Met en avant la rigueur des tests, les pipelines CI/CD et l'historique des releases.
- Filtre les forks inactifs pour valoriser vos projets originaux à fort impact.

### 2. 🤖 Serveur MCP Natif (Model Context Protocol)
Permet aux assistants IA d'analyser vos profils GitHub sans saturer vos quotas d'API :
- Outils JSON-RPC : `gitintel_analyze_profile`, `gitintel_get_signals`, `gitintel_generate_profile_readme`, `gitintel_get_ai_prompt`.
- Compatible avec **Claude Desktop**, **Cursor**, **Antigravity** et tout agent supportant MCP.

### 3. 📊 Six Signaux d'Ingénierie Vérifiables

| Signal | Ce que GitIntel inspecte réellement | Limite honnête du système |
| :--- | :--- | :--- |
| 🔄 **Évolution du Code** | Fraîcheur des horodatages de push et de mise à jour | N'analyse pas le volume brut de commits |
| 🛡️ **Maintenance** | Fréquence des commits ; dépôts archivés classés "Achevés" | L'inactivité ponctuelle ne signifie pas un abandon |
| 🧪 **Preuves de Tests** | Scan de l'arborescence Git pour les dossiers de test et workflows CI | N'exécute pas les tests et ne mesure pas la couverture |
| 🤝 **Collaboration** | Ratio vérifié PR/Issues, contributeurs, forks entrants | L'analyse complète des contributions externes requiert GraphQL |
| 📖 **Documentation** | Détection tri-états du README, GitHub Pages, Wiki et LICENSE | N'évalue pas la qualité rédactionnelle |
| 📦 **Livraison & Releases** | Releases GitHub vérifiées et étiquettes de versions sémantiques | Limité aux dépôts originaux actifs |

### 4. 💼 CV et Portfolio Prêts pour le Recrutement
- Exportez des résumés Markdown étayés par des faits pour vos candidatures techniques.
- Fonction **« Copier comme prompt IA »** en un clic pour un contexte concis et économe en tokens.

---

## ⚡ Démarrage Rapide

### Prérequis

| Outil | Version Minimale |
| :--- | :--- |
| Go | 1.26 |
| Node.js | 24 |
| npm | 11 |

### Windows — Démarrage en un clic
```bat
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel
run.bat
```

### Linux et macOS
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

### Jeton d'accès GitHub (Optionnel)
```bash
cp .env.example .env
# Modifiez .env et renseignez : GITHUB_TOKEN=ghp_your_token_here
```

---

## 🤖 Intégration MCP

### Configuration avec Claude Desktop
Ajoutez à votre fichier `claude_desktop_config.json` :

```json
{
  "mcpServers": {
    "gitintel": {
      "command": "go",
      "args": ["run", "./backend/cmd/mcp"],
      "cwd": "/chemin/vers/GitIntel"
    }
  }
}
```

---

## 🏗️ Architecture du Système

```text
┌─────────────────────────────────────────┐
│      React 19 + Vite (TypeScript)       │
│         http://localhost:5173           │
└──────────────────┬──────────────────────┘
                   │ proxy /api
┌──────────────────▼──────────────────────┐
│          Go 1.26 Backend                │
│        http://localhost:8080            │
│  Cache LRU (1 000 entrées) +            │
│  Client GitHub (ETag · Rate Guard)      │
└──────────────────┬──────────────────────┘
                   │ REST + ETag 304
        ┌──────────▼──────────┐
        │   GitHub REST API   │
        └─────────────────────┘
```

**Points Forts d'Ingénierie :**
- **Cache sans allocations superflues** — LRU de 1 000 éléments avec snapshots immuables.
- **Protection proactive du Rate Limit** — Réserve une marge de sécurité de 10 requêtes.
- **Requêtes conditionnelles ETag** — Les réponses HTTP 304 ne consomment **aucun** quota d'API.
- **Interface immunisée contre les race conditions** — Annulation des requêtes via `AbortController`.
- **Bilingue et accessible** — Support nativo de l'arabe/persan (RTL) et de l'anglais.
- **Sécurité localhost étanche** — CORS strictement limité à `127.0.0.1:5173`.

---

## 🧪 Tests et Validation

```bash
# Tests Backend
cd backend
go vet ./...
go test ./... -v -count=1

# Validation Frontend
cd frontend
npm run lint
npm run build
```

La suite de tests backend totalise **109 tests** sur 7 packages.

---

## 🗺️ Feuille de Route Technique

Consultez l'ensemble des détails dans [ENGINEERING_ROADMAP.md](../../ENGINEERING_ROADMAP.md).

| Domaine | Statut |
| :--- | :--- |
| Moteur de calcul principal et intégrité des données | ✅ Achevé |
| Cache LRU et cycle de vie des identifiants | ✅ Achevé |
| Fiabilité frontend et accessibilité RTL | ✅ Achevé |
| Benchmarks de performance et optimisations | ✅ Achevé |
| Serveur MCP natif | ✅ Cœur achevé |
| Générateur de Profile README | ✅ Cœur achevé |
| Tests de régression sur données réelles | 🔄 En cours (67%) |
| Durcissement de sécurité et audits | 🔄 En cours (50%) |
| Tests frontend avec Vitest | ⬜ Planifié |
| GitHub Action pour la synchronisation automatique | ⬜ Planifié |

---

## 🤝 Contribuer

Toute contribution est la bienvenue ! Consultez [CONTRIBUTING.md](../../CONTRIBUTING.md) pour les consignes et la mise en route.

---

## 🔒 Sécurité et Confidentialité

- GitIntel s'exécute **intégralement sur votre machine (localhost)**. Vos tokens et vos données ne sont jamais transmis à l'extérieur.
- Politiques CORS restreintes à l'environnement local.
- Protections intégrées contre l'injection de prompts dans le module LLM.

---

## 📄 Licence

Distribué sous **Licence GNU Affero General Public License v3.0 (GNU AGPL-3.0)** — consultez le fichier [LICENSE](../../LICENSE) pour plus d'informations.

---

## 🌐 Versions Linguistiques

| Langue | Document |
| :--- | :--- |
| 🇺🇸 English (canonique) | [README.md](../../README.md) |
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
