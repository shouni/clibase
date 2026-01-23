package clibase

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Config は共通フラグの値を保持する内部構造体です。
type Config struct {
	Verbose    bool
	ConfigFile string
}

// globalConfig はパッケージ内でのみ変更可能な設定情報の格納先です。
var globalConfig Config

// GetConfig は現在の設定情報のコピーを返します。
func GetConfig() Config {
	return globalConfig
}

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

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 1. 共通ロジック
			if globalConfig.Verbose {
				// 必要に応じて共通のロギング初期化などを実行
			}

			// 2. カスタムロジックの実行
			if app.PreRunE != nil {
				return app.PreRunE(cmd, args)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				fmt.Fprintf(os.Stderr, "Error displaying help: %v\n", err)
			}
		},
	}

	// 共通フラグの登録（バインド先を非公開変数に変更）
	rootCmd.PersistentFlags().BoolVarP(&globalConfig.Verbose, "verbose", "V", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&globalConfig.ConfigFile, "config", "C", "", "Config file path")

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
		os.Exit(1)
	}
}
