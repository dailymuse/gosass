package compiler

import (
	"errors"
	"gopkg.in/fsnotify.v1"
	"log"
	"os"
	"path/filepath"
	"time"
)

type SassWatcher struct {
	watcher   *fsnotify.Watcher
	ctx       *SassContext
	filecache *FileCache
	deps      *SassDependencyResolver
	staged    map[string]string
}

func NewSassWatcher(ctx *SassContext) (*SassWatcher, error) {
	info, err := os.Stat(ctx.inputPath)

	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, errors.New("Input must be a directory")
	}

	filecache := NewFileCache()

	// Create the watcher
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	watcher.Add(ctx.inputPath)

	// Add subdirectories to be watched
	err = filepath.Walk(ctx.inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() {
			watcher.Add(path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &SassWatcher{
		watcher:   watcher,
		ctx:       ctx,
		filecache: filecache,
		deps:      NewSassDependencyResolver(filecache),
		staged:    make(map[string]string, 100),
	}, nil
}

func (self *SassWatcher) stage(path string) error {

	//If this is a directory, we'll need to add it to the watch list.
	info, directoryErr := os.Stat(path)

	//Even if this isn't a directory, we should still be able to stat it.
	if directoryErr != nil {
		return directoryErr
	}

	if info.IsDir() {
		fileLog(false, path, "Adding newly created directory to watch list.")
		self.watcher.Add(path)

		// Add subdirectories to be watched
		directoryErr = filepath.Walk(path, func(path string, info os.FileInfo, directoryErr error) error {
			if directoryErr != nil {
				return directoryErr
			} else if info.IsDir() {
				self.watcher.Add(path)
			}

			return nil
		})

		if directoryErr != nil {
			return directoryErr
		}
	}

	// If it's not a directory, and not a Sass file, we don't want it.
	if !isSassFile(path) {
		return nil
	}

	// Invalidate caches
	err := self.filecache.Invalidate(path)

	if err != nil {
		return err
	}

	err = self.deps.Invalidate(path)

	if err != nil {
		return err
	}

	// Refresh dependencies
	_, err = self.deps.Resolve(path)

	if err != nil {
		return err
	}

	// Stage the file if it isn't private
	if !isPrivateFile(path) {
		self.staged[path] = self.ctx.resolveOutputPath(path)
	}

	// Stage the non-private dependents
	dependents, err := self.deps.ReverseResolve(path)

	if err != nil {
		return err
	}

	for _, dep := range dependents {
		if !isPrivateFile(dep) {
			self.staged[dep] = self.ctx.resolveOutputPath(dep)
		}
	}

	return nil
}

// Listens for changes in the watched directory/file
func (self *SassWatcher) listener() {
	for {
		select {
		case event := <-self.watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
				err := self.stage(event.Name)

				if err != nil {
					fileLog(false, event.Name, "Could not stage for compilation: %s", err.Error())
				}
			}
		case err := <-self.watcher.Errors:
			if err != nil {
				log.Fatalf("Watcher error: %s", err.Error())
			}
		}
	}
}

// Compiles files that are staged
func (self *SassWatcher) compile() {
	if len(self.staged) > 0 {
		compileMany(self.ctx, self.staged)
		self.staged = make(map[string]string, 100)
	}
}

// CLI endpoint for watching
func Watch(ctx *SassContext) {
	watcher, err := NewSassWatcher(ctx)

	if err != nil {
		log.Fatalf("Could not start watching: %s", err.Error())
	}

	// Make an initial compile of the files
	compilable := findCompilable(ctx)

	// Warm up the dependency cache
	for path := range compilable {
		watcher.deps.Resolve(path)
	}

	go watcher.listener()

	if compileMany(ctx, compilable) {
		log.Fatalf("Bailing due to compilation errors on the first pass")
	}

	// Periodically recompile any staged items. We do it this way to avoid
	// both issues with redundant watcher events on mac, and to prevent the
	// same file from getting compiled many times.
	for {
		watcher.compile()
		time.Sleep(100 * time.Millisecond)
	}
}
