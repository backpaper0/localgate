# Requirements Document

## Project Description (Input)
バージョンを表示するサブコマンドを追加する。

```
localgate version
```

localgateのバージョンとコミットハッシュを出力する。

他にも出力した方が良い項目があれば提案してほしい。なければないでOK。

## はじめに

`localgate version` サブコマンドは、バイナリのバージョン情報を標準出力に表示する機能を提供する。バージョン番号・コミットハッシュに加え、ビルド日時・Goバージョンも表示することを提案する（詳細は各要件の注記を参照）。

## Requirements

### Requirement 1: バージョンサブコマンドの提供

**Objective:** As a localgate ユーザー, I want `localgate version` コマンドを実行してバージョン情報を確認できる, so that 使用中のバイナリのバージョンを素早く把握できる

#### Acceptance Criteria

1. When `localgate version` コマンドが実行される, the localgate CLI shall バージョン情報を標準出力に表示する
2. When `localgate version` コマンドが実行される, the localgate CLI shall 正常終了（終了コード 0）する
3. The localgate CLI shall `version` サブコマンドをcobraコマンドとして `cmd/version.go` に実装する

---

### Requirement 2: バージョン情報の表示内容

**Objective:** As a localgate ユーザー, I want バイナリのバージョン・コミットハッシュを確認できる, so that 動作中のバイナリの出所を特定できる

#### Acceptance Criteria

1. The localgate CLI shall バージョン番号（例: `v1.2.3`）を表示する
2. The localgate CLI shall ビルド時のGitコミットハッシュを表示する
3. The localgate CLI shall ビルド日時を表示する
4. The localgate CLI shall ビルドに使用したGoのバージョンを表示する
5. If バージョン番号がビルド時に埋め込まれていない, the localgate CLI shall `dev` というデフォルト値を表示する
6. If コミットハッシュがビルド時に埋め込まれていない, the localgate CLI shall `unknown` というデフォルト値を表示する
7. If ビルド日時がビルド時に埋め込まれていない, the localgate CLI shall `unknown` というデフォルト値を表示する

---

### Requirement 3: ビルド時のバージョン情報埋め込み

**Objective:** As a localgate 開発者, I want ビルド時にバージョン情報をバイナリへ注入できる, so that リリースバイナリが正確なバージョン情報を持てる

#### Acceptance Criteria

1. The localgate CLI shall `-ldflags` でバージョン番号・コミットハッシュ・ビルド日時を上書きできる変数を公開する
2. When `go build -ldflags "-X ..." -o localgate .` でビルドされる, the localgate CLI shall 指定した値をバージョン情報として表示する
3. The localgate CLI shall CI/CDの `deploy.yml` ワークフローにおいて、Gitタグとコミットハッシュを自動的にビルドフラグとして渡すようにする

---

### Requirement 4: 出力フォーマット

**Objective:** As a localgate ユーザー, I want バージョン情報が読みやすい形式で表示される, so that 必要な情報を素早く確認できる

#### Acceptance Criteria

1. The localgate CLI shall バージョン情報をラベル付きのテキスト形式（例: `Version: v1.2.3`）で表示する
2. The localgate CLI shall 各情報項目を改行で区切って表示する
3. The localgate CLI shall 出力の末尾に余分な空行を含まない
