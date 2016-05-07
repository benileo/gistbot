package main

import (
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

func main() {
	ch := make(chan bool)

	// load the config
	conf := Config{
		RootDir:    "/home/jammin/gists",
		PublicKey:  "/home/jammin/.ssh/id_rsa.pub",
		PrivateKey: "/home/jammin/.ssh/id_rsa",
		Name:       "Ben Irving",
		Email:      "jammin.irving@gmail.com",
	}

	bot, err := NewBot(conf)
	if err != nil {
		log.Fatal(err)
	}

	if err = bot.Start(); err != nil {
		log.Fatal(err)
	}
	defer bot.Stop()

	<-ch
}
