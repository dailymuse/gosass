package compiler

import (
	"os/exec"
)

type SassCommand struct {
	Args []string
}

func NewSassCommand() *SassCommand {
	return &SassCommand{
		Args: make([]string, 0),
	}
}

func (self *SassCommand) AddArgument(arg string) {
	self.Args = append(self.Args, arg)
}

func (self *SassCommand) Create(inputFilePath string, outputFilePath string) *exec.Cmd {
	args := make([]string, len(self.Args)+2)
	copy(args, self.Args)
	args[len(args)-2] = inputFilePath
	args[len(args)-1] = outputFilePath
	return exec.Command("sassc", args...)
}
