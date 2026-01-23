package clibase

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Config は共通フラグの値を保持します。
type Config struct {
	Verbose    bool
	ConfigFile string
}

// GlobalConfig はアプリケーション全体で共有される設定です。
var GlobalConfig Config

// App は CLI アプリケーションの構成を定義します。
type App struct {
	Name     string
	AddFlags func(cmd *cobra.Command)
	PreRunE  func(cmd *cobra.Command, args []string) error
	Commands []*cobra.Command
}

// Execute は、アプリケーションの構築と実行をワンストップで行います。
func Execute(app App) {
	rootCmd := &cobra.Command{
		Use:   app.Name,
		Short: fmt.Sprintf("%s CLI tool", app.Name),
		Long:  fmt.Sprintf("%s is a CLI application built with shouni/clibase.", app.Name),

		// 共通処理とカスタム処理を統合
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 1. 共通ロジック (将来的なロギング基盤の初期化などを想定)
			if GlobalConfig.Verbose {
				// 例: log.SetLevel(log.DebugLevel)
			}

			// 2. カスタムロジックの実行
			if app.PreRunE != nil {
				return app.PreRunE(cmd, args)
			}
			return nil
		},
		// 引数なしで実行された場合にヘルプを表示
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				// ヘルプ表示のエラーは致命的ではないが、デバッグのために標準エラー出力へ出す
				fmt.Fprintf(os.Stderr, "Error displaying help: %v\n", err)
			}
		},
	}

	// 共通フラグの登録
	rootCmd.PersistentFlags().BoolVarP(&GlobalConfig.Verbose, "verbose", "V", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&GlobalConfig.ConfigFile, "config", "C", "", "Config file path")

	// アプリ固有フラグの登録
	if app.AddFlags != nil {
		app.AddFlags(rootCmd)
	}

	// サブコマンドの追加
	if len(app.Commands) > 0 {
		rootCmd.AddCommand(app.Commands...)
	}

	// 実行
	if err := rootCmd.Execute(); err != nil {
		// cobra は内部でエラーメッセージを出力するため、ここでは Exit のみ行う
		os.Exit(1)
	}
}
