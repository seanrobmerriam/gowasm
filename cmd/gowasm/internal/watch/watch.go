package watch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Options struct {
	Dir    string
	Build  func() error
	Reload func()
}

func Watch(opts Options) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	filepath.Walk(opts.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			watcher.Add(path)
		}
		return nil
	})

	debounce := time.NewTimer(200 * time.Millisecond)

	go func() {
		for {
			select {
			case <-debounce.C:
				if err := opts.Build(); err != nil {
					fmt.Fprintf(os.Stderr, "build error: %v\n", err)
				} else {
					opts.Reload()
				}
			}
		}
	}()

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				if strings.HasSuffix(event.Name, ".go") {
					debounce.Reset(200 * time.Millisecond)
				}
			}
		case err := <-watcher.Errors:
			fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
		}
	}
}
