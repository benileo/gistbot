package main

import (
	"log"
)

func main() {
	ch := make(chan bool)

	daemon := NewGistDaemon("/home/jammin/gists",
		NewGitControl("benileo", "home/jammin/.ssh/id_rsa.pub", "home/jammin/.ssh/id_rsa"))
	if err := daemon.Start(); err != nil {
		log.Fatal(err)

	}
	defer daemon.Close()

	<-ch
}
