package main

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
)

type GistDaemon struct {
	rootDir string
	gists   []string
	watcher *fsnotify.Watcher
}

func NewGistDaemon(rootDir string) *GistDaemon {
	return &GistDaemon{rootDir: rootDir}
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
	if err := gd.addWatchers(); err != nil {
		return err
	}
	return nil
}

func main() {
	ch := make(chan bool)

	daemon := NewGistDaemon("/home/jammin/gists/kdjfd")
	if err := daemon.Start(); err != nil {
		log.Fatal(err)
	}
	defer daemon.Close()

	<-ch
}
