package injector

import (
	"log"
	"os"
	"path/filepath"
)

type walker struct {
	mask  []string
	dirs  []string
	rules []Rule
}

func (w *walker) replace(path string) error {
	return replace(path, w.rules)
}

func (w *walker) walkDirs() error {
	for _, dir := range w.dirs {
		if err := filepath.Walk(dir, w.walkItem); err != nil {
			return err
		}
	}
	return nil
}

func (w *walker) walkItem(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}
	if !info.Mode().IsRegular() {
		log.Printf("%s is not a regular file", path)
		return nil
	}
	name := filepath.Base(path)
	for _, m := range w.mask {
		matched, err := filepath.Match(m, name)
		if err != nil {
			log.Printf("%s", err.Error())
			return err
		}
		if !matched {
			log.Printf("%s not matched by %s", path, m)
			continue
		}
		log.Printf("replacing values in %s", path)
		if err := w.replace(path); err != nil {
			log.Printf("%s", err.Error())
			return err
		}
	}
	return nil
}

func Walk(mask, dirs []string, replacements []Rule) error {
	walker := walker{mask: mask, dirs: dirs, rules: replacements}
	return walker.walkDirs()
}
