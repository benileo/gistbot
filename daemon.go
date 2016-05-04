package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
)

type Config struct {
	RootDir string
}

type Daemon struct {
	Conf    Config
	Gists   []string
	Watcher *fsnotify.Watcher
}

func NewDaemon(conf Config, gists []string, watcher *fsnotify.Watcher) (*Daemon) {
	return &Daemon{Conf: conf, Gists: gists, Watcher: watcher}
}

func (d *Daemon) Stop() {
	d.Watcher.Close()
}

func (d *Daemon) Start() {
	d.addWatchers()
	go d.waitForEvents()
}

func (d *Daemon) addWatchers() {
	for _, dir := range d.Gists {
		if err := d.Watcher.Add(dir); err != nil {
			fmt.Errorf("error adding watcher: %v", err)
		} else {
			log.Printf("watching: %s\n", dir)
		}
	}
}

func (d *Daemon) waitForEvents() {
	for {
		select {

		/* events */
		case event := <-d.Watcher.Events:
			switch event.Op {

			case fsnotify.Write:
				log.Printf("Write! %s", event.Name)
			}

		/* errors */
		case err := <-d.Watcher.Errors:
			fmt.Errorf("watcher event: %v", err)
		}
	}
}