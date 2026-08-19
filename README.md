# 📚 CLI Base

[![Status](https://img.shields.io/badge/Status-Active-brightgreen)](#)
[![Language](https://img.shields.io/badge/Language-Go-blue)](https://golang.org/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/clibase)](https://golang.org/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/clibase)](https://github.com/shouni/clibase/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**`clibase`** は、Go言語でコマンドラインインターフェース (CLI) アプリケーションを迅速に構築するための、**`spf13/cobra`** ベースの共通基盤を提供するパッケージです。

## ✨ 特徴

* **`cobra` ベース**: 強力なCLI構築ライブラリ **`spf13/cobra`** を基盤としています。
* **構造体による宣言的定義**: `clibase.App` 構造体に必要な要素を渡すだけで、ボイラープレートを排除した綺麗な `main` 関数を実現します。
* **ライフサイクル管理**: `PreRunE` による初期化に加え、`PostRun` による確実なリソース解放（クローズ処理など）を標準サポート。サブコマンドが独自の `PersistentPreRunE`/`PersistentPostRun` を定義していても、root の hook は必ず実行されます。
* **共通フラグの標準提供**: `verbose` (`-V`) と `config` (`-C`) を標準提供。
* **`--version` フラグ**: `App.Version` を指定するだけで cobra 標準の `--version` が有効になります。
* **シグナル対応**: `Execute` は SIGINT/SIGTERM を受け取ると `context` をキャンセルするため、`cmd.Context()` で中断を検知できます。`ExecuteContext` を使えば独自のコンテキストや終了コード制御も可能です。

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

> **注意**: 独自フラグで `-V` / `-C` のショートハンドを再利用すると cobra が起動時にパニックします。`AddFlags` では別のショートハンドを使ってください。

### 内部構造: `clibase.App`

アプリケーションの構成を定義する中心的な構造体です。

```go
type App struct {
    Name string // アプリ名

    Version       string // 指定すると --version フラグが有効になる
    SilenceUsage  bool   // true でエラー時の usage 自動出力を抑制
    SilenceErrors bool   // true で cobra によるエラーメッセージ自動出力を抑制

    AddFlags func(cmd *cobra.Command)                      // フラグ登録フック
    PreRunE  func(cmd *cobra.Command, args []string) error // 実行前チェック/初期化
    PostRun  func(cmd *cobra.Command, args []string)       // 実行後クリーンアップ
    Commands []*cobra.Command                              // サブコマンド群
}

```

### `Execute` と `ExecuteContext`

* `clibase.Execute(app)` — 従来通りの一発実行 API。内部で SIGINT/SIGTERM 監視用の `context` を用意し、エラー時は `os.Exit(1)` します。
* `clibase.ExecuteContext(ctx, app)` — `os.Exit` を呼ばずにエラーを返します。独自のキャンセル文脈を渡したい場合や、終了コードを自分で制御したい場合に使用します。

```go
if err := clibase.ExecuteContext(ctx, app); err != nil {
    // 呼び出し側で終了コードやログ出力を制御できる
    os.Exit(2)
}
```

---

## 📜 ライセンス (License)

このプロジェクトは [MIT License](https://opensource.org/licenses/MIT) の下で公開されています。

---
