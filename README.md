# 📚 CLI Base

**`clibase`** は、Go言語でコマンドラインインターフェース (CLI) アプリケーションを迅速に構築するための、**`spf13/cobra`** ベースの共通基盤を提供するパッケージです。

## ✨ 特徴

* **`cobra` ベース**: 強力なCLI構築ライブラリ **`spf13/cobra`** を基盤としています。
* **構造体による宣言的定義**: `clibase.App` 構造体に必要な要素を渡すだけで、ボイラープレートを排除した綺麗な `main` 関数を実現します。
* **ライフサイクル管理 (NEW!)**: `PreRunE` による実行前チェックに加え、`PostRun` による確実なリソース解放（クローズ処理など）をサポートします。
* **共通フラグの標準提供**: `verbose` (`-V`) と `config` (`-C`) の2つの永続フラグを標準で提供。
* **安全な設定アクセス**: 共通フラグの値は `clibase.GetConfig()` を通じて読み取り専用として安全にアクセス可能です。

---

## 🛠️ インストール

```bash
go get github.com/shouni/clibase

```

---

## 🚀 使用方法

### 1. アプリケーションの構築と実行

`clibase.App` を使って、フラグ、チェックロジック、クリーンアップ、サブコマンドを集約します。

```go
// main.go
package main

import (
    "fmt"
    "os"

    "github.com/shouni/clibase"
    "github.com/spf13/cobra"
)

var customAPIKey string

func main() {
    clibase.Execute(clibase.App{
        Name: "my-app",
        
        // (1) アプリ固有の永続フラグを追加
        AddFlags: func(rootCmd *cobra.Command) {
            rootCmd.PersistentFlags().StringVar(&customAPIKey, "api-key", "", "API Key for authentication")
        },

        // (2) 実行前チェック（共通処理の後に実行されます）
        PreRunE: func(cmd *cobra.Command, args []string) error {
            if customAPIKey == "" && os.Getenv("APP_API_KEY") == "" {
                return fmt.Errorf("--api-key or APP_API_KEY environment variable is required")
            }
            return nil
        },

        // (3) 実行後処理（リソースの解放などに使用。NEW!）
        PostRun: func(cmd *cobra.Command, args []string) {
            fmt.Println("Cleaning up resources...")
        },

        // (4) サブコマンドの追加
        Commands: []*cobra.Command{
            helloCmd,
        },
    })
}

var helloCmd = &cobra.Command{
    Use:   "hello",
    Short: "Prints a greeting",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Hello!")
        
        // 共通フラグの値をゲッターから取得
        cfg := clibase.GetConfig()
        if cfg.Verbose {
            fmt.Printf("Debug: Config file is %s\n", cfg.ConfigFile)
        }
    },
}

```

---

## ⚙️ 提供される機能

### 共通フラグ

すべてのコマンドで以下のフラグが自動的に利用可能になります。

| フラグ | 短縮形 | 用途 | アクセス方法 |
| --- | --- | --- | --- |
| `--verbose` | `-V` | 詳細ログの有効化 | `clibase.GetConfig().Verbose` |
| `--config` | `-C` | 設定ファイルのパス指定 | `clibase.GetConfig().ConfigFile` |

### 内部構造: `clibase.App`

アプリケーションの構成を定義する中心的な構造体です。

```go
type App struct {
    Name     string                                    // アプリ名
    AddFlags func(cmd *cobra.Command)                  // フラグ登録フック
    PreRunE  func(cmd *cobra.Command, args []string) error // 実行前チェック/初期化
    PostRun  func(cmd *cobra.Command, args []string)       // 実行後クリーンアップ
    Commands []*cobra.Command                          // サブコマンド群
}

```

---

## 📜 ライセンス (License)

このプロジェクトは [MIT License](https://opensource.org/licenses/MIT) の下で公開されています。

---
