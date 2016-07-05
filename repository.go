package main

import (
	"fmt"
	"log"
	"time"

	"github.com/libgit2/git2go"
)

type Repository struct {
	conf *Config
	repo *git.Repository
}

func NewRepository(conf *Config, path string) (*Repository, error) {
	repo, err := git.OpenRepository(path)
	if err != nil {
		return nil, fmt.Errorf("unable to create repository: %v", err)
	}

	return &Repository{conf: conf, repo: repo}, nil
}

func (r *Repository) Add() (*git.Tree, error) {
	index, err := r.repo.Index()
	if err != nil {
		return nil, fmt.Errorf("error getting index: %v", err)
	}

	if err = index.AddAll([]string{}, git.IndexAddDefault, nil); err != nil {
		return nil, fmt.Errorf("error adding all files to the index")
	}

	treeId, err := index.WriteTreeTo(r.repo)
	if err != nil {
		return nil, fmt.Errorf("error creating tree: %v", err)
	}

	if err = index.Write(); err != nil {
		return nil, fmt.Errorf("error writing index: %v", err)
	}

	tree, err := r.repo.LookupTree(treeId)
	if err != nil {
		return nil, fmt.Errorf("error looking up tree: %v", err)
	}

	return tree, nil
}

func (r *Repository) Commit(tree *git.Tree) error {
	var sig *git.Signature = &git.Signature{
		Name:  r.conf.Name,
		Email: r.conf.Email,
		When:  time.Now(),
	}
	var message string = "Committed by the Gist Bot"
	head, _ := r.head()

	commitTarget, err := r.repo.LookupCommit(head.Target())
	if err != nil {
		return fmt.Errorf("error looking up commit on local head: %v", err)
	}

	commitId, err := r.repo.CreateCommit("refs/heads/master", sig, sig, message, tree, commitTarget)
	if err != nil {
		return fmt.Errorf("error creating commit: %v", err)
	}

	log.Printf("commit: %s", commitId)
	return nil
}

func (r *Repository) Push() error {
	remote, err := r.repo.Remotes.Lookup("origin")
	if err != nil {
		return fmt.Errorf("error looking up remote origin: %v", err)
	}

	if err = remote.Push([]string{"refs/heads/master"}, r.pushOptions()); err != nil {
		return fmt.Errorf("error pushing to remote: %v", err)
	}

	return nil
}

func (r *Repository) Pull(ch chan error) {
	if err := r.fetch(); err != nil {
		ch <- err
		return
	}

	if err := r.merge(); err != nil {
		ch <- err
		return
	}

	ch <- nil
}

func (r *Repository) credentialsCallback(url string, username string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
	ret, cred := git.NewCredSshKey("git", r.conf.PublicKey, r.conf.PrivateKey, "")
	return git.ErrorCode(ret), &cred
}

func (r *Repository) certificateCheckCallback(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	return 0
}

func (r *Repository) fetchOptions() *git.FetchOptions {
	fo := git.FetchOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      r.credentialsCallback,
			CertificateCheckCallback: r.certificateCheckCallback,
		},
	}
	return &fo
}

func (r *Repository) pushOptions() *git.PushOptions {
	po := git.PushOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      r.credentialsCallback,
			CertificateCheckCallback: r.certificateCheckCallback,
		},
	}
	return &po
}

func (r *Repository) signature() *git.Signature {
	return &git.Signature{
		Name:  r.conf.Name,
		Email: r.conf.Email,
		When:  time.Now(),
	}
}

func (r *Repository) origin() (*git.Remote, error) {
	remote, err := r.repo.Remotes.Lookup("origin")
	if err != nil {
		fmt.Errorf("error looking up origin: %v", err)
		return nil, err
	}

	return remote, nil
}

func (r *Repository) masterRemote() (*git.Reference, error) {
	master, err := r.repo.References.Lookup("refs/remotes/origin/master") // remote master..
	if err != nil {
		fmt.Errorf("error looking up master branch: %v", err)
		return nil, err
	}

	return master, nil
}

func (r *Repository) fastForward(oid *git.Oid) error {
	head, _ := r.head()

	// Lookup the git tree object for the given ref
	remoteTree, err := r.repo.LookupTree(oid)
	if err != nil {
		return fmt.Errorf("error lookup tree HEAD: %v", err)
	}

	// Honestly, don't fully understand this one
	if err := r.repo.CheckoutTree(remoteTree, nil); err != nil {
		return err
	}

	// lookup local master branch
	branchRef, err := r.repo.References.Lookup("refs/heads/master")
	if err != nil {
		return fmt.Errorf("unable to lookup refs/heads/master: %v", err)
	}

	// Point branch to the object (don't fully understand this one too)
	branchRef.SetTarget(oid, "")
	if _, err := head.SetTarget(oid, ""); err != nil {
		return err
	}

	return nil
}

func (r *Repository) fetch() error {
	remote, err := r.origin()
	if err != nil {
		return err
	}

	if err = remote.Fetch([]string{}, r.fetchOptions(), ""); err != nil {
		return err
	}

	return nil
}

func (r *Repository) merge() error {
	masterRemote, err := r.masterRemote()
	if err != nil {
		return err
	}

	target := masterRemote.Target()
	annotatedCommit, _ := r.repo.AnnotatedCommitFromRef(masterRemote)
	mergeHeads := []*git.AnnotatedCommit{annotatedCommit}
	analysis, _, _ := r.repo.MergeAnalysis(mergeHeads)

	if analysis&git.MergeAnalysisUpToDate != 0 {
		log.Println("Everything up to date")

		return nil
	} else if analysis&git.MergeAnalysisFastForward != 0 {
		// Fast forward

		if err = r.fastForward(target); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) head() (*git.Reference, error) {
	head, err := r.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("error getting HEAD: %v", err)
	}

	return head, nil
}
