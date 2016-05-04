package main

import (
	"gopkg.in/libgit2/git2go.v23"
)

type GitControl struct {
	username     string
	privatekey   string
	publickey    string
	cred         git.Cred
	fetchOptions git.FetchOptions
}

func NewGitControl(username, publickey, privatekey string) *GitControl {
	return &GitControl{
		username:   username,
		publickey:  privatekey,
		privatekey: privatekey,
	}
}

//func (gd *GistDaemon) pullGists() error {
//	for _, repo := range gd.gists {
//		repository, err := git.OpenRepository(repo)
//		if err != nil {
//			return fmt.Errorf("unable to open repo with path %s: %v", repo, err)
//		}
//		remote, err := repository.Remotes.Lookup("origin")
//		if err != nil {
//			return fmt.Errorf("unable to find remote origin: %v", err)
//		}
//		fmt.Println(remote.Name())
//		if err := remote.Fetch(make([]string, 0), nil, ""); err != nil {
//			return fmt.Errorf("error fetching repo: %v", err)
//		}
//	}
//	return nil
//}
