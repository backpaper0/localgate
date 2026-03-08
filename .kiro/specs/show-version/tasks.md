# Implementation Plan

- [x] 1. バージョン情報パッケージの実装
- [x] 1.1 バージョン変数とデフォルト値の定義
  - `internal/version` パッケージを新設し、`Version`・`Commit`・`BuildDate` の 3 つの公開パッケージ変数を宣言する
  - それぞれのデフォルト値は `"dev"`・`"unknown"`・`"unknown"` とする
  - `-ldflags` で上書きできるよう、変数はパッケージスコープで宣言する
  - _Requirements: 2.1, 2.2, 2.3, 2.5, 2.6, 2.7, 3.1_

- [x] 1.2 バージョン情報のフォーマット関数の実装
  - `FormatOutput()` 関数を実装し、4 項目（バージョン・コミット・ビルド日時・Go バージョン）をラベル付きテキスト形式で返す
  - Go バージョンは `runtime.Version()` から取得する
  - 各項目は改行で区切り、末尾に余分な改行を含めない
  - _Requirements: 2.4, 4.1, 4.2, 4.3_

- [x] 2. version サブコマンドの実装
  - `cmd/version.go` に cobra コマンドを定義する（`newVersionCmd()` ファクトリ関数 + `init()` で `rootCmd` に登録）
  - コマンド実行時に `internal/version.FormatOutput()` の結果を標準出力に書き込む
  - 出力には `cmd.OutOrStdout()` を使用してテスト可能にする
  - エラー発生要因がないため `Run` フィールドを使用する（`RunE` 不使用）
  - _Requirements: 1.1, 1.2, 1.3, 3.2_

- [x] 3. (P) CI/CD へのバージョン情報注入の追加
  - `.github/workflows/deploy.yml` のビルドコマンドに `-ldflags` オプションを追加する
  - `Version` に Git タグ（`$GITHUB_REF_NAME`）、`Commit` に `$GITHUB_SHA`、`BuildDate` に UTC の RFC3339 形式（`date -u +%Y-%m-%dT%H:%M:%SZ`）を指定する
  - タスク 1 や 2 と並行して実施可能（異なるファイルを編集）
  - _Requirements: 3.3_

- [x] 4. テストの実装
- [x] 4.1 (P) バージョン情報パッケージのユニットテスト
  - `FormatOutput()` が各変数の値をラベル付きテキストに正しく組み込むことを検証する
  - デフォルト値（`dev` / `unknown`）が出力に含まれることを確認する
  - 出力に `Go Version:` ラベルと `runtime.Version()` の値が含まれることを確認する
  - 出力末尾に余分な改行がないことを確認する
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 4.1, 4.2, 4.3_

- [x] 4.2 (P) version コマンドのコマンドテスト
  - `localgate version` コマンドを実行して標準出力にバージョン情報が表示されることを検証する
  - 任意の値を設定した状態でコマンドを実行し、出力に設定値が含まれることを確認する
  - コマンドが正常終了（エラーなし）することを確認する
  - _Requirements: 1.1, 1.2, 3.2_
