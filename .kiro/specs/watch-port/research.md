# Research & Design Decisions

---
**Purpose**: `watch-port` フィーチャーの設計調査ログ。
---

## Summary
- **Feature**: `watch-port`
- **Discovery Scope**: Extension（既存CLIへのサブコマンド追加）
- **Key Findings**:
  - `/proc/net/tcp` および `/proc/net/tcp6` でLISTENポートを取得できる。外部コマンド不要。
  - 既存の `cmd/` パターン（cobra + `init()` + `resolveServerURL()`）をそのまま踏襲できる。
  - ポート監視ループは `internal/watcher` パッケージに切り出すことで、cobra 層とドメインロジックを分離できる。

## Research Log

### `/proc/net/tcp` フォーマットの確認
- **Context**: Linuxでの外部コマンド不要なLISTENポート取得方法を調査
- **Sources Consulted**: Linux kernel ドキュメント, Debian環境での動作確認
- **Findings**:
  - `/proc/net/tcp` の各行: `sl local_address rem_address st ...`
  - `local_address` は `HEXIP:HEXPORT` (IPv4: little-endianの16進数)
  - `st` フィールドが `0A` の行が LISTEN 状態
  - TCPv6 は `/proc/net/tcp6` に同様のフォーマット（アドレス部が32桁の16進数）
  - ポート番号のみ抽出すれば IPv4/v6 の重複は許容できる（同一ポートが両方に現れる場合がある）
- **Implications**:
  - Go標準ライブラリの `os.Open` + `bufio.Scanner` で完結する
  - `strconv.ParseUint(hexPort, 16, 16)` でポート番号をデコード

### 既存CLIパターンの分析
- **Context**: `cmd/` パッケージの既存コマンドとの統合方法を調査
- **Findings**:
  - 各コマンドは `cmd/xxx.go` に `newXxxCmd()` + `init()` で自己登録
  - 共通クライアント型（`resolveServerURL`, `apiError`, `registerServiceRequest`）は `cmd/client.go` に集約
  - `internal/watcher` から `cmd/client.go` をインポートすると循環依存になるため、watcher パッケージは独自の最小HTTPクライアントを持つ
- **Implications**:
  - `cmd/watch.go` は既存パターンに従い `newWatchCmd()` + `init()` で実装
  - `internal/watcher` パッケージを新設し、ポートスキャンとAPIクライアントをドメインロジックとして切り出す

### シグナルハンドリング
- **Context**: コンテナ環境でのプロセス終了時クリーンアップ手段
- **Findings**:
  - Goの `signal.NotifyContext` (Go 1.16+) で SIGINT/SIGTERM を受け取り `context.Context` をキャンセルできる
  - `cmd/watch.go` で context を生成し `watcher.Run(ctx)` に渡す設計にすることで、watcher は context のキャンセルのみ意識すれば良い
- **Implications**:
  - `Watcher.Run(ctx)` が ctx.Done() を検知したとき、管理済みサービスを全解除してから return する

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| A: cmd/watch.go 単体 | ループ・スキャン・APIコールをすべて cmd パッケージ内に実装 | ファイル数最小 | テスト困難、単一ファイルが肥大化 | 他コマンドより複雑なので不適 |
| B: internal/watcher パッケージ分離 | スキャン・ループ・APIクライアントを internal に切り出し | テスト容易、cobra と分離 | 新パッケージ追加コスト | 採用: 既存 internal パターンと一致 |
| C: internal/client パッケージ共通化 | cmd/client.go の HTTP クライアントを internal に移動して共有 | コード重複なし | 既存コード変更が必要 | 変更コストが要件に対して過大。今回は不採用 |

## Design Decisions

### Decision: `PortScanner` インターフェースの導入
- **Context**: `/proc/net/tcp` パーサをテスタブルに保ちたい
- **Alternatives Considered**:
  1. 直接ファイルパスをハードコード — シンプルだがテスト不可
  2. `PortScanner` インターフェースを定義し、テスト時にモックを注入 — テスト可能
- **Selected Approach**: `PortScanner` インターフェースを定義し、本番実装は `ProcNetScanner`
- **Rationale**: Go の標準的な依存注入パターン。steering の "インターフェース定義" 原則に合致。
- **Trade-offs**: インターフェース分が若干コード量増加するが、テスト容易性のコストとして許容範囲

### Decision: ターゲットアドレスを `localhost:{port}` 固定
- **Context**: watch コマンドが監視するポートの登録 target をどうするか
- **Alternatives Considered**:
  1. OS のホスト名を取得 (`os.Hostname()`) — register コマンドと同様
  2. `localhost:{port}` 固定 — シンプル
- **Selected Approach**: コンテナ内のサービスは同一ホストで動くため `localhost:{port}` を使用
- **Rationale**: ユースケースがコンテナ内バックグラウンド実行に限定されており、同一ホストが保証される

### Decision: IPv4/IPv6 重複ポートの扱い
- **Context**: 同一ポートが `/proc/net/tcp` と `/proc/net/tcp6` 両方に現れる場合がある
- **Selected Approach**: ポート番号の `map[int]struct{}` で重複排除。同一ポートを2回登録しない。
- **Rationale**: localgate のサービス名は `port-{n}` の単一エントリであり、IPv4/v6 を区別する必要はない

## Risks & Mitigations
- `/proc/net/tcp` の読み取り権限がない — コンテナ実行環境では通常読み取り可能。エラー時はログ記録して次サイクルへ継続（Req 2.4）
- 大量ポートが同時に開いた場合の登録ループ — 同期的に処理するため、極端に多い場合は遅延が生じる可能性。今回のユースケース（開発環境）では許容範囲
- watch 起動前から存在するポートは対象外 — 起動時スナップショットをベースラインとするため意図的（設計の境界）

## References
- Linux `/proc/net/tcp` 仕様: https://www.kernel.org/doc/html/latest/networking/proc_net_tcp.html
- Go `signal.NotifyContext`: https://pkg.go.dev/os/signal#NotifyContext
