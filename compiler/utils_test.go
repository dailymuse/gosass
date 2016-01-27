package compiler

import (
	"testing"
)

func TestIsSassFile(t *testing.T) {
	t.Parallel()

	if !isSassFile("foo/bar.scss") {
		t.Error()
	}

	if !isSassFile("foo/bar.css") {
		t.Error()
	}
}

func TestIsPrivateFile(t *testing.T) {
	t.Parallel()

	if !isPrivateFile("foo/_bar.css") {
		t.Error()
	}

	if isPrivateFile("foo/bar.scss") {
		t.Error()
	}
}
