# Research & Design Decisions

---
**Purpose**: 設計フェーズにおけるディスカバリー結果・アーキテクチャ調査・決定根拠を記録する。

---

## Summary

- **Feature**: `service-cli-management`
- **Discovery Scope**: Extension（既存cobraベースCLIへのサブコマンド追加）
- **Key Findings**:
  - 既存の `cmd/start.go` が明確なパターンを提供しており、新コマンドはそのまま踏襲できる
  - 管理HTTP APIはすでに実装済み（`POST /services`、`DELETE /services/{name}`、`GET /services`）であり、CLIはHTTPクライアントとして呼び出すだけでよい
  - サーバーURL解決ロジック（フラグ → 環境変数 → デフォルト）は3コマンドで共通のため、`cmd/` パッケージ内の共有ヘルパー関数として切り出すのが適切

## Research Log

### 既存CLIパターンの分析

- **Context**: 新コマンドのファイル構成とフラグ定義パターンを確認
- **Sources Consulted**: `cmd/root.go`、`cmd/start.go`
- **Findings**:
  - コマンドは1ファイル1コマンドで `cmd/` に配置（`cmd/start.go` → `localgate start`）
  - `init()` 内で `rootCmd.AddCommand()` を呼び出し、フラグは `cmd.Flags()` で定義
  - エラーは `RunE` から `error` として返却（`fmt.Errorf` / センチネルエラー）
  - 標準出力は `fmt.Fprintf(os.Stdout, ...)` / 標準エラーは `fmt.Fprintf(os.Stderr, ...)`
  - 新依存ライブラリは不要。Go標準ライブラリの `net/http`、`encoding/json` で完結
- **Implications**: 新コマンドは既存パターンに完全準拠。アーキテクチャ変更なし

### 管理APIエンドポイントの確認

- **Context**: CLIが呼び出すHTTP APIの正確な仕様を確認
- **Sources Consulted**: `internal/management/api.go`
- **Findings**:
  - `POST /services`: JSONボディ `{"name": string, "target": string}` → 201 Created, レスポンス `{"name": string, "target": string}`
  - `DELETE /services/{name}`: 204 No Content（成功）、404 Not Found（対象なし）
  - `GET /services`: 200 OK, レスポンス `{"services": [{"name": string, "target": string}]}`
  - エラーレスポンス形式: `{"error": string}`
- **Implications**: CLIの型定義はAPIのJSONスキーマに直接対応するstructを定義すればよい

### サーバーURL解決の共通化

- **Context**: 3コマンドすべてが同じ優先順位ルール（`--server` > `LOCALGATE_SERVER` > デフォルト）を持つ
- **Findings**: 重複を避けるため共有ヘルパー関数として実装するのが適切
- **Implications**: `cmd/` パッケージ内に `resolveServerURL(flagValue string) string` を用意し、各コマンドから呼び出す

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| 各コマンドに直接実装 | HTTPロジックを各ファイルに記述 | シンプル | 3ファイルに重複コードが発生 | 小規模なら許容範囲 |
| 共有ヘルパー（採用） | `cmd/` 内にURL解決 + HTTPクライアントヘルパーを集約 | DRY、テスト容易 | 若干のファイル増加 | cobraのパターンに自然に収まる |
| `internal/client` パッケージ | 専用クライアントパッケージを作成 | 最も疎結合 | このスコープでは過剰 | 将来的な拡張時に検討 |

## Design Decisions

### Decision: 共有ヘルパーファイル `cmd/client.go` の導入

- **Context**: 3コマンドで共通のサーバーURL解決とHTTPクライアントロジックが必要
- **Alternatives Considered**:
  1. 各コマンドファイルに直接実装 — 重複コードが発生
  2. `internal/client` パッケージ新設 — このスコープでは過剰な抽象化
- **Selected Approach**: `cmd/client.go` に共有ヘルパーを定義（`resolveServerURL`、共通HTTPリクエスト送信関数）
- **Rationale**: `cmd/` パッケージ内に収めることで既存アーキテクチャを変更せず、かつ重複を排除できる
- **Trade-offs**: ファイルが1つ増えるが、コードの明確性が向上する
- **Follow-up**: 将来的にコマンドが増えた場合は `internal/client` への昇格を検討

### Decision: HTTPクライアントの直接利用（ライブラリなし）

- **Context**: HTTP APIへのリクエストにサードパーティHTTPクライアントライブラリを使用するか
- **Selected Approach**: Go標準ライブラリの `net/http` を直接使用
- **Rationale**: プロジェクトの「標準ライブラリ優先」方針（tech.md）に準拠。単純なREST呼び出しにライブラリ追加は不要
- **Trade-offs**: ボイラープレートが若干増えるが、依存が増えない

## Risks & Mitigations

- サーバーが起動していない場合の接続エラー — `net/http` の接続エラーをキャッチし、ユーザーフレンドリーなメッセージで標準エラー出力へ表示（要件4.5準拠）
- APIレスポンスのJSONパースエラー — デコード失敗時もエラーメッセージを表示して終了コード1で終了

## References

- `cmd/start.go` — 既存コマンド実装パターンの参照元
- `internal/management/api.go` — 管理APIエンドポイントの仕様
- tech.md — 「標準ライブラリ優先」設計方針
