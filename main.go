package main

import (
	"log"
	"github.com/fsnotify/fsnotify"
)

func main() {
	ch := make(chan bool)

	// load the config
	conf := Config{
		RootDir: "/home/jammin/gists",
		PublicKey: "/home/jammin/.ssh/id_rsa.pub",
		PrivateKey: "/home/jammin/.ssh/id_rsa",
	}

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

	// set up git control
	repos, err := createRepos(gists)
	if err != nil {
		log.Fatal(err)
	}
	gitcontrol := NewGitControl(conf, repos)

	// create and start the daemon
	daemon := NewDaemon(conf, gists, watcher, gitcontrol)
	err = daemon.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer daemon.Stop()

	<-ch
}
