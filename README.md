# pokemon-silhouette-quiz

## 概要
PokeAPI を利用し、選択した地方(世代)とメガシンカ/ゲンシカイキを含める設定でポケモンをランダム出題し、シルエット画像から名称を当てる Web アプリ。フロントエンドはReact、バックエンドは Go + chi。

## ディレクトリ構成
```
backend/                Go API サーバ
	cmd/server/main.go    エントリポイント
	internal/api          ルーティング+ハンドラ
	internal/poke         PokeAPIクライアント / 画像シルエット処理 / 地方定義
	internal/quiz         セッション・ロジック
frontend-react/         React + TypeScript + Vite
```

## エンドポイント
- `GET  /health` ヘルスチェック
- `POST /api/quiz/start` Body: `{regions:["kanto",...], allowMega:boolean, allowPrimal:boolean}` -> `{sessionId}`
  - メガシンカ・ゲンシカイキ対応、地域フォーム（アローラ・ガラル等）フィルタ対応
- `POST /api/quiz/guess` Body: `{sessionId, answer}` -> `{correct, solved, retryAfter}` (5秒制限あり)
- `POST /api/quiz/giveup` Body: `{sessionId}` -> `{pokemonId, name, types, region}`
- `GET  /api/quiz/silhouette/{sessionId}` セッション対応シルエット PNG
- `GET  /api/quiz/artwork/{sessionId}` 結果用カラーアートワーク PNG (クリア/ギブアップ後のみ)
- `GET  /api/quiz/hint/{sessionId}` ヒント情報 -> `{types:["ほのお",...], region:"カントー", firstLetter:"フ"}`
- `GET  /api/quiz/search?prefix=フシ` 名前補完候補 -> `["フシギダネ", "フシギソウ", ...]` (日本語優先) 

## セットアップ
### Backend
前提: Go 1.22+
```
cd backend
go build ./...
go run ./cmd/server
```
デフォルトで :8080 で待ち受け。

### Frontend React
```
cd frontend-react
npm install
npm run dev
```
