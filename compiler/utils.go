package compiler

import (
	"path/filepath"
)

// Returns whether the given file is sass
func isSassFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".scss" || ext == ".css"
}

// Returns whether the given file is private (i.e. whether it's included by
// root sass files)
func isPrivateFile(path string) bool {
	base := filepath.Base(path)
	ext := filepath.Ext(path)

	if ext == ".scss" {
		return len(base) > 0 && base[0] == '_'
	} else {
		return ext == ".css"
	}
}
