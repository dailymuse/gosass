package main

import (
	"flag"
	"github.com/dailymuse/gosass/compiler"
	"log"
	"os"
	"runtime/pprof"
)

func main() {
	// Disable date/time outputs of the log
	log.SetFlags(0)

	var style = flag.String("style", "", "Output style. Can be: nested, compressed.")
	var lineNumbers = flag.Bool("line-numbers", false, "Emit comments showing original line numbers.")
	var loadPath = flag.String("load-path", "", "Set Sass import path.")
	var sourcemap = flag.Bool("sourcemap", false, "Emit source map.")
	var omitMapComment = flag.Bool("omit-map-comment", false, "Omits the source map url comment.")
	var watch = flag.Bool("watch", false, "Watch for changes and automatically recompile when needed.")
	var cpuProfile = flag.String("cpuprofile", "", "Write CPU profile to file.")
	var inputPath = flag.String("input", "", "Input file or directory.")
	var outputPath = flag.String("output", "", "Output file or directory.")

	flag.Parse()

	// Enable CPU profiling if requested
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)

		if err != nil {
			log.Fatal(err)
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Validate the arguments
	if *style != "" && *style != "nested" && *style != "compressed" {
		log.Fatal("Style must be 'nested' or 'compressed'")
	}

	// Create the sass command
	cmd := compiler.NewSassCommand()

	if *style != "" {
		cmd.AddArgument("--style")
		cmd.AddArgument(*style)
	}

	if *lineNumbers {
		cmd.AddArgument("--line-numbers")
	}

	if *loadPath != "" {
		cmd.AddArgument("--load-path")
		cmd.AddArgument(*loadPath)
	}

	if *sourcemap {
		cmd.AddArgument("--sourcemap")
	}

	if *omitMapComment {
		cmd.AddArgument("--omit-map-comment")
	}

	inputStat, inputStatErr := os.Stat(*inputPath)

	if inputStatErr != nil {
		log.Fatalf("Could not stat input path: %s", inputStatErr.Error())
	}

	outputStat, outputStatErr := os.Stat(*outputPath)

	if inputStat.IsDir() {
		if outputStatErr != nil {
			log.Fatalf("Could not stat output path: %s", outputStatErr.Error())
		} else if !outputStat.IsDir() {
			log.Fatalf("Input path is a directory, but output path is a file")
		}
	} else {
		if outputStatErr != nil && !os.IsNotExist(outputStatErr) {
			log.Fatalf("Could not stat output path: %s", outputStatErr.Error())
		} else if outputStatErr == nil && outputStat.IsDir() {
			log.Fatalf("Input path is a file, but output path is a directory")
		}
	}

	ctx := compiler.NewSassContext(cmd, *inputPath, *outputPath)

	if *watch {
		compiler.Watch(ctx)
	} else {
		compiler.Compile(ctx)
	}
}
