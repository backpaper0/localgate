# Research & Design Decisions

---
**Purpose**: Capture discovery findings, architectural investigations, and rationale that inform the technical design.

---

## Summary

- **Feature**: `improve-cicd-pipeline`
- **Discovery Scope**: Extension（既存CI/CDワークフローへの統合）
- **Key Findings**:
  - `localgate start` は `--port`（整数）と `--hostname`（文字列）フラグを持ち、コンテナエントリポイントでこれらに環境変数を渡す設計が適合する
  - GitHub Actions の `if: startsWith(github.ref, 'refs/tags/v')` 条件と `needs:` によるジョブ依存でパイプライン一本化が実現できる
  - マルチアーキテクチャコンテナビルドには `docker/build-push-action` + QEMU が標準的だが、単一アーキテクチャ（linux/amd64）から開始するシンプルな方針も選択肢となる

## Research Log

### 既存CLIインターフェースの調査

- **Context**: コンテナエントリポイントの設計に必要な CLI フラグを把握するため
- **Sources Consulted**: `cmd/start.go` 直接調査
- **Findings**:
  - `--port int`（デフォルト9000）: 待ち受けポート番号
  - `--hostname string`（デフォルト空文字）: 追加の自己ホスト名
  - ポート範囲バリデーション（1〜65535）あり
- **Implications**: Dockerfile のエントリポイントで `PORT` 環境変数を `--port` フラグに、コンテナの `$HOSTNAME` 環境変数を `--hostname` フラグに渡す設計で要件を満たせる

### 現行GitHub Actionsワークフローの調査

- **Context**: 統合前の現状把握
- **Sources Consulted**: `.github/workflows/test.yml`, `.github/workflows/deploy.yml`
- **Findings**:
  - `test.yml`: `on: push`（全push対象、タグpushも含む）→ フォーマットチェック・テスト・ビルド
  - `deploy.yml`: `on: push: tags: v*` → クロスコンパイル → GitHub Release ドラフト作成
  - タグpush時に両ワークフローが独立して並行実行される（テスト重複の原因）
- **Implications**: 単一ワークフローに統合し、ジョブ条件で制御することで重複実行を解消できる

### GitHub Actions ジョブ制御パターン

- **Context**: 単一ワークフローでブランチpushとタグpushの挙動を分岐させる方法
- **Findings**:
  - `if: startsWith(github.ref, 'refs/tags/v')` でタグ専用ジョブを定義可能
  - `needs: [test]` でジョブ依存を定義し、テスト完了後にデプロイジョブを実行できる
  - `on: [push, pull_request]` で両イベントをカバーし、プルリクエストのマージ（= baseブランチへのpush）も対応
- **Implications**: 単一ワークフローで全ユースケースをカバーできる

### GHCRへのコンテナイメージ公開

- **Context**: GitHub Container Registry へのイメージ公開方法の確認
- **Findings**:
  - `docker/login-action` で `ghcr.io` へ `GITHUB_TOKEN` 認証が可能
  - `docker/metadata-action` でタグ（`v1.0.0`）やラベルを自動生成できる
  - `docker/build-push-action` でビルドとpushを一括実行できる
  - パッケージ書き込みには `permissions: packages: write` が必要
- **Implications**: `packages: write` パーミッションをリリースジョブに付与する必要がある

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| 単一ワークフロー + ジョブ条件 | `ci.yml` 1ファイル、ジョブ間に `if:` と `needs:` を組み合わせる | 管理容易、重複なし、依存関係明示 | ワークフロー定義がやや複雑になる | 今回採用 |
| 複数ワークフロー + フィルタ改善 | 既存2ファイルを維持しつつ `test.yml` にタグ除外フィルタを追加 | 変更量が少ない | 2ファイルの管理が継続、タグpush時のtest先行保証が難しい | 不採用 |
| Reusable Workflows | テストジョブを再利用可能なワークフローとして抽出 | DRYで拡張しやすい | オーバーエンジニアリング、現規模では不要 | 将来の選択肢 |

## Design Decisions

### Decision: 単一ワークフローへの統合

- **Context**: タグpush時にテストワークフローが重複実行される問題
- **Alternatives Considered**:
  1. `test.yml` に `tags-ignore` フィルタを追加（2ファイル維持）
  2. 単一 `ci.yml` に統合
- **Selected Approach**: 単一 `ci.yml` に統合し、ジョブ条件で制御
- **Rationale**: 単一ファイルで全パイプラインの動作を把握でき、テスト→デプロイの依存関係が明示的になる
- **Trade-offs**: ワークフローファイルが長くなるが、機能ごとにジョブが分かれているので可読性は維持できる
- **Follow-up**: 既存の `test.yml` と `deploy.yml` を削除する

### Decision: リリースジョブの分割（バイナリ vs コンテナ）

- **Context**: バイナリリリースとコンテナイメージ公開は独立した処理
- **Alternatives Considered**:
  1. 単一の `release` ジョブにバイナリとコンテナを統合
  2. `release-binary` と `release-container` を別ジョブとして並行実行
- **Selected Approach**: 別ジョブ（`release-binary`, `release-container`）として並行実行
- **Rationale**: 両処理は独立しており並行実行でパイプライン全体の実行時間を短縮できる。また失敗時の責任範囲が明確になる
- **Trade-offs**: ジョブ数が増えるが、各ジョブの責務が明確になる

### Decision: Dockerfileのエントリポイント設計

- **Context**: `PORT` 環境変数とコンテナホスト名を `localgate start` コマンドに渡す方法
- **Alternatives Considered**:
  1. シェルスクリプトのエントリポイント
  2. `CMD` シェル形式（`CMD localgate start ...`）
- **Selected Approach**: シェルスクリプト (`entrypoint.sh`) を `ENTRYPOINT` に設定
- **Rationale**: シェル変数展開（`${PORT:-9000}`）とシグナル伝播（`exec`）を確実に処理できる
- **Trade-offs**: ファイルが1つ増えるが、動作が明確で保守しやすい

### Decision: コンテナベースイメージ

- **Context**: 最小限のランタイムイメージ選択
- **Alternatives Considered**:
  1. `gcr.io/distroless/static` — シェルなし、最小サイズ
  2. `alpine` — シェルあり、軽量
  3. `debian:slim` — シェルあり、互換性高い
- **Selected Approach**: `alpine` を選択
- **Rationale**: エントリポイントスクリプトにシェルが必要（`sh -c` で変数展開）。distroless はシェルがなくエントリポイント設計が複雑化する。debian:slim より alpine の方が軽量
- **Trade-offs**: distroless より若干大きいが、glibc ではなく musl libc を使うため Go の静的バイナリとの相性も良い

## Risks & Mitigations

- `GITHUB_TOKEN` の `packages: write` パーミッション不足 → ワークフローに `permissions: packages: write` を明示的に付与
- コンテナビルド時のクロスプラットフォーム対応 — 初期は linux/amd64 のみとし、必要に応じて QEMU を追加
- `$HOSTNAME` がコンテナ内で Docker `--hostname` の値を正しく反映するか — Docker仕様で `--hostname` は環境変数 `$HOSTNAME` に反映されるため問題なし

## References

- GitHub Actions: `if` context — https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/evaluate-expressions-in-workflows-and-actions
- `docker/login-action` — https://github.com/docker/login-action
- `docker/build-push-action` — https://github.com/docker/build-push-action
- `docker/metadata-action` — https://github.com/docker/metadata-action
- GHCR認証 — https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry
