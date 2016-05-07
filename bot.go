package main

import (
	"fmt"
	"log"
)

type Bot struct {
	finder  *Finder
	watcher *Watcher
	conf    Config
	repos   []string
	events  chan string
	errors  chan error
}

func NewBot(conf Config) (*Bot, error) {
	bot := Bot{
		conf:   conf,
		events: make(chan string),
		errors: make(chan error),
	}
	bot.finder = NewFinder(conf)

	watcher, err := NewWatcher(conf)
	if err != nil {
		return nil, err
	}
	bot.watcher = watcher

	return &bot, nil
}

func (b *Bot) Start() error {
	if err := b.paths(); err != nil {
		return err
	}

	b.pullAll()

	for _, err := range b.watcher.AddWatches(b.repos) {
		log.Println(err)
	}

	return nil
}

func (b *Bot) Stop() error {
	if err := b.watcher.Watcher.Close(); err != nil {
		return fmt.Errorf("error while closing watcher: %v", err)
	}

	return nil
}

// Crappy name, this sets the git paths of the Bot using the finder
func (b *Bot) paths() error {
	repos, err := b.finder.Find()
	if err != nil {
		return fmt.Errorf("error finding repos %v", err)
	}

	fmt.Println(repos)
	b.repos = repos

	return nil
}

func (b *Bot) pullAll() {
	ch := make(chan error)
	badPaths := make([]string, 0)

	for _, path := range b.repos {
		repo, err := NewRepository(b.conf, path)
		if err != nil {
			badPaths = append(badPaths, path)
			continue
		}

		// Pull the repository in a go routine
		go repo.Pull(ch)
	}

	for i := 0; i < len(b.repos)-len(badPaths); i++ {
		err := <-ch
		if err != nil {
			// Todo: need to add to the bad paths here. how do I know what the path is..?
			fmt.Println(err)
		}
	}

	// Todo: remove all bad paths (in a bad state) from *Bot.repos
}
