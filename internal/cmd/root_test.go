package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/activtrak-mfinlayson/atadmin/internal/cmd"
)

func TestRootVersion(t *testing.T) {
	root := cmd.NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"--version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "atadmin version "+cmd.Version) {
		t.Errorf("version output = %q, want it to contain %q", got, "atadmin version "+cmd.Version)
	}
}

func TestRootHelp(t *testing.T) {
	root := cmd.NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "atadmin") {
		t.Errorf("help output = %q, does not contain %q", got, "atadmin")
	}
	if !strings.Contains(got, "auth") {
		t.Errorf("help output = %q, does not list 'auth' subcommand", got)
	}
}

func TestRootUnknownCommand(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"unknowncmd"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for unknown command, got nil")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("error = %q, want to contain 'unknown command'", err.Error())
	}
}

func TestRootNoArgs(t *testing.T) {
	root := cmd.NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{})

	// No args should show help (not error).
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() with no args returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "atadmin") {
		t.Error("no-args output does not contain 'atadmin'")
	}
}
