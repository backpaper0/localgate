# Research & Design Decisions

---
**Purpose**: Capture discovery findings, architectural investigations, and rationale that inform the technical design.

---

## Summary

- **Feature**: `portal-screen`
- **Discovery Scope**: Extension（既存管理APIへのUI追加）
- **Key Findings**:
  - `internal/server/server.go` は「自己ホスト名 + サブドメインなし」のリクエストをすべて `management.API.ServeHTTP()` へ委譲しており、`GET /` は未定義のため現在404が返る
  - `internal/management/api.go` の `http.ServeMux` に `GET /` を追加するだけで最小限の変更でポータルを配信できる
  - Go 1.16+ の `embed` パッケージでHTMLをバイナリに静的埋め込みでき、外部ファイルへの依存をなくせる

## Research Log

### 既存ルーティング構造の分析

- **Context**: ポータル配信のどのレイヤに手を入れるべきか確認
- **Sources Consulted**: `internal/server/server.go`, `internal/management/api.go`
- **Findings**:
  - `ServeHTTP` → `selfHostnames` チェック → `management.ServeHTTP()` の順でルーティング
  - management の `mux` は `POST /services`, `DELETE /services/{name}`, `GET /services` のみ登録済み
  - `GET /` は未登録 → ServeMux のデフォルト動作（404）が返る
- **Implications**: management の mux に `GET /` を追加することで、既存ルーティングを変更せずにポータルを挿入できる

### HTML 埋め込み方式の選定

- **Context**: Go バイナリ単体で動作する要件（設定ファイル不要）を維持するための HTML 配信方式を検討
- **Findings**:
  - `//go:embed` ディレクティブ（Go 1.16+）で HTML ファイルをバイナリに埋め込める
  - ハードコードされた文字列リテラルよりも HTML ファイルとして分離したほうがメンテナンス性が高い
  - Go 1.22 環境では利用可能
- **Implications**: `internal/portal/portal.html` を `//go:embed portal.html` でバイナリに含める

### リアルタイム更新方式の選定

- **Context**: 「ある程度リアルタイム」な状態反映をどう実現するか
- **Findings**:
  - **Polling**: `setInterval` でAPIを定期呼び出し。サーバ側変更なし。実装コスト最小。
  - **SSE (Server-Sent Events)**: サーバからプッシュ。精度高いが、サーバ側に新しいエンドポイントが必要。
  - **WebSocket**: オーバーキル。
- **Implications**: プロジェクトの「シンプルさを優先」方針に従い Polling（5秒間隔）を採用。SSEへの移行は将来の選択肢として保持。

### CSS フレームワーク選定

- **Context**: 「モダンでかわいい」UIをGo標準ライブラリのみ（外部npmなし）で実現する手段
- **Findings**:
  - **Tailwind CSS Play CDN**: CDN1タグで導入可能。ユーティリティクラスで高度なデザイン。オフライン時は機能しない。
  - **Pico.css CDN**: 軽量・セマンティックHTML向け。カスタマイズ範囲が狭い。
  - **インラインCSS**: 外部依存ゼロ。実装コストが高い。
  - ローカル開発ツールであり常時インターネット接続を想定できる。要件でもCDN利用を許容している。
- **Implications**: Tailwind CSS Play CDN を採用。スタイルが豊富で要件のビジュアル要件（パステル・カード・アニメーション）を最小コストで実現できる。

## Architecture Pattern Evaluation

| Option | 説明 | 強み | リスク | 採用 |
|--------|------|------|--------|------|
| management に直接追加 | api.go に `GET /` を追加し HTML 文字列を返す | 変更箇所最小 | HTML と Go ロジックが混在 | ✗ |
| `internal/portal` 新パッケージ | ポータル配信を分離し management が import | 責務分離・テスト容易 | パッケージが1つ増える | ✓ |
| `internal/server` での処理 | server.go でポータルを先に捕捉 | management 変更なし | server の責務が広がる | ✗ |

## Design Decisions

### Decision: `internal/portal` パッケージの新設

- **Context**: HTML 配信ロジックを management の JSON API から分離したい
- **Alternatives Considered**:
  1. management に直接追加 — HTML文字列が api.go に混入
  2. `internal/portal` 新パッケージ — 責務明確
- **Selected Approach**: `internal/portal` を新設。`Handler` 型が `GET /` を担当し、management の mux から登録される
- **Rationale**: 既存の「パッケージ = ドメイン責務」の構造規約に準拠。management → portal の依存方向は server → management と同様に「外側が内側を呼ぶ」原則に反しないため許容
- **Trade-offs**: パッケージが1つ増えるが、HTML・JS・スタイルを portal パッケージ内で完結させられる
- **Follow-up**: `management_test.go` に `GET /` のレスポンスコード・Content-Type 確認テストを追加する

### Decision: Polling 方式（デフォルト2秒、`--portal-refresh`で変更可能）

- **Context**: サーバ側変更コストを最小化しつつリアルタイム性を実現。デフォルト2秒・設定可能とのユーザー要求に対応
- **Selected Approach**: `setInterval(fetchServices, {{.RefreshIntervalMs}})` でポーリング間隔をテンプレート変数で注入。`html/template` を使用して Go の値を JS 変数に安全に埋め込む
- **Interval Chain**: `start --portal-refresh <sec>` → `ServerConfig.PortalRefreshInterval` → `management.NewAPI(reg, interval)` → `portal.NewHandler(interval)` → HTML テンプレート変数 `RefreshIntervalMs = interval * 1000`
- **Rationale**: ローカル開発ツールとして2秒は即時性が高く、設定可能にすることで開発スタイルに合わせられる。SSEへの変更は将来対応で十分。
- **Trade-offs**: 常時ポーリングによる軽微なネットワーク負荷 vs 実装シンプルさ。`html/template` 導入でポータルが純粋な静的ファイル配信でなくなるが、間隔値の注入は1変数のみ

## Risks & Mitigations

- Tailwind Play CDN はオフライン環境で読み込めない → フォールバック用の最小インラインスタイルをHTML内に定義しておく（基本レイアウトのみ）
- `//go:embed` は Go 1.16+ 必須 → 既存スタックが Go 1.22+ のため問題なし
- 解除操作の誤操作 → 解除ボタンクリック時に `confirm()` ダイアログで確認を挟む
