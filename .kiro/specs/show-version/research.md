# Research & Design Decisions

---
**Purpose**: Capture discovery findings, architectural investigations, and rationale that inform the technical design.

---

## Summary

- **Feature**: `show-version`
- **Discovery Scope**: Simple Addition（既存 cobra CLI へのサブコマンド追加）
- **Key Findings**:
  - 既存のサブコマンドはすべて `cmd/<name>.go` に `newXxxCmd()` ファクトリ関数 + `init()` で登録するパターンを採用
  - Go `-ldflags` による変数上書きはパッケージパス付きの完全修飾名が必要（例: `-X 'github.com/.../internal/version.Version=v1.0.0'`）
  - `runtime.Version()` は外部注入不要で Go バージョンを取得できる

## Research Log

### 既存コマンドパターンの確認

- **Context**: 新コマンドを既存スタイルに揃えるため
- **Sources Consulted**: `cmd/list.go`, `cmd/root.go`
- **Findings**:
  - `newListCmd() *cobra.Command` を定義し、`init()` 内で `rootCmd.AddCommand(newListCmd())` を呼ぶ
  - `RunE` で `error` を返し、`cmd.SilenceUsage = true` でエラー時のヘルプ出力を抑制
  - 出力には `fmt.Fprintln(cmd.OutOrStdout(), ...)` を使用（テスト可能性のため）
- **Implications**: `cmd/version.go` も同パターンで実装する

### ldflags 注入先パッケージの選択

- **Context**: バージョン変数を `-ldflags` でどのパッケージに置くか
- **Findings**:
  - `cmd` パッケージに置く場合: `cmd/version.go` 内変数に直接注入できる
  - `internal/version` パッケージに置く場合: 将来 `rootCmd.Version` フィールドや他の場所での参照も容易
- **Implications**: → 下記 Design Decisions 参照

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| cmd パッケージに変数定義 | `cmd/version.go` にパッケージ変数を置き ldflags で上書き | ファイル数が増えない、シンプル | 将来 root.Version フィールド等から参照するとき import cycle の恐れ | 現状の要件では問題なし |
| internal/version パッケージ | 専用パッケージに変数を集約 | 再利用しやすい、責務が明確 | ファイルが 1 つ増える | Go の慣習に沿った構造 |

## Design Decisions

### Decision: バージョン変数の配置先

- **Context**: `-ldflags` で上書きするパッケージ変数をどこに置くか
- **Alternatives Considered**:
  1. `cmd` パッケージ内 — `cmd/version.go` にまとめてシンプル
  2. `internal/version` パッケージ — 専用パッケージで再利用可能
- **Selected Approach**: `internal/version` パッケージを新設し、`Version`, `Commit`, `BuildDate` 変数を定義する
- **Rationale**: cobra の `rootCmd.Version` フィールドへの割り当て等、将来の拡張時に import cycle を避けられる。Go 標準の内部パッケージ慣習にも沿う
- **Trade-offs**: ファイルが 1 つ増えるが、コード量は最小限
- **Follow-up**: CI/CD (`deploy.yml`) のビルドコマンドに `-ldflags` オプションを追加する

## Risks & Mitigations

- ビルド日時の形式が環境依存になる可能性 — CI/CD 側で RFC3339 形式（`$(date -u +%Y-%m-%dT%H:%M:%SZ)`）を明示的に渡す
- `deploy.yml` の変更漏れ — タスクリストで明示的にチェック項目とする

## References

- [Go ldflags documentation](https://pkg.go.dev/cmd/link) — `-X importpath.name=value` による変数上書き
- [cobra Command.Version field](https://pkg.go.dev/github.com/spf13/cobra#Command) — `--version` フラグとの連携方法
