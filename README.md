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

[GitHub Releases](https://github.com/backpaper0/localgate/releases) からビルド済みバイナリをダウンロードできる。

またはGoでインストール:

```bash
go install github.com/backpaper0/localgate@latest
```

またはソースからビルド:

```bash
git clone https://github.com/backpaper0/localgate
cd localgate
go build -o localgate .
```

またはDockerで利用:

```bash
docker run -d -p 9000:9000 ghcr.io/backpaper0/localgate
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

コンテナ環境など、`localhost` 以外のホスト名で管理APIにアクセスする場合は `--hostname` を指定する:

```bash
localgate start --hostname localgate.test
```

`--hostname` を指定すると、そのホスト名への直接アクセスが管理APIとして扱われ、サブドメイン付き（例: `foo.localgate.test`）はプロキシとして動作する。

### DNS / hosts 設定

`*.localhost` がローカルホストへ解決されるよう設定する。macOS / Linux では `/etc/hosts` に追加するか、`dnsmasq` 等を利用する。

```
127.0.0.1  foo.localhost
127.0.0.1  bar.localhost
```

## CLIによるサービス管理

サーバー起動後、CLIサブコマンドでサービスを管理できる。

### サービス登録

```bash
localgate register <name> <target>
```

`target` にはホスト名とポートの組み合わせ（`localhost:3000`）またはポート番号のみ（`3000`）を指定できる。ポート番号のみの場合は自動でマシンのホスト名が補完される。

```bash
# ホスト:ポートで登録
localgate register foo localhost:3000

# ポートのみで登録（ホスト名を自動補完）
localgate register foo 3000
```

同名サービスが既に登録されている場合は確認プロンプトが表示される:

```
サービス 'foo' は既に 'localhost:3000' として登録されています。上書きしますか？ [y/N]:
```

確認をスキップして強制上書きする場合は `--force` / `-f` フラグを使う:

```bash
localgate register --force foo localhost:4000
```

### サービス解除

```bash
localgate unregister <name>
```

### サービス一覧

```bash
localgate list
```

### バージョン確認

```bash
localgate version
```

### ポートの自動監視

```bash
localgate watch
```

起動中のLISTENポートを定期的にスキャンし、新たに検出されたポートを自動登録、閉じられたポートを自動解除する。サービス名はホスト名とポート番号から自動生成される（例: `myhost-3000`）。

ポーリング間隔（秒）を変更する場合:

```bash
localgate watch --interval 5
```

### サーバーURLの指定

各コマンドは `--server` フラグまたは環境変数 `$LOCALGATE_SERVER` でサーバーURLを指定できる。省略時は `http://localhost:9000` が使われる。

```bash
localgate register --server http://localgate.test:9000 foo localhost:3000
```

## ポータル画面

サーバ起動後、ブラウザで `http://localhost:9000` にアクセスすると、登録済みサービスの一覧をグラフィカルに確認できる管理ポータルが表示される。サービス一覧は定期的に自動更新される。

## 管理API

サーバ起動後、サブドメインを含まないホスト（例: `localhost:9000`）へのリクエストが管理APIとして処理される。

### サービス登録

```
POST /services
Content-Type: application/json

{"name": "foo", "target": "localhost:3000"}
```

同名サービスが既に存在する場合は 409 Conflict が返る:

```json
{"error": "service already exists", "existing_target": "localhost:3000"}
```

`X-Force-Overwrite: true` ヘッダーを付与すると確認なしで上書きできる:

```
POST /services
Content-Type: application/json
X-Force-Overwrite: true

{"name": "foo", "target": "localhost:4000"}
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
{
  "services": [
    {"name": "foo", "target": "localhost:3000"},
    {"name": "bar", "target": "localhost:8000"}
  ]
}
```

## コンテナ環境での利用

Docker ネットワーク内から管理APIにアクセスする場合、`--hostname` でコンテナのホスト名を指定する。

```bash
docker network create --internal=true prv_net
docker network create pub_net
docker run -d --name=lg --hostname=localgate.test -p 9000:9000 \
  --network=prv_net localgate start --hostname localgate.test
docker network connect pub_net lg
```

同じネットワーク内のコンテナから:

```bash
# サービス登録
curl -X POST http://localgate.test:9000/services \
  -H 'Content-Type: application/json' \
  -d '{"name":"foobar","target":"foobar:8080"}'

# プロキシ経由でアクセス
curl http://foobar.localgate.test:9000/
```

## ユースケース

- ローカル開発環境で複数サービスを単一ポートで公開
- マイクロサービス開発時のサービスゲートウェイ
- コンテナ環境での簡易リバースプロキシ

## 要件

- Go 1.26 以上

## ライセンス

[LICENSE](LICENSE) を参照。
