# 要件ドキュメント

## プロジェクト説明（入力）
次のような構成を取るコンテナ環境がある。

```
docker network create --internal=true prv_net
docker network create pub_net
docker run -d --name=lg --hostname=localgate.test -p 9000:9000 --network=prv_net localgate
docker network connect pub_net lg
docker run -d --name=foobar --network=prv_net foobarservice
```

foobarコンテナ内からサービスを登録するため`http://localgate.test:9000/services`に対してリクエストを送信すると、`localgate`サービスへのプロキシだと判断されて"service not found"となる。

この課題を解決してほしい。

## 問題分析

現在の実装では、Hostヘッダのサブドメインの有無でプロキシ/管理APIを判定している。コンテナ環境で`localgate.test`というホスト名でアクセスした場合、`localgate`がサブドメインと誤認され、プロキシルーティングに分岐してしまう。

## 要件

### Requirement 1: 自己ホスト名による管理APIルーティング

**Objective:** サービス管理者として、localgate自身のホスト名（例: `localgate.test`）を使ってアクセスした際に管理APIが動作することを期待する。そうすることで、コンテナ環境においてもサービスの登録・管理が正常に行える。

#### 受け入れ基準
1. When Hostヘッダがlocalgate自身のホスト名（例: `localgate.test`）に一致するリクエストを受信したとき、the localgate shall そのリクエストをプロキシではなく管理APIとして処理する。
2. When Hostヘッダが`<サブドメイン>.<自己ホスト名>`の形式（例: `foo.localgate.test`）であるとき、the localgate shall サブドメイン部分（`foo`）に対応するサービスにプロキシする。
3. If リクエストのHostヘッダがlocalgate自身のホスト名に一致し、かつ対応する管理APIパスが存在しない場合、the localgate shall 404 Not Foundを返す。
4. The localgate shall `localhost`（ポートなし、あるいはポート付き）をHostヘッダに持つリクエストを、引き続き管理APIとして処理する（後方互換性の維持）。

### Requirement 2: 自己ホスト名の設定

**Objective:** サービス管理者として、localgate起動時に自身のホスト名を設定できることを期待する。そうすることで、異なるコンテナ・デプロイ環境に対応できる。

#### 受け入れ基準
1. When `--hostname`（または等価の設定）フラグを指定してlocalgate startを実行したとき、the localgate shall 指定されたホスト名を自己ホスト名として使用する。
2. If `--hostname`フラグが指定されていない場合、the localgate shall デフォルト値として`localhost`を使用する（現状の動作を維持する）。
3. The localgate shall 設定されたホスト名をHostヘッダの比較に用い、大文字・小文字を区別しない。
4. The localgate shall ポート番号付きのHostヘッダ（例: `localgate.test:9000`）でも自己ホスト名と正しく照合できる。

REVIEW: ホスト側からはlocalhost:9000でアクセスするので、上手く動かないように思う。

### Requirement 3: コンテナ環境での動作互換性

**Objective:** インフラ担当者として、Docker等のコンテナ環境でlocalgateをデプロイした際に管理APIとプロキシが正常に動作することを期待する。そうすることで、コンテナを使ったローカル開発ワークフローを阻害しない。

#### 受け入れ基準
1. While localgateが`--hostname=localgate.test`で起動しているとき、the localgate shall `http://localgate.test:9000/services`へのPOSTリクエストをサービス登録リクエストとして受け付ける。
2. While localgateが`--hostname=localgate.test`で起動しているとき、the localgate shall `http://localgate.test:9000/services`へのGETリクエストをサービス一覧取得として応答する。
3. If foobarサービスが登録済みで`http://foobar.localgate.test:9000/`にリクエストが送信された場合、the localgate shall 登録済みのfoobarサービスへプロキシする。
4. The localgate shall 既存の`localhost`ベースのルーティング動作を変更しない。
