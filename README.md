ポケモン シルエットクイズ
===========================

概要
----
PokeAPI を利用し、選択した地方(世代)とメガシンカ/ゲンシカイキを含める設定でポケモンをランダム出題し、シルエット画像から名称を当てる Web アプリ。React 版と Vue 版の 2 種フロントを同居させています。バックエンドは Go + chi。

現状の注意
------------
サンプル実装段階: ヒント(タイプ/地方/最初の文字)エンドポイントや正式なオートコンプリート検索 API は最小限/未実装のダミーです。必要であれば追加実装指示ください。

ディレクトリ構成
------------------
```
backend/                Go API サーバ
	cmd/server/main.go    エントリポイント
	internal/api          ルーティング+ハンドラ
	internal/poke         PokeAPIクライアント / 画像シルエット処理 / 地方定義
	internal/quiz         セッション・ロジック
frontend-react/         React + TypeScript + Vite
frontend-vue/           Vue 3 + TypeScript + Vite
```

エンドポイント(暫定)
---------------------
- `GET  /health` ヘルスチェック
- `POST /api/quiz/start` Body: `{regions:["kanto",...], allowMega:boolean, allowPrimal:boolean}` -> `{sessionId}`
- `POST /api/quiz/guess` Body: `{sessionId, answer}` -> `{correct, solved, retryAfter}` (5秒制限あり)
- `POST /api/quiz/giveup` Body: `{sessionId}` -> `{pokemonId, name, types, region}`
- `GET  /api/quiz/silhouette/{id}` シルエット PNG (現状は start で選んだ ID をクライアント側に保持してないためダミー 1 固定参照部分あり)

セットアップ (Backend)
-----------------------
前提: Go 1.22+
```
cd backend
go build ./...
go run ./cmd/server
```
デフォルトで :8080 で待ち受け。

セットアップ (Frontend React)
------------------------------
```
cd frontend-react
npm install
npm run dev
```
Vite dev サーバ (通常は :5173) から API へは開発時 CORS を利用。必要に応じて `vite.config.ts` でプロキシ追加。

セットアップ (Frontend Vue)
----------------------------
```
cd frontend-vue
npm install
npm run dev
```

改善 TODO アイデア
-------------------
1. start API が返す `pokemonId` をセッション開始時に返却しない (チート防止) ため silhouette 取得は sessionId 経由の署名付き /api/quiz/silhouette?session=... に変更する。
2. `/api/quiz/hint` 追加しタイプ・地方・最初の文字(1〜N文字段階的)を返す。
3. `/api/quiz/search?prefix=pi` で該当ポケモン名一覧を返す(キャッシュ)しオートコンプリートに利用。
4. メガシンカ/ゲンシカイキ対応: PokeAPI 上の形態 (forms) や別 ID をマッピングして出題 pool に統合。
5. 正解時にカラー画像を表示するための `/api/quiz/reveal?session=` 実装。
6. セッション永続化 (Redis など) とレート制限。
7. UI: シルエット表示時のローディング、アニメーション、アクセシビリティ改善。

ライセンス
----------
本リポジトリは学習/デモ目的。PokeAPI 利用規約に従ってください。ポケモン名称・画像等は (C) Nintendo / Creatures Inc. / GAME FREAK inc.

# pokemon-silhouette-quiz