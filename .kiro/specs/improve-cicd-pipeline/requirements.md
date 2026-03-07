# Requirements Document

## Project Description (Input)
CI/CDパイプラインの改善。

### 改善したい点1

現在、テストパイプラインとデプロイパイプラインの2系統が存在する。
mainブランチに関してはpushされた、あるいはプルリクエストがマージされた段階でテストパイプラインが動く。
そのためタグを打ってpushした場合、デプロイパイプラインだけが動けば良いのだが、テストパイプラインも動いてしまう。

これを解決するためパイプラインを一本化し、次のような動作をするようにパイプラインを構築したい。
- ブランチのpushやプルリクエストのマージ: テスト関連のジョブだけが動作し、デプロイ関連のジョブは動かないようにする
- タグのpush: テスト関連のジョブもデプロイ関連のジョブも両方動くようにする

ジョブの実行順序はテスト関連のジョブが先行し、そのあとでデプロイ関連のジョブが来るように構成すること。

### 改善したい点2

現在、クロスコンパイルいよって作成されたバイナリをリリースに添付している。
これらに加えて、コンテナイメージをビルドしてGitHub Container Registryへpushしたい。

コンテナの実行コマンドは次のような想定:

```
# 基本
docker run -d --name localgate -p 9000:9000 ghcr.io/backpaper0/localgate
# ホスト名を指定。起動時にホスト名を--hostnameパラメーターに渡す
docker run -d --name localgate -p 9000:9000 --hostname localgate.test ghcr.io/backpaper0/localgate
# ポートを変更
docker run -d --name localgate -p 9999:9999 -e PORT=9999 ghcr.io/backpaper0/localgate
```

## Introduction

本ドキュメントは、localgateプロジェクトのCI/CDパイプライン改善に関する要件を定義する。
現在、`test.yml`（全pushで起動）と`deploy.yml`（タグpushで起動）の2ワークフローが独立して存在しており、
タグpush時に両ワークフローが重複して実行される問題がある。
また、リリース時にコンテナイメージをGitHub Container Registry（GHCR）へ公開する機能が未実装である。
本改善により、単一パイプラインへの統合と、コンテナイメージ公開を実現する。

## Requirements

### Requirement 1: パイプラインの一本化

**Objective:** As a 開発者, I want ブランチpush・タグpushのトリガーに応じてテストジョブとデプロイジョブが適切に制御される単一ワークフロー, so that タグpush時にテストワークフローが重複実行されなくなり、CI/CDの実行コストと混乱が解消される

#### Acceptance Criteria

1. When ブランチへのpushまたはプルリクエストのマージが発生した, the CI/CDパイプライン shall テスト関連ジョブのみを実行し、デプロイ関連ジョブを実行しない
2. When `v*`パターンに一致するタグがpushされた, the CI/CDパイプライン shall テスト関連ジョブを先に実行し、その完了後にデプロイ関連ジョブを実行する
3. The CI/CDパイプライン shall テスト関連ジョブとデプロイ関連ジョブを単一のワークフローファイルで管理する
4. When テスト関連ジョブが失敗した, the CI/CDパイプライン shall デプロイ関連ジョブを実行しない
5. The CI/CDパイプライン shall 現行の`test.yml`および`deploy.yml`を廃止し、単一のワークフローファイルへ統合する

### Requirement 2: テストジョブの内容維持

**Objective:** As a 開発者, I want 現行と同等のテスト内容が統合後のパイプラインでも実行される, so that コードの品質チェックが継続して担保される

#### Acceptance Criteria

1. The CI/CDパイプライン shall `gofmt`によるフォーマットチェックをテストジョブとして実行する
2. The CI/CDパイプライン shall `go test ./...`によるユニットテストをテストジョブとして実行する
3. The CI/CDパイプライン shall `go build ./...`によるビルド確認をテストジョブとして実行する

### Requirement 3: バイナリリリースの維持

**Objective:** As a 利用者, I want タグpush時に従来どおりクロスコンパイル済みバイナリがGitHub Releasesに添付される, so that 既存のリリースフローが継続して利用できる

#### Acceptance Criteria

1. When `v*`タグがpushされた, the CI/CDパイプライン shall linux/amd64、linux/arm64、darwin/amd64、darwin/arm64、windows/amd64の5プラットフォーム向けバイナリをクロスコンパイルする
2. When `v*`タグがpushされた, the CI/CDパイプライン shall クロスコンパイル済みバイナリをGitHub Releasesのドラフトリリースに添付する
3. When `v*`タグがpushされた, the CI/CDパイプライン shall リリースタイトルをタグ名（例: `v1.0.0`）に設定する

### Requirement 4: コンテナイメージのビルドとGHCR公開

**Objective:** As a 利用者, I want タグpush時にコンテナイメージがGitHub Container Registryへ公開される, so that Dockerを使ったlocalgateの導入が容易になる

#### Acceptance Criteria

1. When `v*`タグがpushされた, the CI/CDパイプライン shall `ghcr.io/backpaper0/localgate`イメージをビルドしてGHCRへpushする
2. The CI/CDパイプライン shall コンテナイメージにタグ名（例: `v1.0.0`）をバージョンタグとして付与する
3. When コンテナが`PORT`環境変数を指定して起動された, the localgateコンテナ shall 指定されたポートでHTTPリクエストを受け付ける
4. When コンテナが`--hostname`オプションを指定して起動された, the localgateコンテナ shall 指定されたホスト名を自己ホスト名として管理APIのルーティングに使用する
5. The localgateコンテナ shall デフォルトで9000番ポートでHTTPリクエストを受け付ける
6. The CI/CDパイプライン shall GHCRへのpushにGitHubトークン（`GITHUB_TOKEN`）を使用して認証する
