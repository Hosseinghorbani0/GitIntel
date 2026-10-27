







<div dir="rtl" align="right">

# ⚡ GitIntel

### موتور اطلاعات مهندسی و نمونه‌کار توسعه‌دهنده بر پایه GitHub

**لایه شواهد قابل راستی‌آزمایی برای مهندسان نرم‌افزار، استخدام‌کنندگان فنی و عامل‌های هوش مصنوعی.**

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](../../LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](../../CONTRIBUTING.md)

**[🌐 زبان‌ها](#-نسخه‌های-زبانی)** — [English](../../README.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [中文](README.zh-CN.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [日本語](README.ja.md) · [Русский](README.ru.md) · [한국어](README.ko.md)

---

## 💡 مشکل

ارزیابی توسعه‌دهنده بین دو افراط شکسته است:

```text
┌───────────────────────────────────┐       ┌───────────────────────────────────┐
│         معیارهای بی‌محتوا        │       │          هوش مصنوعی توهم‌زا       │
│  • دستکاری streak کامیت          │  vs.  │  • توصیف‌های پر از اغراق          │
│  • fork های خالی                 │       │  • نمرات مبهم و بی‌دلیل           │
│  • ریپوهای اسباب‌بازی تورم می‌دن │       │  • هیچ شواهد قابل تأیید           │
└───────────────────────────────────┘       └───────────────────────────────────┘
                        ↓
┌───────────────────────────────────────────────────────┐
│                       GitIntel                        │
│  ✅ سیگنال‌های ۱۰۰٪ قابل مشاهده و قطعی               │
│  ✅ توزیع زبان بر اساس حجم بایت واقعی                │
│  ✅ CI/CD، درخت‌های تست و تگ‌های ریلیز تأیید شده      │
│  ✅ لایه شواهد برای عامل‌های هوش مصنوعی               │
└───────────────────────────────────────────────────────┘
```

---

## 🚀 قابلیت‌ها

### ۱. 🌟 تولیدکننده Profile README برای GitHub
صفحه `username/username` شما را به نمایشگاه مهندسی تبدیل کنید:
- تولید خودکار `README.md` بر پایه داده‌های واقعی فنی.
- برجسته‌سازی تست، CI/CD و تاریخچه ریلیز.
- حذف fork های رها شده؛ نمایش کارهای اصلی و پرتأثیر.

### ۲. 🤖 سرور MCP بومی (پروتکل زمینه مدل)
به دستیاران هوش مصنوعی اجازه می‌دهد پروفایل GitHub را بدون مصرف rate limit بررسی کنند:
- ابزارهای JSON-RPC: `gitintel_analyze_profile`، `gitintel_get_signals`، `gitintel_generate_profile_readme`، `gitintel_get_ai_prompt`.
- سازگار با **Claude Desktop**، **Cursor**، **Antigravity** و هر عامل MCP.

### ۳. 📊 شش سیگنال مهندسی قابل تأیید

| سیگنال | چه چیزی بررسی می‌شود | محدودیت |
| :--- | :--- | :--- |
| 🔄 **تکامل کد** | تازگی timestamp ها | حجم commit بررسی نمی‌شود |
| 🛡️ **نگهداری** | فرکانس به‌روزرسانی | بی‌تحرکی ≠ نگهداری ضعیف |
| 🧪 **شواهد تست** | اسکن درخت Git برای پوشه‌های تست | تست‌ها اجرا نمی‌شوند |
| 🤝 **همکاری** | PR/issue جداگانه و تعداد مشارکت‌کنندگان | مشارکت‌های خارجی نیاز به GraphQL دارد |
| 📖 **مستندسازی** | README، Pages، Wiki، LICENSE | کیفیت متن بررسی نمی‌شود |
| 📦 **ریلیز و تحویل** | ریلیزهای GitHub تأیید شده | محدود به ریپوهای اصلی |

### ۴. 💼 رزومه و نمونه‌کار آماده برای استخدام
- خروجی Markdown با شواهد برای درخواست‌های شغلی.
- یک کلیک **"کپی به عنوان پرامپت هوش مصنوعی"**.

---

## ⚡ شروع سریع

### پیش‌نیازها

| ابزار | حداقل نسخه |
| :--- | :--- |
| Go | 1.26 |
| Node.js | 24 |
| npm | 11 |

### ویندوز — اجرای یک‌کلیکه
```bat
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel
run.bat
```

### لینوکس و macOS
```bash
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel

# ترمینال ۱ — Backend
cd backend
go run ./cmd/server

# ترمینال ۲ — Frontend
cd frontend
npm ci
npm run dev
```

### توکن اختیاری GitHub
```bash
cp .env.example .env
# ویرایش .env و تنظیم: GITHUB_TOKEN=ghp_your_token_here
```

---

## 🤖 یکپارچه‌سازی MCP

### اتصال به Claude Desktop
در `claude_desktop_config.json` اضافه کنید:

```json
{
  "mcpServers": {
    "gitintel": {
      "command": "go",
      "args": ["run", "./backend/cmd/mcp"],
      "cwd": "/path/to/GitIntel"
    }
  }
}
```

---

## 🏗️ معماری

```text
┌─────────────────────────────────────────┐
│      React 19 + Vite (TypeScript)       │
│   Vazirmatn RTL/LTR · Lucide React     │
│         http://localhost:5173           │
└──────────────────┬──────────────────────┘
                   │ پروکسی /api
┌──────────────────▼──────────────────────┐
│          Go 1.26 Backend                │
│        http://localhost:8080            │
│                                         │
│  کش LRU (۱۰۰۰ ورودی) + کلاینت GitHub  │
│  (ETag · محافظت نرخ درخواست)           │
└──────────────────┬──────────────────────┘
                   │ REST + ETag 304
        ┌──────────▼──────────┐
        │   GitHub REST API   │
        └─────────────────────┘
```

**نکات مهندسی:**
- **کش بدون تخصیص** — LRU با ۱۰۰۰ ورودی، اسنپ‌شات‌های immutable بدون سربار JSON.
- **دفاع فعال در برابر rate limit** — ذخیره ۱۰ درخواست به عنوان بافر ایمنی.
- **درخواست‌های شرطی ETag** — HTTP 304 هیچ سهمیه‌ای مصرف نمی‌کند.
- **رابط کاربری ایمن از race condition** — لغو درخواست با `AbortController`.
- **دوزبانه و قابل دسترس** — پشتیبانی کامل از فارسی با Vazirmatn و چیدمان RTL.
- **CORS محدود به localhost** — Backend فقط درخواست از `127.0.0.1:5173` قبول می‌کند.

---

## 🧪 تست

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

مجموعه تست Backend شامل **۱۰۹ تست** در ۷ پکیج است.

---

## 🗺️ نقشه راه

جزئیات کامل در [ENGINEERING_ROADMAP.md](../../ENGINEERING_ROADMAP.md) موجود است.

| حوزه | وضعیت |
| :--- | :--- |
| موتور پایه و یکپارچگی داده | ✅ کامل |
| کش LRU و چرخه اعتبارنامه | ✅ کامل |
| قابلیت اطمینان frontend و RTL | ✅ کامل |
| عملکرد و معیارسنجی | ✅ کامل |
| سرور MCP بومی | ✅ هسته کامل |
| تولیدکننده Profile README | ✅ هسته کامل |
| تست‌های regression مبتنی بر fixture | 🔄 ۶۷٪ |
| سخت‌سازی امنیتی | 🔄 ۵۰٪ |
| تست‌های frontend با Vitest | ⬜ برنامه‌ریزی شده |
| GitHub Action برای همگام‌سازی | ⬜ برنامه‌ریزی شده |

---

## 🤝 مشارکت

مشارکت‌های توسعه‌دهندگان از سرتاسر جهان خوش‌آمد است. لطفاً [CONTRIBUTING.md](../../CONTRIBUTING.md) را مطالعه کنید.

**حوزه‌های با تأثیر بالا:**
- تست‌های frontend (`GI-TEST-003`) — Vitest + React Testing Library
- GitHub Action (`GI-AI-003`) — کرون هفتگی برای بازسازی Profile README
- استخراج GraphQL (`GI-PROD-001`) — شناسایی مشارکت‌ها در سازمان‌های خارجی

---

## 🔒 امنیت و حریم خصوصی

- GitIntel **کاملاً روی localhost** اجرا می‌شود. توکن و داده‌های تحلیل هرگز دستگاه شما را ترک نمی‌کنند.
- CORS فقط به `127.0.0.1:5173` محدود است.
- دفاع‌های تزریق prompt در یکپارچه‌سازی LLM تعبیه شده‌اند.

---

## 📄 مجوز

تحت **مجوز GNU Affero General Public License v3.0 (GNU AGPL-3.0)** توزیع می‌شود — برای جزئیات فایل [LICENSE](../../LICENSE) را ببینید.

---

## 🌐 نسخه‌های زبانی

| زبان | فایل |
| :--- | :--- |
| 🇺🇸 English (اصلی) | [README.md](../../README.md) |
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

</div>
