package main

import (
	"strings"
	"testing"
)

func TestPlanCommandAcceptsPositionalComponentArg(t *testing.T) {
	if planCmd.Args == nil {
		t.Fatal("expected plan command to have an Args validator")
	}
	if err := planCmd.Args(planCmd, []string{"my-component"}); err != nil {
		t.Fatalf("expected single positional arg to be accepted: %v", err)
	}
}

func TestPlanCommandRejectsTooManyArgs(t *testing.T) {
	if err := planCmd.Args(planCmd, []string{"comp-a", "comp-b"}); err == nil {
		t.Fatal("expected two positional args to be rejected")
	}
}

func TestPlanCommandUseSyntaxIncludesComponent(t *testing.T) {
	if !strings.Contains(planCmd.Use, "[component]") {
		t.Fatalf("expected planCmd.Use to contain [component], got %q", planCmd.Use)
	}
}

func TestPlanCommandRegistersComponentFlag(t *testing.T) {
	if planCmd.Flags().Lookup("component") == nil {
		t.Fatal("expected plan command to register --component flag")
	}
}
