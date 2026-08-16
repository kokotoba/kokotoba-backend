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
- Firebase Authentication の ID トークン検証
- ローカル開発用の認証バイパス
- `/api/v1/me/settings` のユーザー設定取得・部分更新
- `/api/v1/me/phrases` のよく使う文章の一覧取得・登録
- `/api/v1/me/phrases/order` の表示順保存
- `/api/v1/me/phrases/{phraseID}` のよく使う文章の削除
- `kokotoba-infra` の PostgreSQL への共有接続プール

## 起動

```sh
cd ../kokotoba-infra
cp .env.example .env
bash ./up.sh

cd ../kokotoba-backend
cp .env.example .env
set -a
source .env
set +a
go run ./cmd/api
```

Go 1.25 以上が必要です。
`ADDR` 環境変数で待ち受けアドレスを変更できます。既定値は `:8080` です。
`DATABASE_URL` の既定値は
`postgresql://kokotoba:kokotoba_dev_password@localhost:5432/kokotoba` です。

このアプリケーションは `.env` を自動では読み込みません。起動前に上記のように
環境変数へ読み込んでください。

スキーマやマイグレーションはこのリポジトリには置かず、`kokotoba-infra/postgres/migrations` で一元管理します。
ユーザー本体は `users`、ユーザーごとの設定は `user_settings`、
よく使う文章は `frequent_phrases` テーブルに保存します。

## 認証

`AUTH_MODE=firebase` が既定値です。クライアントがFirebase Authenticationで取得した
IDトークンを、次の形式でAPIへ送信します。

```text
Authorization: Bearer <Firebase ID token>
```

FirebaseモードではADCを使用します。ローカルではサービスアカウントJSONを
リポジトリ外へ保存し、次の環境変数を設定します。

```sh
export APP_ENV=development
export AUTH_MODE=firebase
export FIREBASE_PROJECT_ID=your-firebase-project-id
export GOOGLE_APPLICATION_CREDENTIALS=/absolute/path/to/service-account.json
```

`FIREBASE_PROJECT_ID` は、サービスアカウントまたは実行環境からプロジェクトIDを
取得できる場合は省略できます。サービスアカウントJSONはGitへコミットしないでください。

ローカル開発では `.env.example` の設定により、Firebaseトークンなしでデモユーザー
（ID `1`）としてアクセスできます。Devモードは `APP_ENV=development` の場合だけ
有効です。存在しない `DEV_USER_ID` を指定した場合や、開発環境以外でDevモードを
指定した場合は起動に失敗します。

デモユーザーの設定確認:

```sh
curl http://localhost:8080/api/v1/me/settings
```

設定は変更する項目だけを送信します。

```sh
curl -X PATCH http://localhost:8080/api/v1/me/settings \
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
curl http://localhost:8080/api/v1/me/phrases

curl -X POST http://localhost:8080/api/v1/me/phrases \
  -H 'Content-Type: application/json' \
  -d '{"text":"ゆっくり話してください"}'

curl -X DELETE http://localhost:8080/api/v1/me/phrases/1
```

表示順は現在登録されている全文章のIDを、表示したい順で送信します。

```sh
curl -X PUT http://localhost:8080/api/v1/me/phrases/order \
  -H 'Content-Type: application/json' \
  -d '{"phrase_ids":[3,1,2]}'
```
