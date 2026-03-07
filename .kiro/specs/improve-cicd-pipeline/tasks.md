# Implementation Plan

- [x] 1. 既存ワークフローファイルを廃止する
  - `test.yml` と `deploy.yml` を削除する
  - 削除により、タグpush時の重複実行問題が解消される土台を作る
  - _Requirements: 1.3, 1.5_

- [x] 2. コンテナ定義ファイルを作成する

- [x] 2.1 (P) Dockerfile を作成する
  - `golang:1.22` をビルドステージ、`alpine:3` をランタイムステージとするマルチステージビルドを定義する
  - ビルドステージでは `CGO_ENABLED=0 GOOS=linux` を指定して静的バイナリをビルドする
  - ランタイムステージでビルド済みバイナリと `entrypoint.sh` をコピーし、実行権限を付与する
  - `EXPOSE 9000` でデフォルトポートをドキュメント化する
  - `ENTRYPOINT ["/entrypoint.sh"]` でエントリポイントを指定する
  - _Requirements: 4.1, 4.3, 4.4, 4.5_

- [x] 2.2 (P) entrypoint.sh を作成する
  - `#!/bin/sh` で始まるシェルスクリプトを作成する
  - `PORT` 環境変数を `--port` フラグへ変換し、未指定時はデフォルト `9000` を使用する
  - コンテナの `HOSTNAME` 環境変数（Docker `--hostname` オプションで設定される）を `--hostname` フラグへ渡す
  - `exec` を使用して localgate バイナリを PID 1 として起動し、シグナル伝播を保証する
  - _Requirements: 4.3, 4.4, 4.5_

- [x] 3. 統合ワークフロー ci.yml を作成する

- [x] 3.1 test ジョブを実装する
  - `on: [push, pull_request]` トリガーで ci.yml を作成し、ブランチpush・タグpush・PRすべてを捕捉する
  - `actions/checkout@v4` と `actions/setup-go@v5`（go.mod 指定バージョン）を使用する
  - `test -z "$(gofmt -l .)"` でフォーマット違反を検出するステップを追加する
  - `go test ./...` でユニットテストを実行するステップを追加する
  - `go build ./...` でビルド可能性を確認するステップを追加する
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 2.1, 2.2, 2.3_

- [x] 3.2 release-binary ジョブを実装する
  - `needs: [test]` と `if: startsWith(github.ref, 'refs/tags/v')` を設定し、テスト成功後・タグpush時のみ実行されるようにする
  - `permissions: contents: write` を付与する
  - `GOOS`/`GOARCH` 環境変数を使用し、linux/amd64、linux/arm64、darwin/amd64、darwin/arm64、windows/amd64 の 5 プラットフォーム向けバイナリをクロスコンパイルする
  - `gh release create` で `github.ref_name`（タグ名）をタイトルとするドラフトリリースを作成し、全バイナリを添付する
  - _Requirements: 1.2, 1.4, 3.1, 3.2, 3.3_

- [x] 3.3 release-container ジョブを実装する
  - `needs: [test]` と `if: startsWith(github.ref, 'refs/tags/v')` を設定し、テスト成功後・タグpush時のみ実行されるようにする
  - `permissions: packages: write` を付与する
  - `docker/login-action@v3` で `GITHUB_TOKEN` を使用して `ghcr.io` へ認証する
  - `docker/metadata-action@v5` でタグ名（`v1.0.0`、`v1.0`、`v1`、`latest`）とOCI標準ラベルを生成する
  - `docker/build-push-action@v6` でイメージをビルドし `ghcr.io/backpaper0/localgate` へ push する
  - _Requirements: 1.2, 1.4, 4.1, 4.2, 4.6_
