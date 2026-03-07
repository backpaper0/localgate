# Requirements Document

## Project Description (Input)
`localgate`コマンドのサブコマンドでサービス登録・解除を行えるようにしてほしい。

localgateサーバーのURLについて:
- `--server`パラメーターが指定されている場合はその値とする
- 環境変数`LOCALGATE_SERVER`が設定されている場合はその値とする
- デフォルトは`http://localhost:9000`とする

## Introduction

`localgate` CLIにサービス登録・解除・一覧表示のサブコマンドを追加する。既存の管理HTTP API（`POST /services`、`DELETE /services/{name}`、`GET /services`）をCLIから呼び出せるようにし、ユーザーがサーバーを再起動することなくサービスを動的に管理できるようにする。

## Requirements

### Requirement 1: サービス登録コマンド

**Objective:** As a 開発者, I want `localgate register <name> <target>` コマンドでサービスを登録できること, so that localgateサーバーを再起動せずに新しいサービスをプロキシ対象として追加できる

#### Acceptance Criteria

1. When `localgate register <name> <target>` が実行された場合, the localgate CLI shall 指定のサーバーURL（後述のサーバーURL解決ルールに従う）に `POST /services` リクエストを送信し、`name` と `target` をJSONボディとして渡す
2. When 登録リクエストが成功（HTTP 201）した場合, the localgate CLI shall 登録完了を示すメッセージを標準出力に表示し、終了コード0で終了する
3. If 登録リクエストがエラー（HTTP 4xx/5xx）を返した場合, the localgate CLI shall サーバーから返されたエラーメッセージを標準エラー出力に表示し、終了コード1で終了する
4. If `<name>` または `<target>` 引数が省略された場合, the localgate CLI shall 使い方（usage）メッセージを標準エラー出力に表示し、終了コード1で終了する
5. If サーバーへの接続が失敗した場合, the localgate CLI shall 接続エラーを標準エラー出力に表示し、終了コード1で終了する

### Requirement 2: サービス解除コマンド

**Objective:** As a 開発者, I want `localgate unregister <name>` コマンドでサービスを解除できること, so that 不要になったサービスをプロキシ対象から動的に削除できる

#### Acceptance Criteria

1. When `localgate unregister <name>` が実行された場合, the localgate CLI shall 指定のサーバーURLに `DELETE /services/{name}` リクエストを送信する
2. When 解除リクエストが成功（HTTP 204）した場合, the localgate CLI shall 解除完了を示すメッセージを標準出力に表示し、終了コード0で終了する
3. If 解除リクエストが404を返した場合（サービスが存在しない）, the localgate CLI shall サービスが見つからない旨のエラーメッセージを標準エラー出力に表示し、終了コード1で終了する
4. If 解除リクエストがその他のエラー（HTTP 4xx/5xx）を返した場合, the localgate CLI shall サーバーから返されたエラーメッセージを標準エラー出力に表示し、終了コード1で終了する
5. If `<name>` 引数が省略された場合, the localgate CLI shall 使い方（usage）メッセージを標準エラー出力に表示し、終了コード1で終了する
6. If サーバーへの接続が失敗した場合, the localgate CLI shall 接続エラーを標準エラー出力に表示し、終了コード1で終了する

### Requirement 3: サービス一覧表示コマンド

**Objective:** As a 開発者, I want `localgate list` コマンドで登録済みサービスの一覧を確認できること, so that 現在localgateに登録されているサービスを素早く把握できる

#### Acceptance Criteria

1. When `localgate list` が実行された場合, the localgate CLI shall 指定のサーバーURLに `GET /services` リクエストを送信する
2. When 一覧取得リクエストが成功（HTTP 200）した場合, the localgate CLI shall 登録済みサービスの名前とターゲットURLを標準出力に一覧表示し、終了コード0で終了する
3. When 登録済みサービスが存在しない場合, the localgate CLI shall サービスが登録されていない旨のメッセージを標準出力に表示し、終了コード0で終了する
4. If 一覧取得リクエストがエラー（HTTP 4xx/5xx）を返した場合, the localgate CLI shall サーバーから返されたエラーメッセージを標準エラー出力に表示し、終了コード1で終了する
5. If サーバーへの接続が失敗した場合, the localgate CLI shall 接続エラーを標準エラー出力に表示し、終了コード1で終了する

### Requirement 4: サーバーURL解決

**Objective:** As a 開発者, I want サーバーURLを柔軟に指定できること, so that ローカルだけでなく任意のlocalgateサーバーに対してCLI操作を行える

#### Acceptance Criteria

1. The localgate CLI shall `register`、`unregister`、および `list` コマンドのすべてで `--server` フラグをサポートする
2. When `--server <url>` フラグが指定された場合, the localgate CLI shall その値をサーバーURLとして使用する
3. When `--server` フラグが指定されておらず環境変数 `LOCALGATE_SERVER` が設定されている場合, the localgate CLI shall `LOCALGATE_SERVER` の値をサーバーURLとして使用する
4. When `--server` フラグも `LOCALGATE_SERVER` 環境変数も指定されていない場合, the localgate CLI shall `http://localhost:9000` をデフォルトのサーバーURLとして使用する
5. The localgate CLI shall `--server` フラグの優先度を `LOCALGATE_SERVER` 環境変数より高く扱う
