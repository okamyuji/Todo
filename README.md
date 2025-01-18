# Todo Analytics Dashboard

## 概要

Todo Analytics Dashboardは、タスク管理とその分析を組み合わせたモダンなWebアプリケーションです。
Goによるバックエンド、Vue.jsによるフロントエンド、そしてTailwind CSSによるスタイリングを採用しています。

## 主な機能

- タスクの作成・管理
- リアルタイムの完了状態トグル
- カテゴリ別・優先度別の分析
- ビジュアルダッシュボード（円グラフ・棒グラフ）
- レスポンシブデザイン

## 技術スタック

- バックエンド
    - Go 1.21以上
    - chi（Webフレームワーク）
    - embed（静的ファイル埋め込み）
- フロントエンド
    - Vue.js 3
    - Chart.js（グラフ描画）
    - Tailwind CSS（スタイリング）

## プロジェクト構造

```shell
.
├── main.go              # メインアプリケーションファイル
├── go.mod              # Goモジュール定義
├── go.sum              # Goモジュールバージョン管理
├── static/             # 静的ファイル
│   └── js/
│       └── app.js      # Vue.jsアプリケーション
└── templates/          # HTMLテンプレート
    └── index.html      # メインページテンプレート
```

## セットアップ方法

### 前提条件

- Go 1.21以上がインストールされていること

### インストール手順

1. リポジトリのクローン

    ```bash
    git clone [リポジトリURL]
    cd todo-analytics
    ```

2. 依存関係のインストール

    ```shell
    go mod download
    ```

3. アプリケーションの起動

    ```shell
    go run main.go
    ```

4. ブラウザでアクセス

    ```shell
    http://localhost:8080
    ```

    ```shell
    GET /api/todos
    ```

    レスポンス

    ```json
    [
        {
            "id": "string",
            "title": "string",
            "category": "string",
            "priority": 1-3,
            "done": boolean,
            "created_at": "datetime",
            "done_at": "datetime|null"
        }
    ]
    ```

    ```shell
    POST /api/todos
    ```

    リクエストボディ

    ```json
    {
        "title": "string",
        "category": "string",
        "priority": 1-3
    }
    ```

    ```shell
    PUT /api/todos/{id}/toggle
    ```

    ```shell
    GET /api/analytics
    ```

    レスポンス

    ```json
    {
        "total_todos": integer,
        "completed_todos": integer,
        "completion_rate": float,
        "average_time": float,
        "category_counts": {
            "category_name": integer
        },
        "priority_counts": {
            "priority_level": integer
        }
    }
    ```

## データモデル

### Todo構造体

```go
type Todo struct {
    ID        string     `json:"id"`
    Title     string     `json:"title"`
    Category  string     `json:"category"`
    Priority  int        `json:"priority"`
    Done      bool       `json:"done"`
    CreatedAt time.Time  `json:"created_at"`
    DoneAt    *time.Time `json:"done_at,omitempty"`
}
```

### Analytics構造体

```go
type Analytics struct {
    TotalTodos      int            `json:"total_todos"`
    CompletedTodos  int            `json:"completed_todos"`
    CompletionRate  float64        `json:"completion_rate"`
    AverageTime     float64        `json:"average_time"`
    CategoryCounts  map[string]int `json:"category_counts"`
    PriorityCounts  map[int]int    `json:"priority_counts"`
}
```

## 開発ガイドライン

### バックエンド開発

- 新しいエンドポイントの追加は`main.go`のルーター設定に行う
- データ構造の変更は`Todo`構造体と`Analytics`構造体を更新
- エラーハンドリングは適切なHTTPステータスコードを返す

### フロントエンド開発

- コンポーネントロジックは`app.js`に集中
- 新しい機能は適切なVueメソッドとして実装
- チャート更新は`updateCharts`メソッドを拡張
- スタイリングはTailwind CSSのユーティリティクラスを使用

## エラーハンドリング

アプリケーションは以下の状況で適切なエラーハンドリングを実装

- API呼び出しの失敗
- データベース操作の失敗
- 不正なリクエストデータ
- 存在しないリソースへのアクセス

## パフォーマンス最適化

- インメモリデータストアの使用
- チャートの効率的な更新
- Vue.jsの算出プロパティ活用
- 静的アセットの適切なキャッシング

## セキュリティ考慮事項

- 入力データのバリデーション
- XSS対策
- CSRF対策
- 適切なHTTPヘッダー設定

## 今後の改善点

1. データの永続化（データベース統合）
2. ユーザー認証の実装
3. タスクの期限日設定機能
4. より詳細な分析機能
5. タスクのフィルタリング機能
6. バッチ処理による定期的なタスクのクリーンアップ

## ライセンス

MITライセンス

## 貢献ガイドライン

1. Issueの作成
2. ブランチの作成（feature/fix）
3. 変更の実装
4. テストの実行
5. プルリクエストの作成

## 開発環境設定

推奨される開発環境の設定:

```shell
# Go環境
go version >= 1.21
go mod vendor

# エディタ設定（VSCode）
- Go拡張機能
- Vue.js拡張機能
- Tailwind CSS IntelliSense
```

## トラブルシューティング

よくある問題と解決方法

1. テンプレートエラー

    ```shell
    panic: template: index.html: function not defined
    ```

    解決策: Vue.jsの構文とGoのテンプレート構文が競合していないか確認

2. 静的ファイルが読み込めない

    ```shell
    404 Not Found: /static/js/app.js
    ```

    解決策: ファイルパスとembedディレクティブを確認
