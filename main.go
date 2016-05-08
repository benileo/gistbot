package main

import (
	"log"
	"flag"
)

var configFile = flag.String("config-file", "my_config.json", "The file path of your config file")

func main() {
	ch := make(chan bool)
	flag.Parse()

	// load the config
	conf, err := NewConfig(*configFile)
	if err != nil {
		log.Fatal(err)
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
