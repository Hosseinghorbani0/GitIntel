<div align="center">

# ⚡ GitIntel

### 결정론적 GitHub 엔지니어링 인텔리전스 & 포트폴리오 생성 엔진

**소프트웨어 엔지니어, 기술 채용 담당자, AI 코딩 에이전트를 위한 검증 가능한 객관적 증거 레이어.**

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](../../LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](../../CONTRIBUTING.md)

**[🌐 언어 선택](#-지원-언어)** — [English](../../README.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [中文](README.zh-CN.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [日本語](README.ja.md) · [Русский](README.ru.md) · [한국어](README.ko.md)

</div>

---

## 💡 문제의식과 해결책

오늘날 개발자 역량 평가는 두 가지 극단적인 왜곡에 갇혀 있습니다:

- **허영 지표 (Vanity Metrics)** — 조작된 잔디 심기(Commit Streak), 빈 포크(Fork), 언어 통계를 부풀리기 위한 일회성 토이 저장소.
- **거대언어모델(LLM)의 환각** — 검증 가능한 증거 없이 모호한 미사여구로 채워진 블랙박스 평가표.

**GitIntel**은 제3의 신뢰할 수 있는 대안을 제공합니다: 실제 GitHub 산출물에서 직접 도출된 100% 관측 가능하고 결정론적인 엔지니어링 시그널입니다.

---

## 🚀 주요 기능

### 1. 🌟 GitHub Profile README 자동 생성기
당신의 `username/username` 페이지를 객관적으로 입증된 엔지니어링 쇼케이스로 변환합니다:
- 실제 코드 및 개발 데이터에 기반하여 `README.md`를 자동으로 생성합니다.
- 테스트 작성 습관, CI/CD 파이프라인 구축 현황, 릴리스 히스토리를 강조합니다.
- 방치된 포크를 자동으로 필터링하고 가장 영향력 있는 오리지널 프로젝트를 우선 조명합니다.

### 2. 🤖 네이티브 MCP 서버 (Model Context Protocol)
AI 어시스턴트가 API 호출 한도 소진 없이 GitHub 프로필을 심층 분석할 수 있습니다:
- JSON-RPC 도구: `gitintel_analyze_profile`, `gitintel_get_signals`, `gitintel_generate_profile_readme`, `gitintel_get_ai_prompt`.
- **Claude Desktop**, **Cursor**, **Antigravity** 및 모든 표준 MCP 클라이언트와 완벽 호환.

### 3. 📊 6가지 검증 가능한 엔지니어링 시그널

| 시그널 | GitIntel이 실제로 검사하는 항목 | 솔직한 기술적 한계 |
| :--- | :--- | :--- |
| 🔄 **코드 진화 (Evolution)** | Push 및 Update 타임스탬프의 최신성 | 커밋의 절대적인 양은 측정하지 않음 |
| 🛡️ **유지보수성 (Maintenance)** | 업데이트 주기; 아카이브된 저장소는 '완료'로 분류 | 일시적 활동 공백이 관리 부실을 의미하지 않음 |
| 🧪 **테스트 증거 (Testing)** | 테스트 디렉토리 및 CI 워크플로우에 대한 Git 트리 스캔 | 테스트를 직접 실행하거나 커버리지를 측정하진 않음 |
| 🤝 **협업 지표 (Collaboration)** | 검증된 PR/Issue 비율, 기여자 수, 유입 포크 수 | 외부 기여 심층 분석에는 GraphQL 필요 |
| 📖 **문서화 (Documentation)** | README 3단계 상태 감지, GitHub Pages, Wiki, LICENSE | 작성된 문장의 문학적 품질까지 평가하진 않음 |
| 📦 **릴리스 및 배포 (Release)** | 검증된 GitHub Releases 및 시맨틱 버전 태그 | 활성 오리지널 저장소로 한정 |

### 4. 💼 채용을 위한 이력서 및 포트폴리오
- 면접과 입사 지원에 즉시 활용할 수 있는 증거 기반 Markdown 요약 내보내기.
- 토큰을 절약하는 **"AI 프롬프트로 복사"** 원클릭 지원.

---

## ⚡ 빠른 시작

### 요구 사양

| 도구 | 최소 권장 버전 |
| :--- | :--- |
| Go | 1.26 |
| Node.js | 24 |
| npm | 11 |

### Windows — 원클릭 실행
```bat
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel
run.bat
```

### Linux 및 macOS
```bash
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel

# 터미널 1 — 백엔드
cd backend
go run ./cmd/server

# 터미널 2 — 프론트엔드
cd frontend
npm ci
npm run dev
```

### GitHub 개인 액세스 토큰 (선택사항)
```bash
cp .env.example .env
# .env 파일을 열고 설정: GITHUB_TOKEN=ghp_your_token_here
```

---

## 🤖 MCP 연동 설정

### Claude Desktop에 연결하기
`claude_desktop_config.json` 파일에 다음 설정을 추가하세요:

```json
{
  "mcpServers": {
    "gitintel": {
      "command": "go",
      "args": ["run", "./backend/cmd/mcp"],
      "cwd": "/경로/GitIntel"
    }
  }
}
```

---

## 🏗️ 시스템 아키텍처

```text
┌─────────────────────────────────────────┐
│      React 19 + Vite (TypeScript)       │
│         http://localhost:5173           │
└──────────────────┬──────────────────────┘
                   │ proxy /api
┌──────────────────▼──────────────────────┐
│          Go 1.26 Backend                │
│        http://localhost:8080            │
│  LRU 캐시 (1,000 엔트리) +              │
│  GitHub 클라이언트 (ETag · Rate Guard)  │
└──────────────────┬──────────────────────┘
                   │ REST + ETag 304
        ┌──────────▼──────────┐
        │   GitHub REST API   │
        └─────────────────────┘
```

**엔지니어링 특징:**
- **추가 할당 없는 인메모리 캐시** — 불변 스냅샷을 갖춘 1,000개 슬롯 LRU 캐시.
- **선제적 호출 제한 방어** — 갑작스러운 차단을 방지하는 10회 요청 안전 버퍼.
- **ETag 조건부 요청** — HTTP 304 응답은 API 호출 한도를 **0회** 소모합니다.
- **경쟁 상태 없는 반응형 UI** — `AbortController`를 통한 이전 요청 자동 취소.
- **완벽한 다국어 및 접근성** — 페르시아어/아랍어(RTL)와 영어 완벽 지원.
- **로컬 보안 격리** — 백엔드 CORS는 오직 `127.0.0.1:5173`만 허용.

---

## 🧪 테스트 실행

```bash
# 백엔드 테스트
cd backend
go vet ./...
go test ./... -v -count=1

# 프론트엔드 정적 분석 및 빌드
cd frontend
npm run lint
npm run build
```

백엔드 테스트 스위트는 7개 패키지에 걸쳐 **109개의 테스트**를 통과합니다.

---

## 🗺️ 엔지니어링 로드맵

상세 계획은 [ENGINEERING_ROADMAP.md](../../ENGINEERING_ROADMAP.md)에서 확인하실 수 있습니다.

| 영역 | 상태 |
| :--- | :--- |
| 코어 연산 엔진 및 데이터 무결성 | ✅ 완료 |
| LRU 캐시 및 자격 증명 수명 주기 | ✅ 완료 |
| 프론트엔드 안정성 및 RTL 접근성 | ✅ 완료 |
| 성능 벤치마크 및 최적화 | ✅ 완료 |
| 네이티브 MCP 서버 | ✅ 코어 완료 |
| Profile README 생성기 | ✅ 코어 완료 |
| 실제 픽스처 기반 회귀 테스트 | 🔄 진행 중 (67%) |
| 보안 강화 및 감사 | 🔄 진행 중 (50%) |
| Vitest 프론트엔드 테스트 | ⬜ 예정됨 |
| GitHub Action 자동 동기화 워크플로우 | ⬜ 예정됨 |

---

## 🤝 기여하기

오픈소스 커뮤니티의 모든 기여를 환영합니다! 기여 가이드라인과 개발 환경 설정은 [CONTRIBUTING.md](../../CONTRIBUTING.md)를 참조하세요.

---

## 🔒 보안 및 개인정보 보호

- GitIntel은 **완전히 로컬(localhost)**에서 동작합니다. 토큰과 코드는 절대로 외부로 전송되지 않습니다.
- CORS 정책은 로컬 루프백으로 엄격히 제한됩니다.
- LLM 연동 시 프롬프트 인젝션 방어 레이어 및 소스 코드 보호 장치 내장.

---

## 📄 라이선스

이 프로젝트는 **GNU Affero General Public License v3.0 (GNU AGPL-3.0)** 하에 배포됩니다 — 자세한 내용은 [LICENSE](../../LICENSE) 파일을 참조하세요.

---

## 🌐 지원 언어

| 언어 | 문서 링크 |
| :--- | :--- |
| 🇺🇸 English (표준) | [README.md](../../README.md) |
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
