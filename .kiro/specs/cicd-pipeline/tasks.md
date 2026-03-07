# Implementation Plan

- [ ] 1. (P) テストワークフローの作成
- [ ] 1.1 ワークフロー基本構造とGoセットアップの定義
  - `.github/workflows/test.yml` を新規作成する
  - トリガーを `on: push`（全ブランチ）に設定する
  - `actions/checkout@v4` でソースコードをチェックアウトする
  - `actions/setup-go@v5` で `go-version-file: 'go.mod'` を指定してGo環境をセットアップする
  - _Requirements: 1.1, 1.5_

- [ ] 1.2 フォーマットチェックステップの実装
  - `gofmt -l .` を実行し、出力が空でなければ失敗として扱うステップを追加する
  - `test -z "$(gofmt -l .)"` の形式で1行スクリプトとして記述する
  - _Requirements: 1.1, 1.2_

- [ ] 1.3 テスト実行ステップの実装
  - `go test ./...` を実行するステップをフォーマットチェックの後に追加する
  - テストが失敗した場合はGitHub Actionsが自動でワークフローを失敗ステータスとする
  - _Requirements: 1.1, 1.3_

- [ ] 1.4 ビルド確認ステップの実装
  - `go build ./...` を実行するステップをテストの後に追加する
  - ビルドが失敗した場合はGitHub Actionsが自動でワークフローを失敗ステータスとする
  - _Requirements: 1.1, 1.4_

- [ ] 2. (P) デプロイワークフローの作成
- [ ] 2.1 ワークフロー基本構造とGoセットアップの定義
  - `.github/workflows/deploy.yml` を新規作成する
  - トリガーを `on: push: tags: ['v*']` に設定する
  - `permissions: contents: write` をジョブスコープで設定しリリース作成を許可する
  - `actions/checkout@v4` でソースコードをチェックアウトする
  - `actions/setup-go@v5` で `go-version-file: 'go.mod'` を指定してGo環境をセットアップする
  - _Requirements: 2.1, 2.6_

- [ ] 2.2 バイナリビルドステップの実装
  - `go build -o localgate .` を実行するステップを追加する
  - ビルド失敗時はワークフローを失敗ステータスで終了し、後続のリリース作成ステップを実行しない
  - _Requirements: 2.1, 2.2, 2.5_

- [ ] 2.3 ドラフトリリース作成ステップの実装
  - `gh release create "${{ github.ref_name }}" localgate --draft --title "${{ github.ref_name }}"` を実行するステップをビルドの後に追加する
  - 環境変数 `GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}` を設定する
  - `github.ref_name` でpushされたタグ名をリリースタグとして使用する
  - ビルド成果物 `localgate` バイナリをリリース成果物として添付する
  - _Requirements: 2.2, 2.3, 2.4_
