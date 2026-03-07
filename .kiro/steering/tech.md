# Technology Stack

## Architecture

単一バイナリのCLIアプリケーション。HTTPリクエストをサブドメインで振り分け、登録済みサービスへリバースプロキシする。インメモリのサービスレジストリで動的ルーティングを実現。

## Core Technologies

- **Language**: Go 1.22+
- **CLI Framework**: [cobra](https://github.com/spf13/cobra) v1.8
- **HTTP**: Go標準ライブラリ (`net/http`, `net/http/httputil`)
- **Tool Management**: mise (`mise.toml`)

## Development Standards

### コードスタイル
- Go標準のコーディング規約に従う（`gofmt`準拠）
- パッケージ公開インターフェースはGoDoc形式でコメントを記述
- エラーは戻り値で伝搬し、センチネルエラー (`errors.New`) を活用

### 並行安全性
- 共有状態へのアクセスは `sync.RWMutex` で保護（読み取り多数を想定）

### テスト
- Go標準の `testing` パッケージを使用
- テストファイルは対象ファイルと同パッケージに配置（`_test.go`サフィックス）

## Development Environment

### Required Tools
- Go 1.22+
- mise（バージョン管理）

### Common Commands
```bash
# Build: go build -o localgate .
# Test:  go test ./...
# Run:   go run . start
```

## CI/CD

GitHub Actions による自動化パイプライン（`.github/workflows/`）:

- **Test** (`test.yml`): 全プッシュで実行。`gofmt` フォーマットチェック → `go test ./...` → `go build ./...`
- **Deploy** (`deploy.yml`): `v*` タグのプッシュで起動。バイナリビルド → GitHub Releases にドラフトリリース作成（`gh release create`）

リリースは `v*` タグをプッシュするだけで自動的にドラフト作成される。

## Key Technical Decisions

- **標準ライブラリ優先**: フレームワークを使わずGoの `net/http/httputil.ReverseProxy` をそのまま利用
- **インメモリレジストリ**: 永続化なし、プロセス再起動でリセット（シンプルさを優先）
- **管理API/プロキシの共存**: 単一ポートでサブドメインの有無により動的に振り分け
