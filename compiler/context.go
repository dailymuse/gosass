package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Stores contextual information for CLI invocations
type SassContext struct {
	cmd        *SassCommand
	inputPath  string
	outputPath string
}

func NewSassContext(cmd *SassCommand, inputPath string, outputPath string) *SassContext {
	return &SassContext{
		cmd:        cmd,
		inputPath:  inputPath,
		outputPath: outputPath,
	}
}

// Gets the equivalent output path for the given path. The given path must be
// within the `inputPath`, but it may be in absolute or relative form.
func (self *SassContext) resolveOutputPath(p string) string {
	absInput, err := filepath.Abs(self.inputPath)

	if err != nil {
		panic(err)
	}

	if filepath.IsAbs(p) {
		p, err = filepath.Rel(absInput, p)

		if err != nil {
			panic(err)
		}
	} else {
		wd, err := os.Getwd()

		if err != nil {
			panic(err)
		}

		p, err = filepath.Rel(absInput, filepath.Join(wd, p))

		if err != nil {
			panic(err)
		}
	}

	// Replace .scss with .css
	np := filepath.Join(self.outputPath, p)
	ext := filepath.Ext(np)

	if ext == ".scss" {
		np = np[0:len(np)-len(ext)] + ".css"
	}

	// Sanity check the results - make sure the path is a sub-path of the
	// output path
	rel, err := filepath.Rel(self.outputPath, np)

	if err != nil {
		panic(err)
	}

	if strings.Contains(rel, "..") {
		panic(fmt.Sprintf("Resolved output path not part of the output directory. Resolved output path = %s; output directory = %s", rel, self.outputPath))
	}

	return np
}
