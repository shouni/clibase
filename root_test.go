package clibase

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/cobra"
)

func TestBuildRootCmd_Basics(t *testing.T) {
	app := App{
		Name:          "myapp",
		Version:       "1.2.3",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd := buildRootCmd(app)

	if cmd.Use != "myapp" {
		t.Errorf("Use = %q, want %q", cmd.Use, "myapp")
	}
	if cmd.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", cmd.Version, "1.2.3")
	}
	if !cmd.SilenceUsage {
		t.Error("SilenceUsage = false, want true")
	}
	if !cmd.SilenceErrors {
		t.Error("SilenceErrors = false, want true")
	}
	if f := cmd.PersistentFlags().Lookup("verbose"); f == nil {
		t.Error("expected --verbose flag to be registered")
	}
	if f := cmd.PersistentFlags().Lookup("config"); f == nil {
		t.Error("expected --config flag to be registered")
	}
}

func TestBuildRootCmd_AddFlagsAndCommands(t *testing.T) {
	called := false
	sub := &cobra.Command{Use: "sub", Run: func(cmd *cobra.Command, args []string) {}}
	app := App{
		Name: "myapp",
		AddFlags: func(cmd *cobra.Command) {
			called = true
			cmd.PersistentFlags().String("extra", "", "extra flag")
		},
		Commands: []*cobra.Command{sub},
	}
	cmd := buildRootCmd(app)

	if !called {
		t.Error("AddFlags was not invoked")
	}
	if cmd.PersistentFlags().Lookup("extra") == nil {
		t.Error("expected --extra flag to be registered")
	}

	found := false
	for _, c := range cmd.Commands() {
		if c == sub {
			found = true
		}
	}
	if !found {
		t.Error("expected sub command to be attached to root")
	}
}

// TestExecuteContext_RootAndChildHooksBothFire is a regression test for the
// PersistentPreRunE/PersistentPostRun override problem: cobra normally only
// runs the closest hook in the command tree, silently skipping the root's
// app.PreRunE/app.PostRun whenever a subcommand defines its own persistent
// hooks. buildRootCmd enables cobra.EnableTraverseRunHooks to fix this.
func TestExecuteContext_RootAndChildHooksBothFire(t *testing.T) {
	var order []string

	child := &cobra.Command{
		Use: "child",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			order = append(order, "child-pre")
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			order = append(order, "run")
		},
	}

	app := App{
		Name: "myapp",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			order = append(order, "root-pre")
			return nil
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			order = append(order, "root-post")
		},
		Commands: []*cobra.Command{child},
	}

	cmd := buildRootCmd(app)
	cmd.SetArgs([]string{"child"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("ExecuteContext returned error: %v", err)
	}

	want := []string{"root-pre", "child-pre", "run", "root-post"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}

func TestExecuteContext_ReturnsErrorWithoutExiting(t *testing.T) {
	app := App{
		Name:          "myapp",
		SilenceUsage:  true,
		SilenceErrors: true,
		Commands: []*cobra.Command{
			{
				Use: "fail",
				RunE: func(cmd *cobra.Command, args []string) error {
					return errors.New("boom")
				},
			},
		},
	}

	cmd := buildRootCmd(app)
	cmd.SetArgs([]string{"fail"})
	err := cmd.ExecuteContext(context.Background())
	if err == nil || err.Error() != "boom" {
		t.Fatalf("err = %v, want boom", err)
	}
}

func TestGetConfig_ReflectsParsedFlags(t *testing.T) {
	app := App{Name: "myapp"}
	cmd := buildRootCmd(app)
	cmd.SetArgs([]string{"--verbose", "--config", "myconf.yaml"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("ExecuteContext returned error: %v", err)
	}

	cfg := GetConfig()
	if !cfg.Verbose {
		t.Error("cfg.Verbose = false, want true")
	}
	if cfg.ConfigFile != "myconf.yaml" {
		t.Errorf("cfg.ConfigFile = %q, want %q", cfg.ConfigFile, "myconf.yaml")
	}
}
