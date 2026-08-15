# kokotoba-backend

Kokotoba の API を実装するための Go バックエンドの初期スキャフォールドです。

## ディレクトリ構成

- `cmd/api` 起動点
- `internal/server` HTTP サーバー起動
- `internal/router` ルーティング定義
- `internal/controller` リクエスト処理
- `internal/model` レスポンスなどのデータ定義
- `internal/view` JSON などの出力処理

## いま入っているもの

- `/healthz` のヘルスチェック
- `/api/v1` の API ルート用プレースホルダー
- `kokotoba-infra` の PostgreSQL への共有接続プール

## 起動

```sh
cd ../kokotoba-infra
cp .env.example .env
bash ./up.sh

cd kokotoba-backend
go run ./cmd/api
```

`ADDR` 環境変数で待ち受けアドレスを変更できます。既定値は `:8080` です。
`DATABASE_URL` の既定値は
`postgresql://kokotoba:kokotoba_dev_password@localhost:5432/kokotoba` です。

スキーマやマイグレーションはこのリポジトリには置かず、`kokotoba-infra/postgres/migrations` で一元管理します。
