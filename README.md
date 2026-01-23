# 📚 CLI Base

**`clibase`** は、Go言語でコマンドラインインターフェース (CLI) アプリケーションを迅速に構築するための、**`spf13/cobra`** ベースの共通基盤を提供するパッケージです。

## ✨ 特徴

* **`cobra`** ベース: 強力なCLI構築ライブラリ **`spf13/cobra`** を基盤としています。
* **柔軟なカスタマイズ (NEW\!)**: アプリケーション固有の永続フラグ定義や、**エラーを返す実行前チェック**（例: 環境変数チェック）をコールバックを通じて注入できます。
* **共通フラグの提供**: `verbose` (`-v`) と `config` (`-c`) の2つの永続フラグを標準で提供します。
* **初期化処理の抽象化**: 全てのコマンド実行前に共通処理（詳細モードの有効化、設定ファイルの読み込み）と、アプリケーション固有のチェックを安全に実行するための構造を提供します。
* **シンプルなエントリポイント**: アプリケーションの実行を簡潔な `Execute` 関数にカプセル化します。

-----

## 🛠️ インストール

標準のGoコマンドでインストールできます。

```bash
go get github.com/shouni/go-cli-base
```

-----

## 🚀 使用方法

`clibase`の新しい`Execute`関数は、以下のシグネチャを持ちます。

```go
func Execute(appName string, addFlags CustomFlagFunc, preRunE CustomPreRunEFunc, cmds ...*cobra.Command)
```

### 1\. ルートコマンドの初期化と実行

アプリケーション固有のフラグ追加と実行前チェックのロジックを定義し、`clibase.Execute`に渡してCLIを起動します。

```go
// main.go
package main

import (
    "fmt"
    "os"

    "github.com/shouni/clibase"
    "github.com/spf13/cobra"
)

var (
    // アプリ固有のグローバル変数
    customAPIKey string 
)

// (1) アプリケーション固有の永続フラグを追加する関数
func addAppFlags(rootCmd *cobra.Command) {
    rootCmd.PersistentFlags().StringVar(&customAPIKey, "api-key", "", "Custom API key for the app.")
}

// (2) アプリケーション固有の実行前チェックを行う関数
func preRunAppE(cmd *cobra.Command, args []string) error {
    if customAPIKey == "" && os.Getenv("APP_API_KEY") == "" {
        return fmt.Errorf("エラー: --api-key または環境変数 APP_API_KEY が必須です")
    }
    return nil
}

// アプリケーション固有のサブコマンドを定義
var helloCmd = &cobra.Command{
    Use:   "hello",
    Short: "Prints a greeting message.",
    Run: func(cmd *cobra.Command, args []string) {
       fmt.Println("Hello from the CLI app!")
       // 共通フラグの値にアクセス
       if clibase.Flags.Verbose {
          fmt.Printf("Custom API Key: %s\n", customAPIKey)
       }
    },
}

func main() {
    // アプリケーション名とコールバック、サブコマンドを渡して実行
    clibase.Execute("my-awesome-cli", addAppFlags, preRunAppE, helloCmd)
}
```

### 2\. 共通フラグへのアクセス

定義された共通フラグの値は、**`clibase.Flags`** グローバル変数を通じて、アプリケーションのどこからでもアクセス可能です。（変更なし）

-----

### 3\. 提供される共通機能とフック

`clibase.Execute` によって起動されるルートコマンドには、以下の機能が組み込まれています。

| 機能 | 実行タイミング | カスタマイズ方法 |
| :--- | :--- | :--- |
| **共通フラグ** (`-v`, `-c`) | `NewRootCmd`内で定義 | `clibase.Flags`でアクセス |
| **`clibase`ロジック** | `PersistentPreRunE`内（最初） | `clibase.Flags.Verbose`による初期ロギング設定など。 |
| **アプリ固有のチェック** | `PersistentPreRunE`内（`clibase`処理後） | **`preRunE CustomPreRunEFunc`** を通じて注入。ここでエラーを返すことが可能。 |

-----

## ⚙️ 内部構造

`go-cli-base` パッケージの主要なコンポーネントは以下の通りです。

### コールバック関数型 (NEW\!)

アプリケーション固有の処理を注入するための型です。

```go
// アプリ固有の永続フラグを追加
type CustomFlagFunc func(rootCmd *cobra.Command)

// アプリ固有のエラーを返す実行前チェック
type CustomPreRunEFunc func(cmd *cobra.Command, args []string) error
```

### `NewRootCmd` 関数

アプリケーション名、**カスタムフラグ関数**、**実行前チェック関数**を受け取り、ルートコマンドを生成します。ロジックの複雑な結合は\*\*`createPreRunE`\*\*ヘルパー関数に分離されています。

C## `Execute` 関数

CLIアプリケーションのエントリポイントとして機能し、全てのカスタマイズ関数をルートコマンドに渡し、`rootCmd.Execute()`を呼び出します。エラー処理は`os.Exit(1)`のみです。

-----

### 📜 ライセンス (License)

このプロジェクトは [MIT License](https://opensource.org/licenses/MIT) の下で公開されています。

