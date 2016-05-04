package main

import (
	"log"
	"github.com/fsnotify/fsnotify"
)

func main() {
	ch := make(chan bool)

	// load the config
	conf := Config{RootDir: "/home/jammin/gists"}

	// get the file paths of the gists
	gists, err := findGists(conf.RootDir)
	if err != nil {
		log.Fatal(err)
	}

	// get the watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// create and start the daemon
	daemon := NewDaemon(conf, gists, watcher)
	daemon.Start()
	defer daemon.Stop()

	<-ch
}
