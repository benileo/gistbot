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
