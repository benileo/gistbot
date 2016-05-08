package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"regexp"
)

type Watcher struct {
	Conf    *Config
	Watcher *fsnotify.Watcher
}

func NewWatcher(conf *Config) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creating watcher: %v", err)
	}

	return &Watcher{Conf: conf, Watcher: watcher}, nil
}

func (w *Watcher) AddWatches(paths []string) []error {
	errors := make([]error, 0)
	for _, dir := range paths {
		if err := w.Add(dir); err != nil {
			errors = append(errors, err)
		}
	}

	return nil
}

func (w *Watcher) Add(path string) error {
	if err := w.Watcher.Add(path); err != nil {
		return fmt.Errorf("error adding watcher to (%s): %v", path, err)
	}

	log.Printf("watching: %s\n", path)
	return nil
}

func (w *Watcher) Watch(events chan string, errors chan error) {
	for {
		select {

		/* events */
		case event := <-w.Watcher.Events:
			switch event.Op {

			case fsnotify.Write:
				if !w.isReservedGitPath(event.Name) {
					log.Printf("save: %s", event.Name)
					events <- event.Name
				}
			}

		/* errors */
		case err := <-w.Watcher.Errors:
			errors <- err
		}
	}
}

func (w *Watcher) isReservedGitPath(path string) bool {
	exp, _ := regexp.Compile(`/\.git.*|.*.swp`)

	return exp.MatchString(path)
}
