package main

import (
	"fmt"
	"log"
	"path/filepath"
)

type Bot struct {
	finder  *Finder
	watcher *Watcher
	conf    *Config
	repos   []string
	events  chan string
	errors  chan error
}

func NewBot(conf *Config) (*Bot, error) {
	bot := Bot{
		conf:   conf,
		events: make(chan string, 3),
		errors: make(chan error, 2),
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

	// Watch and Listen in separate go routines
	go b.listenForChanges()
	go b.watcher.Watch(b.events, b.errors)

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

func (b *Bot) listenForChanges() {
	log.Println("listening for changes...")
	for {
		select {

		case changedFile := <-b.events:
			if err := b.updateRepository(changedFile); err != nil {
				log.Printf("error updating the repository %s: %v", changedFile, err)
			}
			log.Printf("repository updated")

		case err := <-b.errors:
			log.Printf("error from the watcher: %v", err)
		}
	}
}

func (b *Bot) updateRepository(changedFile string) error {
	dirPath, err := filepath.Abs(filepath.Dir(changedFile))
	if err != nil {
		return err
	}

	repo, err := NewRepository(b.conf, dirPath)
	if err != nil {
		return err
	}

	tree, err := repo.Add()
	if err != nil {
		return fmt.Errorf("error git add: %v", err)
	}

	if err = repo.Commit(tree); err != nil {
		return err
	}

	if err = repo.Push(); err != nil {
		return fmt.Errorf("error git push: %v", err)
	}

	return nil
}
