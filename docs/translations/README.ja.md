
<div align="center">

# ⚡ GitIntel

### 決定論的 GitHub エンジニアリング・インテリジェンス & ポートフォリオ生成エンジン

**ソフトウェアエンジニア、技術リクルーター、AIコーディングエージェントのための検証可能な客観的根拠レイヤー。**

[![Backend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml/badge.svg)](https://github.com/Hosseinghorbani0/GitIntel/actions/workflows/frontend.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](../../LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](../../CONTRIBUTING.md)

**[🌐 言語一覧](#-言語バージョン)** — [English](../../README.md) · [فارسی](README.fa.md) · [Español](README.es.md) · [中文](README.zh-CN.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [日本語](README.ja.md) · [Русский](README.ru.md) · [한국어](README.ko.md)

</div>

---

## 💡 課題とアプローチ

開発者のスキル評価は現在、2つの破綻した極端なアプローチに陥っています：

- **虚飾のメトリクス（Vanity Metrics）** — 作成されたコミット連続記録、空のフォーク、言語統計を水増しするためのおもちゃのリポジトリ。
- **LLMのハルシネーション** — 検証可能な証拠のないブラックボックススコアカードや根拠のない褒め言葉。

**GitIntel** は第三の選択肢を提供します：実際のGitHubアーティファクトから直接導出される、100%観察可能で決定論的なエンジニアリングシグナルです。

---

## 🚀 主な機能

### 1. 🌟 GitHub Profile README ジェネレーター
あなたの `username/username` ページを客観的に検証された技術ショーケースに変換：
- 実際の開発データに基づき `README.md` を自動生成。
- テスト実施の証拠、CI/CDパイプライン、リリース履歴を強調。
- 放置されたフォークを除外し、最も影響力のあるオリジナル開発を前面に表示。

### 2. 🤖 ネイティブ MCP サーバー (Model Context Protocol)
APIレート制限を無駄遣いすることなく、AIアシスタントがGitHubプロファイルを調査可能：
- JSON-RPCツール：`gitintel_analyze_profile`, `gitintel_get_signals`, `gitintel_generate_profile_readme`, `gitintel_get_ai_prompt`。
- **Claude Desktop**, **Cursor**, **Antigravity** および各種MCPクライアントと完全互換。

### 3. 📊 6つの検証可能なエンジニアリングシグナル

| シグナル | GitIntelが実際に検査する内容 | 誠実な技術的制限事項 |
| :--- | :--- | :--- |
| 🔄 **コードの継続進化** | PushおよびUpdateタイムスタンプの新しさ | コミット数の絶対量は集計しません |
| 🛡️ **メンテナンス品質** | 更新頻度；アーカイブ済みリポジトリを「完了」と分類 | 一時的な休止＝保守不良ではありません |
| 🧪 **テストの実践証拠** | テストディレクトリとCIワークフローのGitツリースキャン | テスト実行やカバレッジ測定は行いません |
| 🤝 **協調開発の実績** | PR/Issue比率、コントリビューター数、着信フォークの検証 | 外部貢献の詳細分析にはGraphQLが必要です |
| 📖 **技術ドキュメント** | README三状態検出、GitHub Pages、Wiki、LICENSE | 文章自体の文学的品質は評価しません |
| 📦 **リリースとデリバリー** | 検証済みGitHubリリースおよびセマンティックバージョニング | アクティブなオリジナルリポジトリに限定 |

### 4. 💼 採用担当者向け職務経歴・ポートフォリオ
- 技術面接や応募書類ですぐに使える、証拠に裏付けられたMarkdownサマリーを出力。
- トークン効率に優れた **「AIプロンプトとしてコピー」** 機能をワンクリックで提供。

---

## ⚡ クイックスタート

### 前提条件

| ツール | 推奨最小バージョン |
| :--- | :--- |
| Go | 1.26 |
| Node.js | 24 |
| npm | 11 |

### Windows — ワンクリック起動
```bat
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel
run.bat
```

### Linux & macOS
```bash
git clone https://github.com/Hosseinghorbani0/GitIntel.git
cd GitIntel

# ターミナル 1 — バックエンド
cd backend
go run ./cmd/server

# ターミナル 2 — フロントエンド
cd frontend
npm ci
npm run dev
```

### GitHub アクセストークン（任意）
```bash
cp .env.example .env
# .env を編集して設定: GITHUB_TOKEN=ghp_your_token_here
```

---

## 🤖 MCP 統合設定

### Claude Desktop との接続
`claude_desktop_config.json` に以下を追記してください：

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

## 🏗️ システムアーキテクチャ

```text
┌─────────────────────────────────────────┐
│      React 19 + Vite (TypeScript)       │
│         http://localhost:5173           │
└──────────────────┬──────────────────────┘
                   │ proxy /api
┌──────────────────▼──────────────────────┐
│          Go 1.26 Backend                │
│        http://localhost:8080            │
│  LRU キャッシュ (1,000 エントリ) +       │
│  GitHub クライアント (ETag · Rate Guard) │
└──────────────────┬──────────────────────┘
                   │ REST + ETag 304
        ┌──────────▼──────────┐
        │   GitHub REST API   │
        └─────────────────────┘
```

**エンジニアリングハイライト：**
- **ゼロアロケーションキャッシュ** — 不変スナップショットを備えた1,000エントリLRU。
- **プロアクティブなレート制限防御** — レート超過を防ぐ10リクエストの安全バッファ。
- **ETag条件付きリクエスト** — HTTP 304 応答はレート制限枠を **一切消費しません**。
- **競合状態を防ぐUI** — `AbortController` による非同期リクエスト自動キャンセル。
- **バイリンガル & アクセシビリティ** — ペルシア語/アラビア語（RTL）と英語の完全サポート。
- **ローカル限定セキュリティ** — バックエンドCORSは `127.0.0.1:5173` のみに制限。

---

## 🧪 テストの実行

```bash
# バックエンドテスト
cd backend
go vet ./...
go test ./... -v -count=1

# フロントエンド検証
cd frontend
npm run lint
npm run build
```

バックエンドテストスイートは7パッケージにわたり **109件のテスト** を網羅しています。

---

## 🗺️ ロードマップ

詳細なマイルストーンは [ENGINEERING_ROADMAP.md](../../ENGINEERING_ROADMAP.md) をご覧ください。

| 領域 | ステータス |
| :--- | :--- |
| コア計算エンジンとデータ完全性 | ✅ 完了 |
| LRUキャッシュと認証情報ライフサイクル | ✅ 完了 |
| フロントエンド堅牢性とRTLアクセシビリティ | ✅ 完了 |
| パフォーマンスベンチマークと最適化 | ✅ 完了 |
| ネイティブMCPサーバー | ✅ コア完了 |
| Profile README ジェネレーター | ✅ コア完了 |
| 実データを用いた回帰テストスイート | 🔄 進行中 (67%) |
| セキュリティ監査と強化 | 🔄 進行中 (50%) |
| Vitest によるフロントエンド単体テスト | ⬜ 計画中 |
| 自動同期のための GitHub Action | ⬜ 計画中 |

---

## 🤝 コントリビューション

コミュニティからの改善提案やバグ報告を歓迎します。設定手順については [CONTRIBUTING.md](../../CONTRIBUTING.md) をご確認ください。

---

## 🔒 セキュリティとプライバシー

- GitIntel は **ローカル（localhost）環境で完全に動作** します。トークンやコードが外部に送信されることはありません。
- CORS ポリシーはローカルループバックに厳密に制限されています。
- LLM統合レイヤーにおけるプロンプトインジェクション防御とソースコード保護。

---

## 📄 ライセンス

本プロジェクトは **GNU Affero General Public License v3.0 (GNU AGPL-3.0)** の下で公開されています — 詳細は [LICENSE](../../LICENSE) をご覧ください。

---

## 🌐 言語バージョン

| 言語 | ドキュメント |
| :--- | :--- |
| 🇺🇸 English (正本) | [README.md](../../README.md) |
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
