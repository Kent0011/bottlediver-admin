# API.md

`bottlediver-admin` の外部参照用 API ドキュメントです。

## ベース URL

```text
https://<lambda-url>.lambda-url.<region>.on.aws
```

例:

```text
https://xxxxx.lambda-url.ap-northeast-1.on.aws
```

## 共通仕様

- リクエスト/レスポンスの文字コード: `UTF-8`
- JSON リクエストの `Content-Type`: `application/json`
- `id` は UNIX 時間ミリ秒を文字列化した値です
- 一覧 API は新しい順で返します
- データが 0 件の場合は空配列を返します
- `image` は未設定時に省略されることがあります

## 認証

- `GET /news`
- `GET /discography`
- `GET /live`
- `GET /video`

上記は認証不要です。

- `POST`
- `PUT`
- `DELETE`

上記は Basic 認証が必要です。

例:

```http
Authorization: Basic <base64(username:password)>
```

`curl` では `-u 'username:password'` を使えます。

## 共通エラーレスポンス

```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "human readable message",
    "details": []
  },
  "request_id": "<request-id>"
}
```

主なステータスコード:

| HTTP | code | 説明 |
| --- | --- | --- |
| 400 | `INVALID_ARGUMENT` | リクエスト不正 |
| 401 | `UNAUTHENTICATED` | 認証情報なし / 不正 |
| 404 | `NOT_FOUND` | 対象リソースなし |
| 409 | `CONFLICT` | 作成時の競合 |
| 500 | `INTERNAL` | サーバー内部エラー |

## エンドポイント一覧

| Method | Path | 認証 | 説明 |
| --- | --- | --- | --- |
| GET | `/news` | 不要 | News 一覧取得 |
| POST | `/news` | 必須 | News 作成 |
| PUT | `/news/{id}` | 必須 | News 更新 |
| DELETE | `/news/{id}` | 必須 | News 削除 |
| GET | `/discography` | 不要 | Discography 一覧取得 |
| POST | `/discography` | 必須 | Discography 作成 |
| PUT | `/discography/{id}` | 必須 | Discography 更新 |
| DELETE | `/discography/{id}` | 必須 | Discography 削除 |
| GET | `/live` | 不要 | Live 一覧取得 |
| POST | `/live` | 必須 | Live 作成 |
| PUT | `/live/{id}` | 必須 | Live 更新 |
| DELETE | `/live/{id}` | 必須 | Live 削除 |
| GET | `/video` | 不要 | Video 一覧取得 |
| POST | `/video` | 必須 | Video 作成 |
| PUT | `/video/{id}` | 必須 | Video 更新 |
| DELETE | `/video/{id}` | 必須 | Video 削除 |

## News

### GET `/news`

レスポンス:

```json
{
  "items": [
    {
      "id": "1746187200000",
      "title": "2030.01.01 - 1st Album 『hogehoge』リリース決定",
      "image": "https://example.com/image.jpg",
      "content": "ニュース本文を記載"
    }
  ]
}
```

### POST `/news`

リクエスト:

```json
{
  "title": "2030.01.01 - 1st Album 『hogehoge』リリース決定",
  "image": "https://example.com/image.jpg",
  "content": "ニュース本文を記載"
}
```

バリデーション:

- `title`: 1〜200 文字
- `image`: URL 形式、省略可
- `content`: 1〜10000 文字

レスポンス `201`:

```json
{
  "id": "1746187200000",
  "title": "2030.01.01 - 1st Album 『hogehoge』リリース決定",
  "image": "https://example.com/image.jpg",
  "content": "ニュース本文を記載"
}
```

### PUT `/news/{id}`

リクエスト body は `POST /news` と同じです。

レスポンス `200`:

```json
{
  "id": "1746187200000",
  "title": "2030.01.01 - 1st Album 『hogehoge』リリース決定",
  "image": "https://example.com/image.jpg",
  "content": "ニュース本文を記載"
}
```

### DELETE `/news/{id}`

レスポンス `204`:

ボディなし

## Discography

### GET `/discography`

レスポンス:

```json
{
  "items": [
    {
      "id": "1746187200000",
      "title": "1st Album 『Scrawl』",
      "image": "https://example.com/image.jpg",
      "musics": ["一閃", "STROBE", "透明人間"],
      "applemusic_link": "https://music.apple.com/...",
      "spotify_link": "https://open.spotify.com/...",
      "youtubemusic_link": "https://music.youtube.com/...",
      "linemusic_link": "https://music.line.me/...",
      "amazonmusic_link": "https://music.amazon.co.jp/..."
    }
  ]
}
```

### POST `/discography`

リクエスト:

```json
{
  "title": "1st Album 『Scrawl』",
  "image": "https://example.com/image.jpg",
  "musics": ["一閃", "STROBE", "透明人間"],
  "applemusic_link": "https://music.apple.com/...",
  "spotify_link": "https://open.spotify.com/...",
  "youtubemusic_link": "https://music.youtube.com/...",
  "linemusic_link": "https://music.line.me/...",
  "amazonmusic_link": "https://music.amazon.co.jp/..."
}
```

バリデーション:

- `title`: 1〜200 文字
- `image`: URL 形式、省略可
- `musics`: 必須、各要素 1〜200 文字
- `applemusic_link`: 必須、URL 形式
- `spotify_link`: 必須、URL 形式
- `youtubemusic_link`: 必須、URL 形式
- `linemusic_link`: 必須、URL 形式
- `amazonmusic_link`: 必須、URL 形式

レスポンス `201` は `GET /discography` の要素 1 件分と同じです。

### PUT `/discography/{id}`

リクエスト body は `POST /discography` と同じです。

レスポンス `200` は `POST /discography` と同じです。

### DELETE `/discography/{id}`

レスポンス `204`:

ボディなし

## Live

### GET `/live`

レスポンス:

```json
{
  "items": [
    {
      "id": "1746187200000",
      "title": "2026.5.14 - Fireloop presents overplugged",
      "image": "https://example.com/image.jpg",
      "where": "寺田町Fireloop",
      "with": ["band A", "バンドB", "Band 01"],
      "ticket": "ADV ¥2400 / DOOR ¥2400",
      "time": "OPEN: 18:00 / START: 18:30",
      "link": "https://example.com/live-detail"
    }
  ]
}
```

### POST `/live`

リクエスト:

```json
{
  "title": "2026.5.14 - Fireloop presents overplugged",
  "image": "https://example.com/image.jpg",
  "where": "寺田町Fireloop",
  "with": ["band A", "バンドB", "Band 01"],
  "ticket": "ADV ¥2400 / DOOR ¥2400",
  "time": "OPEN: 18:00 / START: 18:30",
  "link": "https://example.com/live-detail"
}
```

バリデーション:

- `title`: 1〜200 文字
- `image`: URL 形式、省略可
- `where`: 1〜200 文字
- `with`: 必須、各要素 1〜200 文字
- `ticket`: 1〜200 文字
- `time`: 1〜100 文字
- `link`: 必須、URL 形式

レスポンス `201` は `GET /live` の要素 1 件分と同じです。

### PUT `/live/{id}`

リクエスト body は `POST /live` と同じです。

レスポンス `200` は `POST /live` と同じです。

### DELETE `/live/{id}`

レスポンス `204`:

ボディなし

## Video

### GET `/video`

レスポンス:

```json
{
  "items": [
    {
      "id": "1746187200000",
      "title": "[Live video] 未明 - bottle diver",
      "link": "https://www.youtube.com/watch?v=..."
    }
  ]
}
```

### POST `/video`

リクエスト:

```json
{
  "title": "[Live video] 未明 - bottle diver",
  "link": "https://www.youtube.com/watch?v=..."
}
```

バリデーション:

- `title`: 1〜200 文字
- `link`: 必須、URL 形式

レスポンス `201`:

```json
{
  "id": "1746187200000",
  "title": "[Live video] 未明 - bottle diver",
  "link": "https://www.youtube.com/watch?v=..."
}
```

### PUT `/video/{id}`

リクエスト body は `POST /video` と同じです。

レスポンス `200` は `POST /video` と同じです。

### DELETE `/video/{id}`

レスポンス `204`:

ボディなし

## 利用例

### 一覧取得

```bash
curl https://YOUR_FUNCTION_URL/news
```

### News 作成

```bash
curl -X POST https://YOUR_FUNCTION_URL/news \
  -H "Content-Type: application/json" \
  -u 'admin:password' \
  -d '{
    "title":"2026.05.02 release info",
    "content":"hello"
  }'
```

### News 更新

```bash
curl -X PUT https://YOUR_FUNCTION_URL/news/1746187200000 \
  -H "Content-Type: application/json" \
  -u 'admin:password' \
  -d '{
    "title":"2026.05.02 release info updated",
    "content":"updated"
  }'
```

### News 削除

```bash
curl -X DELETE https://YOUR_FUNCTION_URL/news/1746187200000 \
  -u 'admin:password'
```
