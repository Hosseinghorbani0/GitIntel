<div align="center">

# ⚡ GitIntel

### Детерминированный движок инженерной аналитики GitHub и генератор портфолио

**Слой проверяемых фактов для инженеров-программистов, технических рекрутеров и AI-агентов.**

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](../../LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](../../CONTRIBUTING.md)

**[🌐 Языковые версии](#-языковые-версии)** — [English](../../README.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [中文](README.zh-CN.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [日本語](README.ja.md) · [Русский](README.ru.md) · [한국어](README.ko.md)

</div>

---

## 💡 Проблема и решение

Оценка разработчиков сегодня застряла между двумя крайностями:

- **Метрики тщеславия (Vanity Metrics)** — накрученные серии коммитов, пустые форки и игрушечные репозитории для раздувания языковой статистики.
- **Галлюцинации языковых моделей** — непрозрачные оценочные карточки («черный ящик») и общие комплименты без проверяемых доказательств.

**GitIntel** предлагает надежный третий путь: на 100% наблюдаемые детерминированные инженерные сигналы, вычисляемые напрямую из артефактов GitHub.

---

## 🚀 Основные возможности

### 1. 🌟 Генератор Profile README для GitHub
Превратите вашу страницу `username/username` в подтвержденную техническую витрину:
- Автоматическая генерация `README.md` на основе реальных инженерных фактов.
- Демонстрация культуры тестирования, CI/CD-пайплайнов и истории релизов.
- Фильтрация заброшенных форков; фокус на ваших самых значимых оригинальных проектах.

### 2. 🤖 Нативный MCP-сервер (Model Context Protocol)
Позволяет AI-ассистентам анализировать GitHub-профили без исчерпания лимитов API:
- Инструменты JSON-RPC: `gitintel_analyze_profile`, `gitintel_get_signals`, `gitintel_generate_profile_readme`, `gitintel_get_ai_prompt`.
- Полная совместимость с **Claude Desktop**, **Cursor**, **Antigravity** и любыми MCP-клиентами.

### 3. 📊 Шесть проверяемых инженерных сигналов

| Сигнал | Что именно проверяет GitIntel | Честное техническое ограничение |
| :--- | :--- | :--- |
| 🔄 **Эволюция кода** | Свежесть временных меток push и update | Не подсчитывает общий объем коммитов |
| 🛡️ **Поддержка (Maintenance)** | Регулярность обновлений; архивные репозитории отмечаются «Завершено» | Период неактивности ≠ плохая поддержка |
| 🧪 **Факты тестирования** | Сканирование дерева Git на папки тестов и CI-воркфлоу | Не запускает тесты и не измеряет покрытие кода |
| 🤝 **Коллаборация** | Соотношение PR/Issue, число контрибьюторов, входящие форки | Для глубокого анализа внешних вкладов нужен GraphQL |
| 📖 **Документация** | Трехпозиционная проверка README, GitHub Pages, Wiki и LICENSE | Не оценивает литературное качество текста |
| 📦 **Релизы и поставка** | Проверенные релизы GitHub и теги семантического версионирования | Ограничено активными оригинальными проектами |

### 4. 💼 Резюме и портфолио для рекрутеров
- Экспорт подтвержденных фактами сводок в Markdown для технических собеседований.
- Функция **«Копировать как AI-промпт»** в один клик для экономии токенов в контексте нейросетей.

---

## ⚡ Быстрый старт

### Требования

| Инструмент | Минимальная версия |
| :--- | :--- |
| Go | 1.26 |
| Node.js | 24 |
| npm | 11 |

### Windows — запуск в один клик
```bat
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel
run.bat
```

### Linux и macOS
```bash
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel

# Терминал 1 — Бэкенд
cd backend
go run ./cmd/server

# Терминал 2 — Фронтенд
cd frontend
npm ci
npm run dev
```

### Токен GitHub (опционально)
```bash
cp .env.example .env
# Отредактируйте .env: GITHUB_TOKEN=ghp_your_token_here
```

---

## 🤖 Интеграция с MCP

### Подключение к Claude Desktop
Добавьте в конфигурацию `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "gitintel": {
      "command": "go",
      "args": ["run", "./backend/cmd/mcp"],
      "cwd": "/путь/к/GitIntel"
    }
  }
}
```

---

## 🏗️ Архитектура системы

```text
┌─────────────────────────────────────────┐
│      React 19 + Vite (TypeScript)       │
│         http://localhost:5173           │
└──────────────────┬──────────────────────┘
                   │ proxy /api
┌──────────────────▼──────────────────────┐
│          Go 1.26 Backend                │
│        http://localhost:8080            │
│  LRU-кэш (1 000 записей) +              │
│  GitHub-клиент (ETag · Rate Guard)      │
└──────────────────┬──────────────────────┘
                   │ REST + ETag 304
        ┌──────────▼──────────┐
        │   GitHub REST API   │
        └─────────────────────┘
```

**Инженерные решения:**
- **Кэш без лишних аллокаций** — LRU-кэш на 1 000 элементов с неизменяемыми снимками.
- **Проактивная защита от исчерпания лимитов** — буфер безопасности на 10 запросов.
- **Условные запросы по ETag** — ответы HTTP 304 тратят **ноль** квот API.
- **UI без race conditions** — отмена устаревших запросов через `AbortController`.
- **Двуязычность и доступность** — полная поддержка RTL (персидский, арабский) и английского.
- **Локальная изоляция** — CORS бэкенда строго ограничен `127.0.0.1:5173`.

---

## 🧪 Тестирование

```bash
# Тесты бэкенда
cd backend
go vet ./...
go test ./... -v -count=1

# Проверка фронтенда
cd frontend
npm run lint
npm run build
```

Тестовый набор бэкенда насчитывает **109 тестов** в 7 пакетах.

---

## 🗺️ План развития (Roadmap)

Подробный план доступен в документе [ENGINEERING_ROADMAP.md](../../ENGINEERING_ROADMAP.md).

| Направление | Статус |
| :--- | :--- |
| Вычислительное ядро и целостность данных | ✅ Завершено |
| LRU-кэширование и жизненный цикл ключей | ✅ Завершено |
| Надежность фронтенда и доступность RTL | ✅ Завершено |
| Бенчмарки производительности и оптимизация | ✅ Завершено |
| Нативный MCP-сервер | ✅ Ядро готово |
| Генератор Profile README | ✅ Ядро готово |
| Регрессионные тесты на реальных данных | 🔄 В процессе (67%) |
| Усиление безопасности и аудит | 🔄 В процессе (50%) |
| Фронтенд-тестирование на Vitest | ⬜ Запланировано |
| GitHub Action для автоматической синхронизации | ⬜ Запланировано |

---

## 🤝 Участие в разработке

Мы рады любому вкладу сообщества! Пожалуйста, ознакомьтесь с [CONTRIBUTING.md](../../CONTRIBUTING.md).

---

## 🔒 Безопасность и конфиденциальность

- GitIntel работает **полностью локально (localhost)**. Токены и код никогда не передаются на внешние серверы.
- Правила CORS привязаны строго к локальному окружению.
- Встроенные механизмы защиты от Prompt Injection при взаимодействии с AI.

---

## 📄 Лицензия

Проект распространяется под лицензией **GNU Affero General Public License v3.0 (GNU AGPL-3.0)** — подробности в файле [LICENSE](../../LICENSE).

---

## 🌐 Языковые версии

| Язык | Документ |
| :--- | :--- |
| 🇺🇸 English (канонический) | [README.md](../../README.md) |
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
