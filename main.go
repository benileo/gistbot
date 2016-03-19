package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
	"errors"
)

type GistDaemon struct {
	rootDir string
	gists   []string
	watcher *fsnotify.Watcher
}

func NewGistDaemon(rootDir string, watcher *fsnotify.Watcher) *GistDaemon {
	return &GistDaemon{rootDir: rootDir, watcher: watcher}
}

func (gd *GistDaemon) findGists() ([]string, error) {
	gitDirectories := make([]string, 0)
	err := filepath.Walk(gd.rootDir, func(path string, info os.FileInfo, err error) error {
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
		return errors.New("No gists were found")
	}
	gd.gists = gists
	return nil
}

//func (gd *GistDaemon) SetupWatchers(){
//
//}

func (gd *GistDaemon) Start() error {
	if err := gd.setGists(); err != nil {
		return err
	}
	return nil
}

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	daemon := NewGistDaemon("/home/jammin/gists", watcher)
	daemon.Start()
	fmt.Println(daemon.gists)
}
