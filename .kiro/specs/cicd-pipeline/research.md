# Research & Design Decisions

---
**Purpose**: CI/CDパイプライン設計における調査結果とアーキテクチャ判断の記録。

---

## Summary
- **Feature**: `cicd-pipeline`
- **Discovery Scope**: Simple Addition（新規ワークフローYAMLファイルの追加）
- **Key Findings**:
  - 既存の `.github/workflows/` ディレクトリは存在しない。新規作成のみ必要。
  - `go.mod` のGoバージョンは `1.22.12`。mise.tomlは `1.26` を指定しているが、`go.mod` を正規バージョンとして扱う。
  - `gh` CLIはGitHub Actions の `ubuntu-latest` ランナーにプリインストール済みのため、追加アクション不要。
  - `gofmt -l .` の出力が空でなければフォーマット違反として失敗させる方法が最もシンプル。

## Research Log

### GitHub Actions: Go のバージョン指定
- **Context**: go.mod は 1.22.12、mise.toml は 1.26 を指定しており、どちらを基準にするか判断が必要。
- **Findings**:
  - `actions/setup-go` の `go-version` には `go.mod` から自動読み取りする `go-version-file: 'go.mod'` オプションが利用可能。
  - これにより go.mod の `go` ディレクティブを正規バージョンとして扱える。
- **Implications**: `go-version-file: 'go.mod'` を採用することでバージョン管理の一元化が可能。

### リリース作成方法の選定
- **Context**: デプロイパイプラインでドラフトリリースを作成する方法を選定。
- **Alternatives**:
  1. `softprops/action-gh-release` — 人気のサードパーティアクション、設定が簡潔
  2. `actions/create-release` — GitHub公式だが非推奨
  3. `gh release create --draft` — CLIを使用、ubuntu-latestにプリインストール済み
- **Selected**: `gh release create --draft` を採用。外部サードパーティへの依存を避け、mise.tomlにすでに `gh` が含まれていることと一致する。
- **Implications**: `GITHUB_TOKEN` 環境変数の設定が必要。

### gofmt チェック方法
- **Context**: ソースコードのフォーマット確認方法。
- **Selected**: `gofmt -l .` の出力が空でなければ `exit 1` で失敗させる1行スクリプト。
- **Rationale**: 追加ツール不要で最もシンプル。

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| 単一ワークフローファイル | test・deployを1ファイルに統合 | ファイル数少 | トリガー条件が複雑化、関心の分離が困難 | 不採用 |
| 分離ワークフローファイル | test.yml / deploy.yml に分割 | 責務明確、トリガーが独立 | ファイル数2つ | **採用** |

## Design Decisions

### Decision: `go-version-file` を使用したGoバージョン管理
- **Context**: ワークフロー内のGoバージョン指定
- **Selected Approach**: `go-version-file: 'go.mod'` を使用
- **Rationale**: go.mod を唯一の正規バージョン定義にすることで管理の一元化
- **Trade-offs**: mise.toml との乖離（1.22 vs 1.26）は許容する

### Decision: `gh` CLIによるリリース作成
- **Context**: ドラフトリリースの作成方法
- **Selected Approach**: `gh release create --draft ${{ github.ref_name }}`
- **Rationale**: サードパーティアクション依存を排除、ubuntu-latestにプリインストール済み
- **Trade-offs**: アクションのUI的な設定より冗長だが依存が少ない

## Risks & Mitigations
- `GITHUB_TOKEN` の権限不足によるリリース作成失敗 — ワークフローに `permissions: contents: write` を明示的に設定することで対応
- go.mod と mise.toml のGoバージョン乖離 — `go-version-file: 'go.mod'` 採用により go.mod を正規バージョンとして固定

## References
- GitHub Actions: actions/setup-go — Go セットアップアクション
- GitHub CLI: `gh release create` — リリース作成コマンド
- gofmt: Go公式フォーマットツール
