# Research & Design Decisions

---
**Purpose**: Capture discovery findings, architectural investigations, and rationale that inform the technical design.

---

## Summary

- **Feature**: `container-environment-fix`
- **Discovery Scope**: Extension（既存ルーティングロジックの拡張）
- **Key Findings**:
  - バグの根本原因は `proxy.ExtractSubdomain` ではなく `server.ServeHTTP` のルーティング判断にある
  - `ExtractSubdomain("localgate.test:9000")` が `"localgate"` を返すのは仕様通りの動作であり、関数自体の変更は不要
  - 自己ホスト名チェックを `ExtractSubdomain` の呼び出し**前**に挿入することで最小限の変更で修正可能

## Research Log

### ルーティング判定ロジックの分析

- **Context**: `foobar` コンテナが `http://localgate.test:9000/services` にアクセスすると "service not found" になる原因
- **Sources Consulted**: `internal/server/server.go`, `internal/proxy/subdomain.go`
- **Findings**:
  - `server.ServeHTTP` は `proxy.ExtractSubdomain(host)` の結果が空文字列かどうかだけで管理API/プロキシを判定している
  - `ExtractSubdomain("localgate.test:9000")` → `strings.Split("localgate.test", ".")` → `["localgate", "test"]` → `parts[0]` = `"localgate"` を返す
  - サーバは自己ホスト名の概念を持たないため、`localgate.test` を `localgate` サービスへのプロキシ対象と誤認する
- **Implications**: 修正箇所は `server.go` のルーティング判定部分のみ。`ExtractSubdomain` の変更は不要。

### `ExtractSubdomain` の既存動作確認

- **Context**: `foo.localgate.test:9000` のような「サブドメイン.自己ホスト名」形式が正しくプロキシされるか確認
- **Findings**:
  - `ExtractSubdomain("foo.localgate.test:9000")` → parts=`["foo","localgate","test"]` → `parts[0]` = `"foo"` を返す ✓
  - この動作は偶然ではなく「先頭ラベルをサブドメインとして返す」という既存の仕様と一致している
  - `foo.localhost:9000` の場合と同じロジックで正しく動作する
- **Implications**: `ExtractSubdomain` への変更は不要。変更スコープを `server.go` と `cmd/start.go` に限定できる。

### テスト構造の確認

- **Context**: 既存テストへの影響範囲の把握
- **Findings**:
  - `server_test.go` の `newTestServer()` は `ServerConfig{Port: 0}` で生成しており、`Hostname` フィールド追加はゼロ値（空文字列）で後方互換
  - `TestManagementAPIRegisterAndList` など管理APIテストはHostヘッダを明示指定していないため、`net/http/httptest` のデフォルトURL（`127.0.0.1:PORT`）が使われる
  - IPアドレスの場合 `ExtractSubdomain` が `""` を返すため、既存テストは引き続き管理APIへルーティングされ壊れない
- **Implications**: 既存テストへの影響なし。新規テストケースを追加するだけでよい。

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| A: 自己ホスト名の事前チェック | `ServeHTTP` で `ExtractSubdomain` の前に自己ホスト名セットとの照合を行う | 変更箇所最小・`ExtractSubdomain` 不変・後方互換 | なし | **採用** |
| B: `ExtractSubdomain` に base hostname を渡す | `ExtractSubdomain(host, baseHostname string)` にシグネチャ変更 | ロジック集約 | シグネチャ変更による既存テストへの影響・関数の責務拡大 | 不採用 |
| C: ミドルウェアチェーン | 自己ホスト名チェックを独立ミドルウェアとして実装 | 関心の分離 | 過剰設計（1箇所のみの修正に対して構造が重い） | 不採用 |

## Design Decisions

### Decision: 自己ホスト名セットを `ProxyServer` に保持する

- **Context**: 複数の自己ホスト名（`localhost` + 設定値）を効率的に照合する必要がある
- **Alternatives Considered**:
  1. スライス (`[]string`) でリニアサーチ
  2. マップ (`map[string]struct{}`) でO(1)ルックアップ
- **Selected Approach**: `map[string]struct{}` を `ProxyServer` フィールドとして保持し、キーはすべて小文字に正規化
- **Rationale**: 実用上の追加ホスト名は少数（1〜2件）だがO(1)ルックアップは実装コストゼロで将来的にも正確
- **Trade-offs**: わずかにメモリを使うが無視できる
- **Follow-up**: `NewProxyServer` で初期化時に `"localhost"` を常に挿入することを確認

### Decision: `--hostname` フラグは追加セマンティクス（置き換えではない）

- **Context**: ホストマシンから `localhost:9000` でアクセスするユースケースを維持しながら、コンテナ内からの `localgate.test:9000` も管理APIとして扱いたい
- **Alternatives Considered**:
  1. `--hostname` でデフォルト値を置き換える（= `localhost` を使えなくなる）
  2. `--hostname` を追加するセマンティクス（`localhost` は常に固定）
- **Selected Approach**: オプション 2
- **Rationale**: 要件 2.1「`localhost` は設定に関わらず常に管理APIホスト名」を満たす。ホストマシンからの操作性を損なわない。
- **Trade-offs**: なし

## Risks & Mitigations

- **大文字・小文字の不一致**: Hostヘッダは大文字を含む可能性がある → `strings.ToLower` で正規化してからマップ照合
- **`net.SplitHostPort` のエラー処理**: ポートなしHostヘッダで `SplitHostPort` がエラーを返す場合 → エラー時はそのまま元の文字列を使う（既存 `ExtractSubdomain` と同じパターン）
- **既存テストへの影響**: `ServerConfig` に `Hostname` フィールドを追加してもゼロ値が空文字列であるため、既存の `ServerConfig{Port: 0}` コードは変更不要

## References

- `internal/server/server.go` — 現在のルーティング実装
- `internal/proxy/subdomain.go` — `ExtractSubdomain` の実装
- `internal/server/server_test.go` — 既存テスト構造
