// Package clibase は、cobra を用いた CLI の共通のルートコマンドと
// シグナル処理・共通フラグの組み立てを提供します。
package clibase

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// Config は共通フラグの値を保持する内部構造体です。
type Config struct {
	Verbose    bool
	ConfigFile string
}

// globalConfig はパッケージ内でのみ変更可能な設定情報の格納先です。
// Execute/ExecuteContext はプロセスにつき一度だけ呼び出される想定のため、
// 同時実行に対する排他制御は行っていません。
var globalConfig Config

// GetConfig は現在の設定情報のコピーを返します。
// これにより、利用側は読み取り専用として安全に設定を参照できます。
func GetConfig() Config {
	return globalConfig
}

// App は CLI アプリケーションの構成を定義します。
type App struct {
	Name string

	// Version を指定すると cobra 標準の --version フラグが有効になります。
	Version string

	// SilenceUsage / SilenceErrors は cobra のエラー時自動出力を抑制します。
	// 実行時エラーのたびに usage 全文が出るのを避けたい場合は true にします。
	SilenceUsage  bool
	SilenceErrors bool

	AddFlags func(cmd *cobra.Command)                      // 独自フラグ登録用
	PreRunE  func(cmd *cobra.Command, args []string) error // 実行前バリデーション/初期化用
	PostRun  func(cmd *cobra.Command, args []string)       // 実行後のリソース解放用
	Commands []*cobra.Command                              // サブコマンド群
}

// buildRootCmd は App の設定から cobra のルートコマンドを構築します。
func buildRootCmd(app App) *cobra.Command {
	// サブコマンドが独自の PersistentPreRunE/PersistentPostRun を定義していても
	// root の hook が無視されないようにします。
	// cobra のデフォルトではコマンドツリー中「最も近い」hook しか実行されないため、
	// このフラグなしでは app.PreRunE/PostRun がサブコマンド実行時に呼ばれません。
	cobra.EnableTraverseRunHooks = true

	rootCmd := &cobra.Command{
		Use:     app.Name,
		Short:   fmt.Sprintf("%s CLI tool", app.Name),
		Long:    fmt.Sprintf("%s is a CLI application built with shouni/clibase.", app.Name),
		Version: app.Version,

		SilenceUsage:  app.SilenceUsage,
		SilenceErrors: app.SilenceErrors,

		// アプリケーション固有の実行前処理を呼び出す
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if app.PreRunE != nil {
				return app.PreRunE(cmd, args)
			}
			return nil
		},

		// コマンド実行後に必ず呼び出されるクリーンアップ処理
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if app.PostRun != nil {
				app.PostRun(cmd, args)
			}
		},

		// 引数なしで実行された場合にヘルプを表示
		// RunE にすることで、エラーハンドリングを上位の Execute() に委ねます
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	// 共通フラグの登録
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

	return rootCmd
}

// ExecuteContext は App からルートコマンドを構築し、ctx を伝搬させて実行します。
// os.Exit を呼ばずにエラーを返すため、終了コードの制御や独自のキャンセル文脈を
// 呼び出し側で扱いたい場合に使用します。
func ExecuteContext(ctx context.Context, app App) error {
	return buildRootCmd(app).ExecuteContext(ctx)
}

// Execute は、アプリケーションの構築と実行をワンストップで行います。
// SIGINT/SIGTERM を受信すると ctx をキャンセルするため、PreRunE/PostRun や
// 各コマンドの Run 内で cmd.Context() を参照すれば中断処理に反応できます。
func Execute(app App) {
	// os.Exit は defer を実行せずにプロセスを終えるため、シグナル購読の解除は
	// 別関数へ閉じ込めて、終了コードの決定だけをここに残します。
	if err := executeWithSignals(app); err != nil {
		// Cobraがエラーを出力するため、ここでは適切な終了コードで終了します
		os.Exit(1)
	}
}

// executeWithSignals は SIGINT/SIGTERM の購読を伴ってアプリケーションを実行し、
// 戻る際に必ず購読を解除します。
func executeWithSignals(app App) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return ExecuteContext(ctx, app)
}
