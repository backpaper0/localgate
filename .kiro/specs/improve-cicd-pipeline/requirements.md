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

## Requirements
<!-- Will be generated in /kiro:spec-requirements phase -->
