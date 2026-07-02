# sample-go
タクシーをボタン一発で呼び出すWebページサンプル

## セットアップ

### 前提条件

- [Go](https://golang.org/dl/) 1.24 以上

### インストール

```bash
git clone https://github.com/nksato/sample-go.git
cd sample-go
```

### 環境変数

Google マップを使用するには、API キーを環境変数に設定してください。

```bash
export GOOGLE_MAPS_API_KEY=your_api_key_here
```

> API キーを設定しない場合でも、マップ機能が制限された状態でサーバーは起動します。

### サーバーの起動

```bash
go run .
```

サーバーが起動したら、ブラウザで [http://localhost:8080](http://localhost:8080) にアクセスしてください。

## その他のコマンド

```bash
# ビルド
go build ./...

# テスト
go test ./...

# 静的解析
go vet ./...
```
