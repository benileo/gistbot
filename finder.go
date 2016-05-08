package main

import (
	"os"
	"path/filepath"
)

type Finder struct {
	conf *Config
}

func NewFinder(conf *Config) *Finder {
	return &Finder{conf: conf}
}

func (f *Finder) Find() ([]string, error) {
	// Leave room to hook into multiple directories or other config options
	return f.find(f.conf.RootDir)
}

func (f *Finder) find(rootDir string) ([]string, error) {
	gitDirectories := make([]string, 0)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
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
