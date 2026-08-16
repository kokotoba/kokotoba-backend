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
- `/api/v1/users/{userID}/settings` のユーザー設定取得
- `PATCH /api/v1/users/{userID}/settings` のユーザー設定部分更新
- `/api/v1/users/{userID}/phrases` のよく使う文章の一覧取得・登録
- `/api/v1/users/{userID}/phrases/order` の表示順保存
- `/api/v1/users/{userID}/phrases/{phraseID}` のよく使う文章の削除
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
ユーザー本体は `users`、ユーザーごとの設定は `user_settings`、
よく使う文章は `frequent_phrases` テーブルに保存します。

デモユーザー（ID `1`）の設定確認:

```sh
curl http://localhost:8080/api/v1/users/1/settings
```

設定は変更する項目だけを送信します。

```sh
curl -X PATCH http://localhost:8080/api/v1/users/1/settings \
  -H 'Content-Type: application/json' \
  -d '{"speech_volume":70,"use_history_for_suggestions":false}'
```

更新できる項目は以下です。

- `text_size`, `button_size`, `contrast`, `suggestion_count`
- `speech_rate`, `speech_volume`, `speech_voice`
- `use_history_for_suggestions`, `use_location_for_suggestions`
- `use_profile_for_suggestions`, `show_confirmation_after_selection`
- `save_conversation_history`, `allow_external_communication`

よく使う文章の一覧取得・登録・削除:

```sh
curl http://localhost:8080/api/v1/users/1/phrases

curl -X POST http://localhost:8080/api/v1/users/1/phrases \
  -H 'Content-Type: application/json' \
  -d '{"text":"ゆっくり話してください"}'

curl -X DELETE http://localhost:8080/api/v1/users/1/phrases/1
```

表示順は現在登録されている全文章のIDを、表示したい順で送信します。

```sh
curl -X PUT http://localhost:8080/api/v1/users/1/phrases/order \
  -H 'Content-Type: application/json' \
  -d '{"phrase_ids":[3,1,2]}'
```
