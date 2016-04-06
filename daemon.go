package main

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/libgit2/git2go.v23"
	"log"
	"os"
	"path/filepath"
)

type GistDaemon struct {
	gc      *GitControl
	rootDir string
	gists   []string
	watcher *fsnotify.Watcher
}

func NewGistDaemon(rootDir string, gc *GitControl) *GistDaemon {
	return &GistDaemon{rootDir: rootDir, gc: gc}
}

func (gd *GistDaemon) findGists() ([]string, error) {
	gitDirectories := make([]string, 0)
	err := filepath.Walk(gd.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == ".git" && info.IsDir() {
			absPath, err := filepath.Abs(filepath.Dir(path))
			if err != nil {
				return err
			}
			gitDirectories = append(gitDirectories, absPath)
			return filepath.SkipDir
		}
		return nil
	})
	return gitDirectories, err
}

func (gd *GistDaemon) setGists() error {
	gists, err := gd.findGists()
	if err != nil {
		return err
	}
	if len(gists) == 0 {
		return errors.New("No github gists were found")
	}
	gd.gists = gists
	return nil
}

func (gd *GistDaemon) createWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	gd.watcher = watcher
	return nil
}

func (gd *GistDaemon) addWatchers() error {
	for _, dir := range gd.gists {
		if err := gd.watcher.Add(dir); err != nil {
			return err
		}
		log.Println("Setup watch on gist: ", dir)
	}
	return nil
}

func (gd *GistDaemon) waitForEvents() {
	for {
		select {
		case event := <-gd.watcher.Events:
			gd.handleEvent(event)
		case err := <-gd.watcher.Errors:
			log.Println("error: ", err)
			// handle errors here
		}
	}
}

func (gd *GistDaemon) handleEvent(event fsnotify.Event) {
	if event.Op == fsnotify.Write {
		log.Println("Write event to", event.Name)
		if err := gd.handleWrite(event); err != nil {
			log.Println(err)
		}
	}
}

func (gd *GistDaemon) handleWrite(event fsnotify.Event) error {
	return nil
}

func (gd *GistDaemon) pullGists() error {
	for _, repo := range gd.gists {
		repository, err := git.OpenRepository(repo)
		if err != nil {
			return fmt.Errorf("unable to open repo with path %s: %v", repo, err)
		}
		remote, err := repository.Remotes.Lookup("origin")
		if err != nil {
			return fmt.Errorf("unable to find remote origin: %v", err)
		}
		fmt.Println(remote.Name())
		if err := remote.Fetch(make([]string, 0), nil, ""); err != nil {
			return fmt.Errorf("error fetching repo: %v", err)
		}
	}
	return nil
}

func (gd *GistDaemon) Close() {
	if gd.watcher != nil {
		gd.watcher.Close()
	}
}

func (gd *GistDaemon) Start() error {
	if err := gd.setGists(); err != nil {
		return err
	}
	if err := gd.createWatcher(); err != nil {
		return err
	}
	if err := gd.pullGists(); err != nil {
		return err
	}
	if err := gd.addWatchers(); err != nil {
		return err
	}
	go gd.waitForEvents()
	return nil
}
