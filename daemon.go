package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
)

type Config struct {
	RootDir    string
	PublicKey  string
	PrivateKey string
	Username   string
	Name       string
	Email      string
}

type Daemon struct {
	Conf       Config
	Gists      []string
	Watcher    *fsnotify.Watcher
	GitControl *GitControl
}

func NewDaemon(conf Config, gists []string, watcher *fsnotify.Watcher, gitcontrol *GitControl) *Daemon {
	return &Daemon{Conf: conf, Gists: gists, Watcher: watcher, GitControl: gitcontrol}
}

func (d *Daemon) Stop() {
	d.Watcher.Close()
}

func (d *Daemon) Start() error {
	d.GitControl.PullAll()
	//err := d.GitControl.PullAll() todo: this should probabaly return something..
	//if err != nil {
	//	return err
	//}
	d.addWatchers()
	go d.waitForEvents()

	return nil
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
				// todo: changes to the .git directory should not be updated.
				if err := d.GitControl.Update(event.Name); err != nil {
					fmt.Errorf("error updating repo: %v", err)
				}
			}

		/* errors */
		case err := <-d.Watcher.Errors:
			fmt.Errorf("watcher event: %v", err)
		}
	}
}
