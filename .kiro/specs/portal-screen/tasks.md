# Implementation Plan

## Task 1. ポータルHTTPハンドラの実装

- [x] 1.1 ポータルパッケージを作成し、HTMLテンプレートをポーリング間隔と共に配信するハンドラを実装する
  - `internal/portal` パッケージを新設し、HTMLテンプレートをバイナリに静的埋め込みするハンドラを実装する
  - `html/template` を使い、ポーリング間隔（ミリ秒）をテンプレート変数 `RefreshIntervalMs` として注入する
  - `GET /` に対して `text/html; charset=utf-8` でステータス200を返す
  - テンプレートのパースは起動時に行い、不正なテンプレートはビルド時にエラーとして検出されるようにする
  - _Requirements: 1.1, 1.2_

- [x] 1.2 ポータルHTMLを実装する（サービス一覧・ポーリング・解除操作・モダンデザイン）
  - Tailwind CSS Play CDN をヘッダーで読み込み、オフライン時の基本レイアウト用インラインスタイルをフォールバックとして含める
  - localgate のタイトル／アイコンをヘッダーに配置する
  - ページ読み込み時に `GET /services` を呼び出し、登録済みサービスをカード形式で一覧表示する
  - サービスが0件の場合は空状態メッセージを表示する
  - テンプレート変数 `{{.RefreshIntervalMs}}` を JS の `setInterval` に渡し、定期的にサービス一覧を更新する（ページリロードなし）
  - ポーリング失敗時はエラーバナーを表示し、次の更新タイミングで自動リトライする
  - 各サービスカードに解除ボタンを配置し、クリック時に `confirm()` ダイアログを表示してから `DELETE /services/{name}` を呼び出す
  - 解除成功時はカードをDOMから即時削除し、失敗時はエラーメッセージを表示して一覧を変更しない
  - サービス名はDOM挿入時にXSS対策としてテキストノードで設定する
  - パステルカラー・丸みのあるカード・ホバー時の影アニメーション・レスポンシブレイアウトを実装する
  - _Requirements: 1.1, 1.2, 2.1, 2.2, 3.1, 3.2, 3.3, 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

---

## Task 2. CLIフラグとサーバー設定の拡張

- [x] 2.1 (P) サーバー設定にポーリング間隔フィールドを追加し、管理APIへ渡す
  - `server.ServerConfig` に `PortalRefreshInterval int`（単位: 秒）フィールドを追加する
  - `server.NewProxyServer` が `management.NewAPI` を呼び出す際にこの値を渡すよう修正する
  - Task 1 と並列実装可能（変更ファイルが独立している）
  - _Requirements: 3.1_

- [x] 2.2 `start` サブコマンドに `--portal-refresh` フラグを追加する
  - `cmd/start.go` に `--portal-refresh` フラグ（型 `int`、デフォルト `2`、単位: 秒）を追加する
  - 値が1未満の場合は起動時エラーを返す（既存の `--port` バリデーションと同パターン）
  - `server.ServerConfig.PortalRefreshInterval` に値を設定する（2.1 完了後に実装）
  - _Requirements: 3.1_

---

## Task 3. 管理APIへの統合

- [x] 3.1 管理APIのコンストラクタを変更し、ポータルハンドラを `GET /` に登録する
  - `management.NewAPI` のシグネチャに `refreshIntervalSec int` 引数を追加する
  - `portal.NewHandler(refreshIntervalSec)` を内部でインスタンス化し、`mux.HandleFunc("GET /", ...)` で登録する
  - 既存の `POST /services`・`DELETE /services/{name}`・`GET /services` ハンドラへの影響がないことを確認する
  - Task 1 と Task 2 の完了後に実装する
  - _Requirements: 1.1, 1.2, 1.3, 3.1_

---

## Task 4. テストの実装

- [x] 4.1 (P) ポータルハンドラのユニットテストを実装する
  - `GET /` に対してステータス200・`Content-Type: text/html` が返ることを検証する
  - レスポンスボディが空でないこと（HTMLコンテンツが埋め込まれていること）を確認する
  - `RefreshIntervalMs` の値がレスポンスボディに含まれることを確認する
  - 4.2 と並列実装可能（別パッケージ）
  - _Requirements: 1.1, 1.2_

- [x] 4.2 (P) 管理APIの統合テストを更新する
  - `management.NewAPI` が `refreshIntervalSec` 引数を受け取れることを検証する
  - `GET /` がHTMLを返し、`GET /services` が引き続きJSONを返すことを同一テストスイートで確認する
  - `POST /services`・`DELETE /services/{name}` の既存テストが変更後も通過することを確認する
  - 4.1 と並列実装可能（別パッケージ）
  - _Requirements: 1.3, 2.3_
