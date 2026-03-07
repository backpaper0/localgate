# Requirements Document

## Project Description (Input)
GitHub Actionsを使用してCI/CDパイプラインを構築する。

次の2つのパイプラインを構築すること。

**テストパイプライン**
- トリガー: pushされたとき
- 実施すること:
    - ソースコードがフォーマットされていることのチェック
    - すべてのテストの実行
    - バイナリのビルド

**デプロイパイプライン**
- トリガー: `v<バージョン番号>`タグがpushされたとき
- 実施すること:
    - バイナリのビルド
    - ドラフト状態のリリースを作成

## Introduction

localgate（Go製の動的リバースプロキシ）のCI/CDパイプラインをGitHub Actionsで構築する。コードの品質担保を自動化するテストパイプラインと、バージョンタグを起点にリリース成果物を作成するデプロイパイプラインの2本立てで構成する。

## Requirements

### Requirement 1: テストパイプライン

**Objective:** As a 開発者, I want pushのたびに自動でフォーマット・テスト・ビルドが実行される, so that コードの品質問題を早期に検出できる

#### Acceptance Criteria

1. When コードがいずれかのブランチにpushされた, the CI Pipeline shall フォーマットチェック・テスト・ビルドの3ステップを順に実行する
2. When `gofmt` によるフォーマットチェックを実行した and フォーマットが崩れているファイルが存在する, the CI Pipeline shall ワークフローを失敗ステータスで終了する
3. When `go test ./...` によるテストを実行した and いずれかのテストが失敗した, the CI Pipeline shall ワークフローを失敗ステータスで終了する
4. When `go build` によるバイナリビルドを実行した and ビルドが失敗した, the CI Pipeline shall ワークフローを失敗ステータスで終了する
5. The CI Pipeline shall Go 1.22以上のバージョンでジョブを実行する

### Requirement 2: デプロイパイプライン

**Objective:** As a メンテナー, I want バージョンタグのpushを起点にリリース成果物が自動生成される, so that 手動作業なしに再現性のあるリリースを作成できる

#### Acceptance Criteria

1. When `v` で始まるタグ（例: `v1.0.0`）がpushされた, the Deploy Pipeline shall バイナリビルドおよびドラフトリリースの作成を実行する
2. When `go build` によるバイナリビルドが完了した, the Deploy Pipeline shall ビルドされたバイナリをリリース成果物としてアップロードする
3. When GitHub Releaseを作成する, the Deploy Pipeline shall ドラフト状態（`draft: true`）でリリースを作成する
4. When GitHub Releaseを作成する, the Deploy Pipeline shall pushされたタグ名をリリースのタグとして使用する
5. If バイナリビルドが失敗した, the Deploy Pipeline shall ワークフローを失敗ステータスで終了し、リリースを作成しない
6. The Deploy Pipeline shall Go 1.22以上のバージョンでジョブを実行する
