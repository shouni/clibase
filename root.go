package clibase

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Config は共通フラグの値を保持します。
// グローバルな Flags 変数を廃止し、この構造体に集約します。
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
		Long:  fmt.Sprintf("%s is a CLI application built with shouni/cli.", app.Name),

		// 共通処理とカスタム処理を統合
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 1. 共通ロジック (例: Verbose モードの反映)
			if GlobalConfig.Verbose {
				// ここで logger のレベル設定などを行う
			}

			// 2. カスタムロジックの実行
			if app.PreRunE != nil {
				return app.PreRunE(cmd, args)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// 共通フラグの登録
	rootCmd.PersistentFlags().BoolVarP(&GlobalConfig.Verbose, "verbose", "V", false, "enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&GlobalConfig.ConfigFile, "config", "C", "", "config file path")

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
		// cobra はデフォルトでエラーを出力するため、ここでは Exit のみ
		os.Exit(1)
	}
}
