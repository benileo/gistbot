package main

import (
    "path/filepath"
    "os"
)

func findGists(rootDir string) ([]string, error) {
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
