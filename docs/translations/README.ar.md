<div dir="rtl" align="right">

# ⚡ GitIntel

### محرك استخبارات هندسة البرمجيات وملفات إنجاز المطورين عبر GitHub

**طبقة الأدلة القابلة للتحقق لمهندسي البرمجيات، مسؤولي التوظيف التقني، ووكلاء البرمجة بالذكاء الاصطناعي.**

<div align="center">

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](../../LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](../../CONTRIBUTING.md)

**[🌐 اللغات](#-إصدارات-اللغات)** — [English](../../README.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [中文](README.zh-CN.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [日本語](README.ja.md) · [Русский](README.ru.md) · [한국어](README.ko.md)

</div>

---

## 💡 المشكلة والحل

يعاني تقييم المطورين اليوم من طرفين معيبين:

- **مقاييس التفاخر الزائفة (Vanity Metrics)** — تتابع التعديلات المصطنع، التفرعات الفارغة (Forks)، والمستودعات الوهمية لتضخيم إحصائيات لغات البرمجة.
- **هلوسات نماذج الذكاء الاصطناعي** — بطاقات تقييم مجهولة المصدر ونصوص عامة خالية من أي أدلة ملموسة.

**GitIntel** يقدم المسار الثالث الموثوق: إشارات هندسية قطعية وقابلة للملاحظة بنسبة 100%، مستخرجة مباشرة من عناصر GitHub الحقيقية.

---

## 🚀 الميزات الرئيسية

### 1. 🌟 مولّد ملف التعريف الشخصي (GitHub Profile README)
حوّل صفحة `username/username` إلى واجهة تقنية معتمدة:
- إنشاء تلقائي لملف `README.md` بالاعتماد على بياناتك البرمجية الفعلية.
- إبراز ممارسات الاختبار، خطوط أنابيب CI/CD، وسجل الإصدارات.
- استبعاد التفرعات المهملة وعرض مشاريعك الأصلية الأكثر تأثيراً.

### 2. 🤖 خادم MCP أصلي (Model Context Protocol)
يمكّن مساعدي الذكاء الاصطناعي من فحص ملفات GitHub دون استنزاف حدود الطلبات (Rate Limit):
- أدوات JSON-RPC: `gitintel_analyze_profile` و `gitintel_get_signals` و `gitintel_generate_profile_readme` و `gitintel_get_ai_prompt`.
- متوافق تماماً مع **Claude Desktop** و **Cursor** و **Antigravity** وأي وكيل ذكاء اصطناعي يدعم بروتوكول MCP.

### 3. 📊 ست إشارات هندسية قابلة للتحقق

| الإشارة | ما يفحصه GitIntel بدقة | حدود النظام بوضوح |
| :--- | :--- | :--- |
| 🔄 **تطور الشفرة البرمجية** | حداثة الطوابع الزمنية للدفع (Push) والتحديث | لا يفحص حجم أو عدد الالتزامات (Commits) |
| 🛡️ **استمرارية الصيانة** | وتيرة التحديثات؛ تصنيف المستودعات المؤرشفة كـ "مكتملة" | عدم النشاط لا يعني بالضرورة سوء الصيانة |
| 🧪 **أدلة الاختبارات** | فحص شجرة Git لمجلدات الاختبار وتدفقات CI | لا يشغل الاختبارات ولا يقيس نسبة التغطية |
| 🤝 **التعاون البرمجي** | التحقق من نسب طلبات الدمج والمشاكل، والمساهمين | يتطلب تحليل المساهمات الخارجية واجهة GraphQL |
| 📖 **التوثيق الفني** | كشف ثلاثي لحالة التوثيق (README)، وGitHub Pages، والويكي، والترخيص | لا يقيّم جودة الصياغة اللغوية |
| 📦 **الإصدار والتسليم** | التحقق من إصدارات GitHub المعتمدة والوسوم الدلالية | يقتصر على المستودعات الأصلية النشطة |

### 4. 💼 سيرة ذاتية وملف إنجاز جاهز للتوظيف
- تصدير ملخصات بصيغة Markdown مدعومة بالأدلة للاستخدام في طلبات التوظيف والمقابلات.
- ميزة **"نسخ كمطالبة للذكاء الاصطناعي (AI Prompt)"** بضغطة زر لتوفير سياق موجز واقتصادي للرموز (Tokens).

---

## ⚡ البدء السريع

### المتطلبات الأساسية

| الأداة | الحد الأدنى للإصدار |
| :--- | :--- |
| Go | 1.26 |
| Node.js | 24 |
| npm | 11 |

### نظام ويندوز — تشغيل بضغطة واحدة
```bat
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel
run.bat
```

### أنظمة لينكس وماك
```bash
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel

# الطرفية الأولى — الواجهة الخلفية
cd backend
go run ./cmd/server

# الطرفية الثانية — الواجهة الأمامية
cd frontend
npm ci
npm run dev
```

### رمز وصول GitHub (اختياري)
```bash
cp .env.example .env
# قم بتعديل ملف .env وأضف: GITHUB_TOKEN=ghp_your_token_here
```

---

## 🤖 تكوين خادم MCP

### الربط مع Claude Desktop
أضف التالي إلى ملف `claude_desktop_config.json`:

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

## 🏗️ البنية المعمارية للنظام

```text
┌─────────────────────────────────────────┐
│      React 19 + Vite (TypeScript)       │
│         http://localhost:5173           │
└──────────────────┬──────────────────────┘
                   │ proxy /api
┌──────────────────▼──────────────────────┐
│          Go 1.26 Backend                │
│        http://localhost:8080            │
│  ذاكرة مؤقتة LRU (1,000 عنصر) +          │
│  عميل GitHub (ETag · Rate Guard)        │
└──────────────────┬──────────────────────┘
                   │ REST + ETag 304
        ┌──────────▼──────────┐
        │   GitHub REST API   │
        └─────────────────────┘
```

**أبرز المميزات الهندسية:**
- **ذاكرة مؤقتة منعدمة التخصيص الإضافي** — ذاكرة LRU تتسع لـ 1,000 إدخال مع لقطات بيانات غير قابلة للتغيير.
- **حماية استباقية من استنزاف الطلبات** — حجز هامش أمان يتسع لـ 10 طلبات لمنع الانقطاع المفاجئ.
- **طلبات ETag المشروطة** — استجابات HTTP 304 تستهلك **صفراً** من رصيد طلبات GitHub.
- **واجهة مستخدم خالية من تضارب الحالات** — إلغاء الطلبات المتداخلة عبر `AbortController`.
- **دعم كامل للاتجاه من اليمين لليسار (RTL)** — تجربة مستخدم سلسة باللغتين العربية والفارسية إلى جانب الإنجليزية.
- **أمان محلي حصري** — حصر مشاركة الموارد (CORS) بنطاق `127.0.0.1:5173`.

---

## 🧪 الاختبارات الآلية

```bash
# اختبارات الواجهة الخلفية
cd backend
go vet ./...
go test ./... -v -count=1

# فحص الواجهة الأمامية
cd frontend
npm run lint
npm run build
```

تتضمن حزمة اختبارات الواجهة الخلفية **109 اختبارات** تغطي 7 حزم برمجية.

---

## 🗺️ خارطة الطريق التقنية

للاطلاع على الخطة الشاملة، تفضل بزيارة [ENGINEERING_ROADMAP.md](../../ENGINEERING_ROADMAP.md).

| المسار | الحالة |
| :--- | :--- |
| المحرك الحسابي الأساسي وسلامة البيانات | ✅ مكتمل |
| الذاكرة المؤقتة LRU ودورة حياة بيانات الاعتماد | ✅ مكتمل |
| اعتمادية الواجهة الأمامية ودعم نمط RTL | ✅ مكتمل |
| معايير الأداء والتحسينات | ✅ مكتمل |
| خادم MCP الأصلي | ✅ الأساس مكتمل |
| مولّد ملفات Profile README | ✅ الأساس مكتمل |
| اختبارات التراجع المستندة لبيانات حقيقية | 🔄 جاري العمل (67%) |
| تعزيز الأمان والحماية | 🔄 جاري العمل (50%) |
| اختبارات الواجهة الأمامية باستخدام Vitest | ⬜ مخطط له |
| أتمتة المزامنة عبر GitHub Actions | ⬜ مخطط له |

---

## 🤝 المساهمة في المشروع

نرحب بجميع المساهمات المجتمعية! يرجى مراجعة [CONTRIBUTING.md](../../CONTRIBUTING.md) للتعرف على إرشادات التطوير.

---

## 🔒 الأمان والخصوصية

- يعمل GitIntel **محلياً بالكامل (localhost)**؛ لا تغادر بياناتك أو رموزك جهازك أبداً.
- قواعد CORS محصورة تماماً في العناوين المحلية.
- حماية مدمجة ضد هجمات حقن التعليمات (Prompt Injection) عند استخدام نماذج الذكاء الاصطناعي.

---

## 📄 الترخيص

هذا المشروع منشور تحت مظلة **رخصة GNU Affero General Public License v3.0 (GNU AGPL-3.0)** — انظر [LICENSE](../../LICENSE) لمزيد من التفاصيل.

---

## 🌐 إصدارات اللغات

| اللغة | الملف |
| :--- | :--- |
| 🇺🇸 English (النسخة المرجعية) | [README.md](../../README.md) |
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
