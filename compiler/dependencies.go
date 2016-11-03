package compiler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var importPattern = regexp.MustCompile("(?m)^\\s*([^/]?)\\s*\\@import (\\'|\\\")([^'\"]+)(\\'|\\\")")

type SassDependencyResolver struct {
	filecache   *FileCache
	shallowDeps map[string][]string
	deepDeps    map[string][]string
}

func NewSassDependencyResolver(filecache *FileCache) *SassDependencyResolver {
	return &SassDependencyResolver{
		filecache:   filecache,
		shallowDeps: make(map[string][]string, 100),
		deepDeps:    make(map[string][]string, 100),
	}
}

// Resolves the path of a given import
func resolveRefPath(basePath string, ref string) (string, error) {
	var filename string
	if filepath.Ext(ref) == "" {
		filename = fmt.Sprintf("_%s.scss", filepath.Base(ref))
	} else {
		filename = filepath.Base(ref)

	}
	sassImportPath := filepath.Join(basePath, filepath.Dir(ref), filename)

	stat, err := os.Stat(sassImportPath)

	if err == nil && !stat.IsDir() {
		return sassImportPath, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", errors.New(fmt.Sprintf("Error when trying to stat ref `%s` (which resolved to '%s'): %s", ref, sassImportPath, err.Error()))
	}

	cssImportPath := filepath.Join(basePath, filepath.Dir(ref), fmt.Sprintf("%s.css", filepath.Base(ref)))
	stat, err = os.Stat(cssImportPath)

	if err == nil && !stat.IsDir() {
		return cssImportPath, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", errors.New(fmt.Sprintf("Error when trying to stat ref `%s` (which resolved to '%s'): %s", ref, cssImportPath, err.Error()))
	} else {
		return "", errors.New(fmt.Sprintf("Could not find ref `%s` (tried '%s' and '%s')", ref, sassImportPath, cssImportPath))
	}
}

// Gets the files imported directly by the given file
func (self *SassDependencyResolver) shallowResolve(path string) ([]string, error) {
	deps, ok := self.shallowDeps[path]

	if ok {
		return deps, nil
	}

	// Get the file contents
	contents, err := self.filecache.Get(path)

	if err != nil {
		return nil, err
	}

	// Build the matches
	matches := importPattern.FindAllSubmatch(contents, -1)
	deps = make([]string, len(matches))

	for i, match := range matches {
		ref := string(match[3])
		refPath, err := resolveRefPath(filepath.Dir(path), ref)

		if err != nil {
			return nil, err
		}

		deps[i] = refPath
	}

	self.shallowDeps[path] = deps
	return deps, nil
}

// Gets all files imported by the given file, including indirect imports
func (self *SassDependencyResolver) Resolve(path string) ([]string, error) {
	abs, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	deps, ok := self.deepDeps[abs]

	if ok {
		return deps, nil
	}

	scanned := make(map[string]bool, 100)
	unscanned := make(map[string]bool, 100)
	unscanned[abs] = true

	for len(unscanned) > 0 {
		for subpath := range unscanned {
			// Move the file to scanned
			delete(unscanned, subpath)
			scanned[subpath] = true

			// Get the dependencies or read them if needed
			deps, err := self.shallowResolve(subpath)

			if err != nil {
				return nil, err
			}

			// Add the dependency to unscanned if it hasn't been scanned
			// already
			for _, dep := range deps {
				_, ok = scanned[dep]

				if !ok {
					unscanned[dep] = true
				}
			}
		}
	}

	deps = make([]string, 0, len(scanned))

	for dep := range scanned {
		if dep != abs {
			deps = append(deps, dep)
		}
	}

	self.deepDeps[abs] = deps
	return deps, nil
}

// Gets what files are dependent on the given file, including indirectly
func (self *SassDependencyResolver) ReverseResolve(path string) ([]string, error) {
	abs, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	reverseDeps := make([]string, 0)

	for otherPath, deps := range self.deepDeps {
		for _, dep := range deps {
			if dep == abs {
				reverseDeps = append(reverseDeps, otherPath)
				break
			}
		}
	}

	return reverseDeps, nil
}

// Invalidates the cached entry for the given file
func (self *SassDependencyResolver) Invalidate(path string) error {
	abs, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	delete(self.shallowDeps, abs)
	delete(self.deepDeps, abs)
	return nil
}
