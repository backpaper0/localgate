# localgate

ローカルサービスのための動的リバースプロキシ。

サブドメインベースのルーティングにより、複数のローカルHTTPサービスを単一エンドポイントから利用できる。サービスの登録・解除は再起動不要のHTTP APIで行える。

## 概要

```
http://foo.localhost:9000  →  localhost:3000
http://bar.localhost:9000  →  localhost:8000
http://baz.localhost:9000  →  localhost:8080
```

Hostヘッダのサブドメイン部分（`foo`、`bar` など）をキーにバックエンドサービスへリクエストを転送する。

## インストール

```bash
go install localgate@latest
```

またはソースからビルド:

```bash
git clone https://github.com/urgm/localgate
cd localgate
go build -o localgate .
```

## 使い方

### サーバ起動

```bash
localgate start
```

デフォルトでポート `9000` で待ち受ける。ポートを変更する場合:

```bash
localgate start --port 8080
```

### DNS / hosts 設定

`*.localhost` がローカルホストへ解決されるよう設定する。macOS / Linux では `/etc/hosts` に追加するか、`dnsmasq` 等を利用する。

```
127.0.0.1  foo.localhost
127.0.0.1  bar.localhost
```

## 管理API

サーバ起動後、サブドメインを含まないホスト（例: `localhost:9000`）へのリクエストが管理APIとして処理される。

### サービス登録

```
POST /services
Content-Type: application/json

{"name": "foo", "target": "localhost:3000"}
```

### サービス削除

```
DELETE /services/foo
```

### サービス一覧

```
GET /services
```

レスポンス例:

```json
[
  {"name": "foo", "target": "localhost:3000"},
  {"name": "bar", "target": "localhost:8000"}
]
```

## ユースケース

- ローカル開発環境で複数サービスを単一ポートで公開
- マイクロサービス開発時のサービスゲートウェイ
- コンテナ環境での簡易リバースプロキシ

## 要件

- Go 1.22 以上

## ライセンス

[LICENSE](LICENSE) を参照。
