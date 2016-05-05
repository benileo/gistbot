package main

import (
    "path/filepath"
    "os"
    "gopkg.in/libgit2/git2go.v23"
    "log"
)
// todo: find a good home for these functions.


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

func createRepos(gists []string) ([]*git.Repository, error) {
    repos := make([]*git.Repository, 0)

    for _, repo := range gists {
        repo, err := git.OpenRepository(repo)
        if err != nil {
            log.Printf("unable to open repo with path %s: %v", repo, err)
            continue
        }

        repos = append(repos, repo)
    }

    return repos, nil //todo: should we return an error if we are not able to open a repository or just log?
}
