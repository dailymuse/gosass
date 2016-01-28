package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func relPath(path string) string {
	return filepath.Join("../integration/src", path)
}

func absPath(path string) string {
	wd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	abs, err := filepath.Abs(filepath.Join(wd, relPath(path)))

	if err != nil {
		panic(err)
	}

	return abs
}

func checkArray(name string, t *testing.T, actual []string, expected ...string) {
	expectedMap := make(map[string]bool, 0)

	for _, val := range expected {
		expectedMap[val] = true
	}

	if len(expected) != len(actual) {
		t.Errorf("[%s] Unexpected result: %s", name, actual)
	} else {
		for _, val := range actual {
			_, ok := expectedMap[val]

			if !ok {
				t.Errorf("[%s] Unexpected value: %s", name, val)
			}

			delete(expectedMap, val)
		}
	}
}

func TestResolveRefPath(t *testing.T) {
	t.Parallel()

	sassPath, err := resolveRefPath(".", relPath("includes/first"))

	if err != nil {
		t.Error(err)
	} else if sassPath != "../integration/src/includes/_first.scss" {
		t.Error(fmt.Sprintf("Unexpected path: %s", sassPath))
	}

	cssPath, err := resolveRefPath(".", relPath("includes/fourth"))

	if err != nil {
		t.Error(err)
	} else if cssPath != "../integration/src/includes/fourth.css" {
		t.Error(fmt.Sprintf("Unexpected path: %s", cssPath))
	}
}

func TestShallowResolve(t *testing.T) {
	t.Parallel()

	resolver := NewSassDependencyResolver(NewFileCache())

	// Check simple resolution
	deps, err := resolver.shallowResolve(relPath("01.simple.scss"))

	if err != nil {
		t.Error(err)
	} else if len(deps) != 0 {
		t.Error(deps)
	}

	// Check resolution of multiple imports
	deps, err = resolver.shallowResolve(relPath("03.multiple-imports.scss"))

	if err != nil {
		t.Error(err)
	} else {
		checkArray("shallow-resolve:1", t, deps, relPath("includes/_first.scss"), relPath("includes/_second.scss"))
	}

	// Check that an error happens when trying to resolve a missing import
	deps, err = resolver.shallowResolve(relPath("04.missing.scss"))

	if err == nil {
		t.Error("Expected an error when resolving dependencies for 04.missing.scss")
	}

	// Check resolution of a CSS dependency
	deps, err = resolver.shallowResolve(relPath("05.rawcss.scss"))

	if err != nil {
		t.Error(err)
	} else {
		checkArray("shallow-resolve:2", t, deps, relPath("includes/fourth.css"))
	}
}

func TestResolve(t *testing.T) {
	t.Parallel()

	resolver := NewSassDependencyResolver(NewFileCache())

	// Resolve all the files
	deps, err := resolver.Resolve(relPath("01.simple.scss"))

	if err != nil {
		t.Error(err)
	} else if len(deps) != 0 {
		t.Error()
	}

	deps, err = resolver.Resolve(relPath("02.simple-import.scss"))

	if err != nil {
		t.Error(err)
	} else {
		checkArray("resolve:1", t, deps, absPath("includes/_first.scss"))
	}

	deps, err = resolver.Resolve(relPath("03.multiple-imports.scss"))

	if err != nil {
		t.Error(err)
	} else {
		checkArray("resolve:2", t, deps, absPath("includes/_first.scss"), absPath("includes/_second.scss"), absPath("includes/_third.scss"))
	}

	deps, err = resolver.Resolve(relPath("04.missing.scss"))

	if err == nil {
		t.Error("Expected an error when resolving a missing import")
	}

	deps, err = resolver.Resolve(relPath("05.rawcss.scss"))

	if err != nil {
		t.Error(err)
	} else {
		checkArray("resolve:3", t, deps, absPath("includes/fourth.css"))
	}

	deps, err = resolver.Resolve(relPath("includes/_second.scss"))

	if err != nil {
		t.Error(err)
	} else {
		checkArray("resolve:4", t, deps, absPath("includes/_third.scss"))
	}

	// Reverse resolve one
	deps, err = resolver.ReverseResolve(relPath("includes/_third.scss"))

	if err != nil {
		t.Error(err)
	} else {
		checkArray("resolve:5", t, deps, absPath("03.multiple-imports.scss"), absPath("includes/_second.scss"))
	}

	// Invalidate an item
	err = resolver.Invalidate(relPath("includes/_second.scss"))

	if err != nil {
		t.Error(err)
	}

	// Make sure reverse resolve was affected by the invalidation
	deps, err = resolver.ReverseResolve(relPath("includes/_third.scss"))

	if err != nil {
		t.Error(err)
	} else {
		checkArray("resolve:6", t, deps, absPath("03.multiple-imports.scss"))
	}

	// Make sure re-resolving a file includes all dependencies (including invalidated ones)
	deps, err = resolver.Resolve(relPath("03.multiple-imports.scss"))

	if err != nil {
		t.Error(err)
	} else {
		checkArray("resolve:7", t, deps, absPath("includes/_first.scss"), absPath("includes/_second.scss"), absPath("includes/_third.scss"))
	}
}
