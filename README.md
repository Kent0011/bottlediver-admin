# bottlediver-admin

`docs/api-design.md` に基づく Go / AWS Lambda / DynamoDB 実装です。

## 環境変数

- `TABLE_NAME`
- `BASIC_AUTH_USERNAME`
- `BASIC_AUTH_PASSWORD`
- `ALLOWED_ORIGINS` (`https://example.com,https://admin.example.com` のようなカンマ区切り)

## ビルド

```bash
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap ./cmd/api
zip -j function.zip bootstrap
```
