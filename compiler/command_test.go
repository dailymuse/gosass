package compiler

import (
	"strings"
	"testing"
)

func TestSassCommand(t *testing.T) {
	t.Parallel()

	cmd := NewSassCommand()
	cmd.AddArgument("--help")
	proc := cmd.Create("a")

	stdout, err := proc.Output()

	if err != nil {
		t.Error(err)
	}

	// Make sure it prints help since we added the --help flag
	if !strings.HasPrefix(string(stdout), "Usage: sassc") {
		t.Errorf("Unexpected stdout: %s", stdout)
	}
}
