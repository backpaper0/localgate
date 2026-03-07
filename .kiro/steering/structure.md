# Project Structure

## Organization Philosophy

Go標準のプロジェクトレイアウト。CLIエントリポイント (`cmd/`)、公開不可の内部実装 (`internal/`)、最小限のルートエントリポイントの3層構成。

## Directory Patterns

### CLIコマンド
**Location**: `cmd/`
**Purpose**: cobraコマンド定義。各ファイルが1コマンドに対応。
**Example**: `cmd/start.go` → `localgate start` サブコマンド

### 内部パッケージ
**Location**: `internal/`
**Purpose**: 外部公開しない実装。ドメインごとにサブパッケージに分割。

| パッケージ | 役割 |
|---|---|
| `internal/registry` | サービスのルーティングテーブル管理（登録・解除・検索） |
| `internal/proxy` | リバースプロキシ転送、Hostヘッダからのサブドメイン抽出 |
| `internal/management` | 管理HTTP API（サービスのCRUD） |
| `internal/server` | HTTPサーバ本体、プロキシ/管理APIへのルーティング判定。`ServerConfig.Hostname` で追加の自己ホスト名を指定可能（コンテナ環境対応） |

### エントリポイント
**Location**: `main.go`
**Purpose**: `cmd.Execute()` を呼ぶだけのシンプルなエントリポイント

## Naming Conventions

- **Files**: snake_case（Go標準）、テストは `xxx_test.go`
- **Types**: PascalCase（インターフェース含む）
- **Functions/Methods**: PascalCase（公開）/ camelCase（非公開）
- **Errors**: センチネルエラーは `ErrXxx` 形式（例: `ErrNotFound`）

## Code Organization Principles

- **インターフェース定義**: 各パッケージが自身のインターフェースを定義し、依存を疎結合に保つ
  - 例: `registry.ServiceRegistry`、`proxy.Handler`
- **依存方向**: `server` → `registry`, `proxy`, `management`（内側に向かう）
- **コンストラクタパターン**: `NewXxx()` 関数でインターフェースを返す（実装型を隠蔽）

## Routing Decision Pattern

`internal/server` のルーティング判定順序:
1. Hostヘッダからポートを除去し小文字正規化
2. `selfHostnames`（`localhost` + `--hostname` 指定値）に一致 → 管理API
3. サブドメインなし → 管理API
4. サブドメインあり → レジストリ検索 → プロキシ or 404
