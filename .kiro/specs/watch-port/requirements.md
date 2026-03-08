# 要件定義書

## Project Description (Input)
ポート監視と自動登録が行えるサブコマンドを追加したい。

`watch`サブコマンドを実行するとポート監視を開始する。
新規ポートでLISTENが開始されると自動でサービスを登録したい。
ポートがLISTENを終了すると自動でサービスを解除したい。
`watch`コマンドが終了するとき、残っているサービスをすべて解除したい。

## はじめに

`watch` サブコマンドは、ホストOS上のLISTENポートを継続的に監視し、新たにLISTENが開始されたポートを localgate サーバへ自動登録、LISTENが終了したポートを自動解除する機能を提供する。これにより、開発者はサービスの起動・停止を手動で管理することなく、localgate を通じて自動的にサービスにアクセスできるようになる。

## Requirements

### Requirement 1: watchサブコマンドの起動

**Objective:** 開発者として、`localgate watch` コマンドを実行することで、ポートの自動監視・登録・解除を開始したい。それにより、localgate へのサービス登録作業を手動で行わずに済むようにしたい。

#### Acceptance Criteria

1. The localgate shall provide a `watch` subcommand that starts port monitoring when invoked.
2. When `watch` コマンドが起動されると, the localgate shall localgate サーバへの接続を確認し、接続に成功した場合のみ監視を開始する。
3. If localgate サーバへの接続に失敗した場合, the localgate shall エラーメッセージを標準エラー出力に表示し、ゼロ以外の終了コードで終了する。
4. The localgate shall `--server` フラグおよび `$LOCALGATE_SERVER` 環境変数、デフォルト値 `http://localhost:9000` の優先順位でサーバーURLを解決する。
5. While 監視中, the localgate shall 監視状態であることを示すログを標準出力に出力し続ける。

---

### Requirement 2: LISTENポートの定期検出

**Objective:** 開発者として、稼働中のLISTENポートの変化をリアルタイムに近い形で検出したい。それにより、サービスの起動を素早く localgate に反映させたい。

#### Acceptance Criteria

1. While 監視中, the localgate shall 設定されたポーリング間隔ごとにホスト上のTCPv4/v6 LISTENポートの一覧を取得する。
2. The localgate shall デフォルトのポーリング間隔として1秒を使用する。
3. The localgate shall `--interval` フラグで秒単位のポーリング間隔を上書きできるようにする。
4. If ポート一覧の取得に失敗した場合, the localgate shall エラーをログに記録し、次のポーリングサイクルまで処理を継続する。

---

### Requirement 3: 新規LISTENポートの自動登録

**Objective:** 開発者として、新たにLISTENを開始したポートが自動的に localgate に登録されることを望む。それにより、サービス起動後すぐにサブドメイン経由でアクセスできるようにしたい。

#### Acceptance Criteria

1. When 前回ポーリング時には存在せず、今回ポーリング時に新たに検出されたLISTENポートがある場合, the localgate shall そのポートを対象として localgate サーバの管理APIへサービス登録リクエストを送信する。
2. When ポートの自動登録が成功した場合, the localgate shall 登録されたサービス名とポート番号を標準出力にログとして出力する。
3. If ポートの登録リクエストが失敗した場合, the localgate shall エラーをログに記録し、そのポートを未登録状態のまま次のサイクルで再試行する。
4. The localgate shall 自動登録時のサービス名を `port-{ポート番号}` の形式で生成する。
5. The localgate shall watchコマンドが自動登録したサービスの一覧を内部で保持する。

---

### Requirement 4: LISTENポート終了時の自動解除

**Objective:** 開発者として、LISTENが終了したポートのサービスが自動的に localgate から解除されることを望む。それにより、廃止されたサービスへのルーティングが自動的にクリアされるようにしたい。

#### Acceptance Criteria

1. When 前回ポーリング時には存在し、今回ポーリング時に消滅したLISTENポートがある場合, the localgate shall そのポートに対応するサービスの解除リクエストを localgate サーバの管理APIへ送信する。
2. When ポートの自動解除が成功した場合, the localgate shall 解除されたサービス名とポート番号を標準出力にログとして出力する。
3. If ポートの解除リクエストが失敗した場合, the localgate shall エラーをログに記録し、そのサービスを登録済みのまま次のサイクルで再試行する。
4. The localgate shall watchコマンドが登録したサービスのみを解除対象とし、他の手段で登録されたサービスには干渉しない。

---

### Requirement 5: watchコマンド終了時のクリーンアップ

**Objective:** 開発者として、`watch` コマンドを終了させたとき、監視中に登録されたすべてのサービスが自動的に解除されることを望む。それにより、watch 終了後に不要なルーティングエントリが残留しないようにしたい。

#### Acceptance Criteria

1. When `watch` コマンドがシグナル（SIGINT / SIGTERM）を受信した場合, the localgate shall 正常終了シーケンスを開始する。
2. While 正常終了シーケンス中, the localgate shall watchコマンドが自動登録したすべてのサービスを localgate サーバの管理APIへ解除リクエストを送信する。
3. When すべてのサービスの解除が完了した場合, the localgate shall 「クリーンアップ完了」を示すメッセージを出力し、ゼロの終了コードで終了する。
4. If 一部のサービスの解除に失敗した場合, the localgate shall 失敗したサービス名をエラーとして出力し、残りのサービスの解除を継続する。
5. The localgate shall クリーンアップの完了を待たずに強制終了（SIGKILL等）が発生した場合を除き、必ず登録済みサービスを解除してから終了する。
