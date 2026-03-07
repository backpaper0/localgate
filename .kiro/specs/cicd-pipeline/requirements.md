# Requirements Document

## Project Description (Input)
GitHub Actionsを使用してCI/CDパイプラインを構築する。

次の2つのパイプラインを構築すること。

**テストパイプライン**
- トリガー: pushされたとき
- 実施すること:
    - ソースコードがフォーマットされていることのチェック
    - すべてのテストの実行
    - バイナリのビルド

**デプロイパイプライン**
- トリガー: `v<バージョン番号>`タグがpushされたとき
- 実施すること:
    - バイナリのビルド
    - ドラフト状態のリリースを作成

## Requirements
<!-- Will be generated in /kiro:spec-requirements phase -->

