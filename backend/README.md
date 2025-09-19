# 麻雀 AI gRPC サーバー

このプロジェクトは、Gemini API を使用した麻雀 AI の gRPC サーバーです。クリーンアーキテクチャに基づいて設計されており、同期・ストリーミング・ヘルスチェック機能を提供します。

## アーキテクチャ

```
internal/
├── domain/          # ドメイン層（ビジネスルール）
│   ├── entity/      # エンティティ
│   └── repository/  # リポジトリインターフェース
├── usecase/         # ユースケース層（アプリケーションロジック）
├── infrastructure/  # インフラストラクチャ層（外部サービス）
└── interface/       # インターフェース層（入力/出力）
    ├── config/      # 設定管理
    └── grpc/        # gRPCハンドラー
```

## 環境変数

以下の環境変数を設定してください：

- `GEMINI_API_KEY`: Gemini API キー（必須）
- `GRPC_PORT`: gRPC サーバーのポート（デフォルト: 8080）
- `LOG_LEVEL`: ログレベル（デフォルト: info）

## 使用方法

### 1. 環境変数の設定

```bash
export GEMINI_API_KEY="your-gemini-api-key"
export GRPC_PORT="8080"
export LOG_LEVEL="info"
```

### 2. サーバーの起動

```bash
# ビルド
go build -o bin/server .

# 実行
./bin/server
```

### 3. gRPC クライアントでテスト

grpcurl を使用してテストできます：

```bash
# grpcurlのインストール
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# サービス一覧の確認
grpcurl -plaintext localhost:8080 list

# メソッドの詳細確認
grpcurl -plaintext localhost:8080 describe mahjong.ai.v1.MahjongAIService

# 麻雀AIへの質問（同期）
grpcurl -plaintext -d '{"prompt": "麻雀で最も重要な戦術は何ですか？", "max_tokens": 500, "temperature": 0.7}' \
  localhost:8080 mahjong.ai.v1.MahjongAIService/AskMahjongAI

# 麻雀AIへの質問（ストリーミング）
grpcurl -plaintext -d '{"prompt": "麻雀の基本ルールを教えてください", "max_tokens": 1000, "temperature": 0.5}' \
  localhost:8080 mahjong.ai.v1.MahjongAIService/AskMahjongAIStream

# ヘルスチェック
grpcurl -plaintext -d '{}' \
  localhost:8080 mahjong.ai.v1.MahjongAIService/HealthCheck
```

## 依存関係

- Go 1.24.3+
- Google Generative AI Go SDK
- gRPC
- Logrus（ログ）

## 開発

### ビルド

```bash
go build -o bin/server .
```

### テスト

```bash
go test ./...
```

### 依存関係の更新

```bash
go mod tidy
```

## プロトコルバッファ

プロトファイルは `../proto/proto/mahjong/ai/v1/ai.proto` に定義されており、
生成された Go コードは `proto/gen/go/mahjong/ai/v1/` にあります。

## ライセンス

このプロジェクトは MIT ライセンスの下で公開されています。
