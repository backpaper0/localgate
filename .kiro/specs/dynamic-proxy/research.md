# Research & Design Decisions

---
**Purpose**: Discovery findings and design rationale for `dynamic-proxy`.

---

## Summary

- **Feature**: `dynamic-proxy`
- **Discovery Scope**: New Feature（グリーンフィールド）
- **Key Findings**:
  - Go の `net/http/httputil.ReverseProxy` が要件を満たす最適な標準ライブラリ実装
  - 単一ポートで管理API／プロキシを振り分けるサブドメイン有無判定が最もシンプルな設計
  - `sync.RWMutex` を用いたインメモリマップで再起動不要の動的ルーティングを実現できる

## Research Log

### 言語・ランタイム選定

- **Context**: CLIツール + HTTPリバースプロキシ + 動的ルーティングに最適な言語を選定する
- **Sources Consulted**: Go net/http ドキュメント、Traefik/Caddy のアーキテクチャ参考
- **Findings**:
  - Go: 標準ライブラリに `httputil.ReverseProxy` が含まれ、単一バイナリコンパイル可能
  - Node.js: `http-proxy-middleware` で実装可能だが、ランタイムが必要
  - Rust: 高性能だが開発速度が遅く、ローカルツールには過剰
- **Implications**: Go を採用。標準ライブラリのみで主要機能を実装でき、外部依存が最小

### ポート設計（管理APIとプロキシの共存）

- **Context**: 管理API（POST /services 等）とプロキシ機能を同一ポートで提供するか分けるかを決定
- **Sources Consulted**: 要件 4.4〜4.6 の API 定義、使用例
- **Findings**:
  - Option A: 単一ポート（Host にサブドメインなし → 管理API、あり → プロキシ）
  - Option B: 別ポート（プロキシ用 9000、管理API用 9001）
- **Implications**: Option A を採用。`localgate` が単一プロセスで動作するという要件（5.2）に合致し、設定がシンプル

### ルーティングテーブルの並行制御

- **Context**: 複数のHTTPリクエストが同時にルーティングテーブルを読み書きする場面への対応
- **Findings**:
  - `sync.RWMutex` で読み取りは並行、書き込み（登録・削除）は排他的に制御
  - インメモリマップで十分（永続化不要 — 再起動後にAPIで再登録する設計）
- **Implications**: `ServiceRegistry` に `sync.RWMutex` を持たせる

### CLIフレームワーク

- **Context**: `localgate start` コマンドの実装方法
- **Findings**:
  - `cobra`: サブコマンド対応、広く使われる（kubectl, helm 等）
  - 標準 `flag`: 軽量だがサブコマンド管理が煩雑
- **Implications**: 将来の拡張を考慮し `cobra` を採用

## Architecture Pattern Evaluation

| Option | 説明 | 強み | リスク | 備考 |
|--------|------|------|--------|------|
| 単一ポート（Host判定） | サブドメイン有無でルーティング先を振り分け | シンプル、設定不要 | 管理APIが外部から見えうる | ローカル用途なので許容 |
| 別ポート（プロキシ/管理分離） | プロキシポートと管理ポートを分ける | 明確な分離 | 設定が増える | 要件に「単一コマンド起動」があるため不採用 |
| サービスメッシュ型 | Envoy/Istio的な複雑構成 | 高機能 | 過剰設計 | ローカル開発用途に不適 |

## Design Decisions

### Decision: 言語選定 — Go

- **Context**: HTTPリバースプロキシ + CLIツールをローカル開発者向けに提供
- **Alternatives Considered**:
  1. Node.js/TypeScript — エコシステム豊富だがランタイム必要
  2. Rust — 高性能だが開発コスト高
- **Selected Approach**: Go（標準ライブラリ中心）
- **Rationale**: `net/http/httputil.ReverseProxy` が即使用可能、単一バイナリ、並行処理が簡潔
- **Trade-offs**: Go に不慣れな場合の学習コストあり
- **Follow-up**: Go 1.21+ の `net/http` の `ServeMux` 改良点を活用できるか確認

### Decision: 単一ポートでの管理API／プロキシ共存

- **Context**: 要件 4.4〜4.6 の管理API と 4.2 のプロキシが同一サーバで動く
- **Alternatives Considered**:
  1. 別ポート — 明確な分離だが設定が増える
  2. 単一ポート＋Host判定 — シンプルで要件に合致
- **Selected Approach**: Host ヘッダにサブドメインが含まれない場合は管理APIへルーティング
- **Rationale**: 「単一コマンド起動」「単一プロセス動作」要件に直接合致
- **Trade-offs**: ローカル外からの管理API露出リスク（ローカル用途なので許容）

### Decision: インメモリルーティングテーブル（永続化なし）

- **Context**: 要件 7（動的設定）でプロセス再起動不要を求めているが、プロセス終了後の永続化は要件外
- **Alternatives Considered**:
  1. ファイル永続化（JSON）
  2. SQLite
  3. インメモリのみ
- **Selected Approach**: インメモリ `map[string]string` + `sync.RWMutex`
- **Rationale**: 要件にデータ永続化の記述がなく、ローカル開発用途では起動後に API で登録し直せば十分
- **Trade-offs**: プロセス再起動でルーティングテーブルがリセットされる

## Risks & Mitigations

- ルーティングテーブルへの同時読み書き競合 → `sync.RWMutex` で対処
- バックエンドサービスへの接続失敗時の処理 → `httputil.ReverseProxy` の `ErrorHandler` で 502 を返す
- サブドメイン抽出のエッジケース（ポート番号付きHostヘッダ等） → Host から `:port` を除去してからパース

## References

- Go net/http/httputil.ReverseProxy — 標準リバースプロキシ実装
- cobra CLI フレームワーク — `github.com/spf13/cobra`
- sync.RWMutex — Go 標準並行制御プリミティブ
